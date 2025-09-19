package leaderboard

import (
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// ScoreSubmission enforces idempotent leaderboard writes.
type ScoreSubmission struct {
	PlayerID       shared.PlayerID
	SeasonID       shared.SeasonID
	Value          int64
	IdempotencyKey shared.IdempotencyKey
	SubmittedAt    time.Time
}

// Season aggregates leaderboard policy.
type Season struct {
	ID       shared.SeasonID
	StartsAt time.Time
	EndsAt   time.Time
	Active   bool
}

func (s *Season) Activate(now time.Time) {
	s.Active = now.After(s.StartsAt) && now.Before(s.EndsAt)
}

func (submission ScoreSubmission) Validate() error {
	if err := submission.PlayerID.Validate(); err != nil {
		return err
	}
	if err := submission.SeasonID.Validate(); err != nil {
		return err
	}
	if err := submission.IdempotencyKey.Validate(); err != nil {
		return err
	}
	return nil
}
