package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

type tournamentCreatePayload struct {
	Authoritative bool   `json:"authoritative"`
	SortOrder     string `json:"sort_order"`
	Operator      string `json:"operator"`
	ResetSchedule string `json:"reset_schedule"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Category      int    `json:"category"`
	StartTime     int    `json:"start_time"`
	EndTime       int    `json:"end_time"`
	Duration      int    `json:"duration"`
	MaxSize       int    `json:"max_size"`
	MaxNumScore   int    `json:"max_num_score"`
	JoinRequired  bool   `json:"join_required"`
}

type tournamentIDPayload struct {
	TournamentID string `json:"tournament_id"`
}

type tournamentAddAttemptPayload struct {
	TournamentID string `json:"tournament_id"`
	OwnerID      string `json:"owner_id"`
	Count        int    `json:"count"`
}

func registerTournamentRuntime(initializer runtime.Initializer) error {
	if err := initializer.RegisterTournamentEnd(tournamentEndCallback); err != nil {
		return err
	}
	if err := initializer.RegisterTournamentReset(tournamentResetCallback); err != nil {
		return err
	}
	if err := initializer.RegisterLeaderboardReset(leaderboardResetCallback); err != nil {
		return err
	}

	if err := initializer.RegisterRpc("clientrpc.create_same_tournament_multiple_times", rpcCreateSameTournamentMultipleTimes); err != nil {
		return err
	}
	if err := initializer.RegisterRpc("clientrpc.create_tournament", rpcCreateTournament); err != nil {
		return err
	}
	if err := initializer.RegisterRpc("clientrpc.delete_tournament", rpcDeleteTournament); err != nil {
		return err
	}
	if err := initializer.RegisterRpc("clientrpc.addattempt_tournament", rpcAddAttemptTournament); err != nil {
		return err
	}

	return nil
}

func tournamentEndCallback(ctx context.Context, _ runtime.Logger, _ *sql.DB, nk runtime.NakamaModule, tournament *api.Tournament, _ int64, reset int64) error {
	records, _, _, _, err := nk.LeaderboardRecordsList(ctx, tournament.GetId(), nil, 1, "", reset)
	if err != nil {
		return fmt.Errorf("fetching tournament records: %w", err)
	}

	if len(records) == 0 {
		return nil
	}

	metadata := map[string]interface{}{"won": tournament.GetId()}
	if err := nk.AccountUpdateId(ctx, records[0].GetOwnerId(), "", metadata, "", "", "", "", ""); err != nil {
		return fmt.Errorf("updating winner account metadata: %w", err)
	}

	return nil
}

func tournamentResetCallback(ctx context.Context, _ runtime.Logger, _ *sql.DB, nk runtime.NakamaModule, tournament *api.Tournament, _ int64, _ int64) error {
	records, _, _, _, err := nk.LeaderboardRecordsList(ctx, tournament.GetId(), nil, 1, "", 0)
	if err != nil {
		return fmt.Errorf("fetching tournament records: %w", err)
	}

	if len(records) == 0 {
		return nil
	}

	metadata := map[string]interface{}{"expiry_tournament": tournament.GetId()}
	if err := nk.AccountUpdateId(ctx, records[0].GetOwnerId(), "", metadata, "", "", "", "", ""); err != nil {
		return fmt.Errorf("updating account metadata on tournament reset: %w", err)
	}

	return nil
}

func leaderboardResetCallback(ctx context.Context, _ runtime.Logger, _ *sql.DB, nk runtime.NakamaModule, leaderboard *api.Leaderboard, reset int64) error {
	records, _, _, _, err := nk.LeaderboardRecordsList(ctx, leaderboard.GetId(), nil, 1, "", reset)
	if err != nil {
		return fmt.Errorf("fetching leaderboard records: %w", err)
	}

	if len(records) == 0 {
		return nil
	}

	metadata := map[string]interface{}{"expiry_leaderboard": leaderboard.GetId()}
	if err := nk.AccountUpdateId(ctx, records[0].GetOwnerId(), "", metadata, "", "", "", "", ""); err != nil {
		return fmt.Errorf("updating account metadata on leaderboard reset: %w", err)
	}

	return nil
}

func rpcCreateSameTournamentMultipleTimes(ctx context.Context, _ runtime.Logger, _ *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	args, err := decodeTournamentCreatePayload(payload)
	if err != nil {
		return "", err
	}

	id, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("generating tournament id: %w", err)
	}
	idStr := id.String()

	if err := createTournament(ctx, nk, idStr, args); err != nil {
		return "", err
	}
	if err := createTournament(ctx, nk, idStr, args); err != nil {
		return "", err
	}

	return encodeTournamentIDResponse(idStr)
}

func rpcCreateTournament(ctx context.Context, _ runtime.Logger, _ *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	args, err := decodeTournamentCreatePayload(payload)
	if err != nil {
		return "", err
	}

	id, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("generating tournament id: %w", err)
	}
	idStr := id.String()

	if err := createTournament(ctx, nk, idStr, args); err != nil {
		return "", err
	}

	return encodeTournamentIDResponse(idStr)
}

func rpcDeleteTournament(ctx context.Context, _ runtime.Logger, _ *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var args tournamentIDPayload
	if err := json.Unmarshal([]byte(payload), &args); err != nil {
		return "", runtime.NewError("invalid payload", 3)
	}
	if args.TournamentID == "" {
		return "", runtime.NewError("tournament_id is required", 3)
	}

	if err := nk.TournamentDelete(ctx, args.TournamentID); err != nil {
		return "", fmt.Errorf("deleting tournament: %w", err)
	}

	return "{}", nil
}

func rpcAddAttemptTournament(ctx context.Context, _ runtime.Logger, _ *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var args tournamentAddAttemptPayload
	if err := json.Unmarshal([]byte(payload), &args); err != nil {
		return "", runtime.NewError("invalid payload", 3)
	}
	if args.TournamentID == "" || args.OwnerID == "" {
		return "", runtime.NewError("tournament_id and owner_id are required", 3)
	}
	if args.Count == 0 {
		return "", runtime.NewError("count must be non-zero", 3)
	}

	if err := nk.TournamentAddAttempt(ctx, args.TournamentID, args.OwnerID, args.Count); err != nil {
		return "", fmt.Errorf("adding tournament attempt: %w", err)
	}

	return "{}", nil
}

func createTournament(ctx context.Context, nk runtime.NakamaModule, id string, args *tournamentCreatePayload) error {
	if err := nk.TournamentCreate(ctx, id, args.Authoritative, args.SortOrder, args.Operator, args.ResetSchedule, nil, args.Title, args.Description, args.Category, args.StartTime, args.EndTime, args.Duration, args.MaxSize, args.MaxNumScore, args.JoinRequired, false); err != nil {
		return fmt.Errorf("creating tournament %s: %w", id, err)
	}
	return nil
}

func decodeTournamentCreatePayload(payload string) (*tournamentCreatePayload, error) {
	var args tournamentCreatePayload
	if err := json.Unmarshal([]byte(payload), &args); err != nil {
		return nil, runtime.NewError("invalid payload", 3)
	}

	return &args, nil
}

func encodeTournamentIDResponse(id string) (string, error) {
	response := map[string]string{"tournament_id": id}
	out, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("encoding response: %w", err)
	}
	return string(out), nil
}
