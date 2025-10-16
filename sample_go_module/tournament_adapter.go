package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/src/app/tournaments"
	"github.com/heroiclabs/nakama/v3/src/domain/shared"
	"github.com/heroiclabs/nakama/v3/src/domain/tournament"
	infraTournament "github.com/heroiclabs/nakama/v3/src/infra/tournament"
)

// TournamentServiceAdapter adapts the DDD service to Nakama RPC handlers.
type TournamentServiceAdapter struct {
	service *tournaments.Service
}

// NewTournamentServiceAdapter creates a new adapter with DDD service.
func NewTournamentServiceAdapter(nk runtime.NakamaModule) *TournamentServiceAdapter {
	repo := infraTournament.NewMemoryRepository()
	participantRepo := infraTournament.NewMemoryParticipantRepository()
	provider := infraTournament.NewNakamaProvider(nk)

	service := tournaments.NewService(repo, participantRepo, provider)

	return &TournamentServiceAdapter{
		service: service,
	}
}

// CreateTournament creates a tournament using the DDD service.
func (a *TournamentServiceAdapter) CreateTournament(ctx context.Context, payload tournamentCreatePayload) (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("generating tournament id: %w", err)
	}

	cmd := tournaments.CreateTournamentCommand{
		ID:            shared.TournamentID(id.String()),
		Title:         payload.Title,
		Description:   payload.Description,
		Category:      payload.Category,
		SortOrder:     tournament.SortOrder(payload.SortOrder),
		Operator:      tournament.Operator(payload.Operator),
		ResetSchedule: payload.ResetSchedule,
		Authoritative: payload.Authoritative,
		JoinRequired:  payload.JoinRequired,
		MaxSize:       payload.MaxSize,
		MaxNumScore:   payload.MaxNumScore,
		StartTime:     time.Unix(int64(payload.StartTime), 0),
		Duration:      time.Duration(payload.Duration) * time.Second,
	}

	if payload.EndTime > 0 {
		endTime := time.Unix(int64(payload.EndTime), 0)
		cmd.EndTime = &endTime
	}

	result, err := a.service.CreateTournament(ctx, cmd)
	if err != nil {
		return "", err
	}

	response := map[string]string{"tournament_id": string(result.TournamentID)}
	out, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("encoding response: %w", err)
	}

	return string(out), nil
}

// DeleteTournament deletes a tournament using the DDD service.
func (a *TournamentServiceAdapter) DeleteTournament(ctx context.Context, tournamentID string) error {
	cmd := tournaments.DeleteTournamentCommand{
		TournamentID: shared.TournamentID(tournamentID),
	}
	return a.service.DeleteTournament(ctx, cmd)
}

// AddAttempt adds attempts using the DDD service.
func (a *TournamentServiceAdapter) AddAttempt(ctx context.Context, tournamentID, ownerID string, count int) error {
	cmd := tournaments.AddAttemptCommand{
		TournamentID: shared.TournamentID(tournamentID),
		PlayerID:     shared.PlayerID(ownerID),
		Count:        count,
	}
	return a.service.AddAttempt(ctx, cmd)
}

// RPC handler functions using the adapter
func rpcCreateTournamentWithAdapter(ctx context.Context, _ runtime.Logger, _ *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	args, err := decodeTournamentCreatePayload(payload)
	if err != nil {
		return "", err
	}

	adapter := NewTournamentServiceAdapter(nk)
	return adapter.CreateTournament(ctx, *args)
}

func rpcDeleteTournamentWithAdapter(ctx context.Context, _ runtime.Logger, _ *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var args tournamentIDPayload
	if err := json.Unmarshal([]byte(payload), &args); err != nil {
		return "", runtime.NewError("invalid payload", 3)
	}
	if args.TournamentID == "" {
		return "", runtime.NewError("tournament_id is required", 3)
	}

	adapter := NewTournamentServiceAdapter(nk)
	if err := adapter.DeleteTournament(ctx, args.TournamentID); err != nil {
		return "", fmt.Errorf("deleting tournament: %w", err)
	}

	return "{}", nil
}

func rpcAddAttemptTournamentWithAdapter(ctx context.Context, _ runtime.Logger, _ *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
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

	adapter := NewTournamentServiceAdapter(nk)
	if err := adapter.AddAttempt(ctx, args.TournamentID, args.OwnerID, args.Count); err != nil {
		return "", fmt.Errorf("adding tournament attempt: %w", err)
	}

	return "{}", nil
}
