package tournament

import (
	"context"

	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// Repository manages tournament persistence.
type Repository interface {
	Save(ctx context.Context, tournament *Tournament) error
	Get(ctx context.Context, id shared.TournamentID) (*Tournament, error)
	Delete(ctx context.Context, id shared.TournamentID) error
	List(ctx context.Context, limit, offset int) ([]*Tournament, error)
}

// ParticipantRepository manages participant persistence.
type ParticipantRepository interface {
	Save(ctx context.Context, participant *Participant) error
	Get(ctx context.Context, tournamentID shared.TournamentID, playerID shared.PlayerID) (*Participant, error)
	ListByTournament(ctx context.Context, tournamentID shared.TournamentID) ([]*Participant, error)
	Delete(ctx context.Context, tournamentID shared.TournamentID, playerID shared.PlayerID) error
}
