package tournament_test

import (
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/shared"
	"github.com/heroiclabs/nakama/v3/src/domain/tournament"
)

func TestNewParticipant(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		tournamentID shared.TournamentID
		playerID     shared.PlayerID
		joinedAt     time.Time
		wantErr      bool
	}{
		{
			name:         "valid participant",
			tournamentID: "tournament-123",
			playerID:     "player-456",
			joinedAt:     now,
			wantErr:      false,
		},
		{
			name:         "empty tournament id",
			tournamentID: "",
			playerID:     "player-456",
			joinedAt:     now,
			wantErr:      true,
		},
		{
			name:         "empty player id",
			tournamentID: "tournament-123",
			playerID:     "",
			joinedAt:     now,
			wantErr:      true,
		},
		{
			name:         "zero joined time",
			tournamentID: "tournament-123",
			playerID:     "player-456",
			joinedAt:     time.Time{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			participant, err := tournament.NewParticipant(tt.tournamentID, tt.playerID, tt.joinedAt)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewParticipant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if participant.Attempts != 0 {
					t.Errorf("Expected attempts 0, got %v", participant.Attempts)
				}
			}
		})
	}
}

func TestParticipant_AddAttempts(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		count   int
		wantErr bool
	}{
		{
			name:    "add positive attempts",
			count:   5,
			wantErr: false,
		},
		{
			name:    "add zero attempts",
			count:   0,
			wantErr: true,
		},
		{
			name:    "add negative attempts",
			count:   -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := tournament.NewParticipant("tournament-123", "player-456", now)
			initialAttempts := p.Attempts

			err := p.AddAttempts(tt.count, now)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddAttempts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				expected := initialAttempts + tt.count
				if p.Attempts != expected {
					t.Errorf("Expected attempts %v, got %v", expected, p.Attempts)
				}
			}
		})
	}
}

func TestParticipant_ResetAttempts(t *testing.T) {
	now := time.Now()
	participant, _ := tournament.NewParticipant("tournament-123", "player-456", now)

	participant.AddAttempts(10, now)
	if participant.Attempts != 10 {
		t.Fatalf("Expected 10 attempts after adding, got %v", participant.Attempts)
	}

	participant.ResetAttempts(now)
	if participant.Attempts != 0 {
		t.Errorf("Expected 0 attempts after reset, got %v", participant.Attempts)
	}
}
