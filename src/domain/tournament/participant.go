package tournament

import (
	"errors"
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// Participant represents a player in a tournament.
type Participant struct {
	TournamentID shared.TournamentID
	PlayerID     shared.PlayerID
	Attempts     int
	JoinedAt     time.Time
	UpdatedAt    time.Time
}

// NewParticipant creates a new tournament participant.
func NewParticipant(tournamentID shared.TournamentID, playerID shared.PlayerID, joinedAt time.Time) (*Participant, error) {
	if err := tournamentID.Validate(); err != nil {
		return nil, err
	}
	if err := playerID.Validate(); err != nil {
		return nil, err
	}
	if joinedAt.IsZero() {
		return nil, errors.New("joined time is required")
	}
	return &Participant{
		TournamentID: tournamentID,
		PlayerID:     playerID,
		Attempts:     0,
		JoinedAt:     joinedAt,
		UpdatedAt:    joinedAt,
	}, nil
}

// AddAttempts increments the attempt count.
func (p *Participant) AddAttempts(count int, now time.Time) error {
	if count <= 0 {
		return errors.New("attempt count must be positive")
	}
	p.Attempts += count
	p.UpdatedAt = now
	return nil
}

// ResetAttempts sets attempts back to zero.
func (p *Participant) ResetAttempts(now time.Time) {
	p.Attempts = 0
	p.UpdatedAt = now
}
