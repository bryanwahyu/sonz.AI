package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

// InitModule is the entrypoint for the Sand-ai Nakama runtime extension.
func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	if err := initializer.RegisterBeforeAuthenticateDevice(beforeAuthenticateDevice); err != nil {
		return err
	}
	if err := initializer.RegisterAfterAuthenticateDevice(afterAuthenticateDevice); err != nil {
		return err
	}
	if err := initializer.RegisterBeforeWriteLeaderboardRecord(beforeWriteLeaderboardRecord); err != nil {
		return err
	}
	if err := initializer.RegisterMatch("sandai_battle", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		return &battleMatch{}, nil
	}); err != nil {
		return err
	}
	logger.Info("Sand-ai runtime module registered")
	return nil
}

func beforeAuthenticateDevice(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *api.AuthenticateDeviceRequest) (*api.AuthenticateDeviceRequest, error) {
	if in.Account == nil || in.Account.Id == "" {
		return nil, runtime.BadInputError("device id required", nil)
	}
	return in, nil
}

func afterAuthenticateDevice(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out *api.Session, in *api.AuthenticateDeviceRequest) error {
	logger.Info("device login", "device", in.GetAccount().GetId(), "token", out.GetToken())
	return nil
}

func beforeWriteLeaderboardRecord(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *api.WriteLeaderboardRecordRequest) (*api.WriteLeaderboardRecordRequest, error) {
	if in.Record == nil {
		return nil, runtime.BadInputError("missing leaderboard record", nil)
	}
	if in.Record.Metadata == "" {
		metadata := map[string]any{"validated_at": time.Now().UTC()}
		payload, _ := json.Marshal(metadata)
		in.Record.Metadata = string(payload)
	}
	return in, nil
}

type battleMatch struct{}

type matchState struct {
	Tick    int64                       `json:"tick"`
	Players map[string]runtime.Presence `json:"-"`
}

func (m *battleMatch) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]any) (interface{}, int, string) {
	state := &matchState{Tick: 0, Players: make(map[string]runtime.Presence)}
	return state, 10, "sandai"
}

func (m *battleMatch) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, st interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	return st, true, ""
}

func (m *battleMatch) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, st interface{}, presences []runtime.Presence) interface{} {
	state := st.(*matchState)
	for _, p := range presences {
		state.Players[p.GetSessionId()] = p
	}
	return state
}

func (m *battleMatch) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, st interface{}, presences []runtime.Presence) interface{} {
	state := st.(*matchState)
	for _, p := range presences {
		delete(state.Players, p.GetSessionId())
	}
	return state
}

func (m *battleMatch) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, st interface{}, messages []runtime.MatchData) interface{} {
	state := st.(*matchState)
	state.Tick = tick
	if len(messages) > 0 {
		for _, msg := range messages {
			dispatcher.BroadcastMessage(1, msg.GetData(), nil, nil, true)
		}
	}
	if len(state.Players) == 0 && tick > 30 {
		return nil
	}
	return state
}

func (m *battleMatch) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, st interface{}, graceSeconds int) interface{} {
	return st
}

func (m *battleMatch) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, st interface{}, data string) (interface{}, string) {
	return st, data
}
