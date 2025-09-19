package battles

import (
	"context"
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/battle"
	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// MatchProvider abstracts Nakama matchmaker or authoritative match calls.
type MatchProvider interface {
	CreateMatch(ctx context.Context, payload StartBattlePayload) (StartBattleResult, error)
}

type Repository interface {
	battle.Repository
}

// StartBattlePayload describes the Nakama RPC fields.
type StartBattlePayload struct {
	LeaderID shared.PlayerID
	Metadata map[string]any
	Preset   string
}

// StartBattleResult contains the match ID returned by Nakama.
type StartBattleResult struct {
	BattleID shared.BattleID
	MatchID  string
}

// Service coordinates battle creation.
type Service struct {
	Repo     Repository
	Provider MatchProvider
	Clock    func() time.Time
}

func NewService(repo Repository, provider MatchProvider) *Service {
	return &Service{
		Repo:     repo,
		Provider: provider,
		Clock:    func() time.Time { return time.Now().UTC() },
	}
}

type StartCommand struct {
	LeaderID       shared.PlayerID
	IdempotencyKey shared.IdempotencyKey
	Metadata       map[string]any
	Preset         string
}

type StartResult struct {
	BattleID shared.BattleID
	MatchID  string
}

func (s *Service) StartBattle(ctx context.Context, cmd StartCommand) (StartResult, error) {
	if err := cmd.LeaderID.Validate(); err != nil {
		return StartResult{}, err
	}
	if err := cmd.IdempotencyKey.Validate(); err != nil {
		return StartResult{}, err
	}
	payload := StartBattlePayload{
		LeaderID: cmd.LeaderID,
		Metadata: cmd.Metadata,
		Preset:   cmd.Preset,
	}
	result, err := s.Provider.CreateMatch(ctx, payload)
	if err != nil {
		return StartResult{}, err
	}
	now := s.Clock()
	aggregate, err := battle.NewBattle(result.BattleID, cmd.LeaderID, cmd.IdempotencyKey, now)
	if err != nil {
		return StartResult{}, err
	}
	if err := s.Repo.Save(ctx, aggregate); err != nil {
		return StartResult{}, err
	}
	return StartResult{BattleID: result.BattleID, MatchID: result.MatchID}, nil
}
