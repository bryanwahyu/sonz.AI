package leaderboard

import "context"

import "github.com/heroiclabs/nakama/v3/src/domain/shared"

type Repository interface {
	SubmitScore(ctx context.Context, submission ScoreSubmission) error
	GetSeason(ctx context.Context, id shared.SeasonID) (*Season, error)
}
