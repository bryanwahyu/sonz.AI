package tournament

import (
	"context"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/src/app/tournaments"
	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// NakamaProviderImpl implements NakamaProvider using Nakama runtime.
type NakamaProviderImpl struct {
	nk runtime.NakamaModule
}

// NewNakamaProvider creates a new Nakama provider.
func NewNakamaProvider(nk runtime.NakamaModule) *NakamaProviderImpl {
	return &NakamaProviderImpl{nk: nk}
}

// CreateTournament creates a tournament in Nakama.
func (p *NakamaProviderImpl) CreateTournament(ctx context.Context, params tournaments.CreateTournamentParams) error {
	return p.nk.TournamentCreate(
		ctx,
		params.ID,
		params.Authoritative,
		params.SortOrder,
		params.Operator,
		params.ResetSchedule,
		nil, // metadata
		params.Title,
		params.Description,
		params.Category,
		params.StartTime,
		params.EndTime,
		params.Duration,
		params.MaxSize,
		params.MaxNumScore,
		params.JoinRequired,
		false, // enableRanks
	)
}

// DeleteTournament deletes a tournament from Nakama.
func (p *NakamaProviderImpl) DeleteTournament(ctx context.Context, id shared.TournamentID) error {
	return p.nk.TournamentDelete(ctx, string(id))
}

// AddAttempt adds attempts for a player in a tournament.
func (p *NakamaProviderImpl) AddAttempt(ctx context.Context, tournamentID shared.TournamentID, playerID shared.PlayerID, count int) error {
	return p.nk.TournamentAddAttempt(ctx, string(tournamentID), string(playerID), count)
}
