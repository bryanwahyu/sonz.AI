package leaderboard

import (
	"context"
	"time"

	domain "github.com/heroiclabs/nakama/v3/src/domain/leaderboard"
	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

type Repository interface {
	domain.Repository
}

// Service coordinates leaderboard submissions.
type Service struct {
	Repo  Repository
	Clock func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		Repo:  repo,
		Clock: func() time.Time { return time.Now().UTC() },
	}
}

type SubmitCommand struct {
	PlayerID       shared.PlayerID
	SeasonID       shared.SeasonID
	Score          int64
	IdempotencyKey shared.IdempotencyKey
}

type SubmitResult struct {
	Acknowledged bool
}

func (s *Service) Submit(ctx context.Context, cmd SubmitCommand) (SubmitResult, error) {
	submission := domain.ScoreSubmission{
		PlayerID:       cmd.PlayerID,
		SeasonID:       cmd.SeasonID,
		Value:          cmd.Score,
		IdempotencyKey: cmd.IdempotencyKey,
		SubmittedAt:    s.Clock(),
	}
	if err := submission.Validate(); err != nil {
		return SubmitResult{}, err
	}
	if err := s.Repo.SubmitScore(ctx, submission); err != nil {
		return SubmitResult{}, err
	}
	return SubmitResult{Acknowledged: true}, nil
}
