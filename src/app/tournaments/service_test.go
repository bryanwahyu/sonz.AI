package tournaments_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/src/app/tournaments"
	"github.com/heroiclabs/nakama/v3/src/domain/shared"
	"github.com/heroiclabs/nakama/v3/src/domain/tournament"
)

// Mock implementations
type mockTournamentRepo struct {
	saveFunc   func(ctx context.Context, t *tournament.Tournament) error
	getFunc    func(ctx context.Context, id shared.TournamentID) (*tournament.Tournament, error)
	deleteFunc func(ctx context.Context, id shared.TournamentID) error
	listFunc   func(ctx context.Context, limit, offset int) ([]*tournament.Tournament, error)
}

func (m *mockTournamentRepo) Save(ctx context.Context, t *tournament.Tournament) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, t)
	}
	return nil
}

func (m *mockTournamentRepo) Get(ctx context.Context, id shared.TournamentID) (*tournament.Tournament, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return nil, tournament.ErrTournamentNotFound
}

func (m *mockTournamentRepo) Delete(ctx context.Context, id shared.TournamentID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockTournamentRepo) List(ctx context.Context, limit, offset int) ([]*tournament.Tournament, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset)
	}
	return []*tournament.Tournament{}, nil
}

type mockParticipantRepo struct {
	saveFunc           func(ctx context.Context, p *tournament.Participant) error
	getFunc            func(ctx context.Context, tournamentID shared.TournamentID, playerID shared.PlayerID) (*tournament.Participant, error)
	listByTournamentFunc func(ctx context.Context, tournamentID shared.TournamentID) ([]*tournament.Participant, error)
	deleteFunc         func(ctx context.Context, tournamentID shared.TournamentID, playerID shared.PlayerID) error
}

func (m *mockParticipantRepo) Save(ctx context.Context, p *tournament.Participant) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, p)
	}
	return nil
}

func (m *mockParticipantRepo) Get(ctx context.Context, tournamentID shared.TournamentID, playerID shared.PlayerID) (*tournament.Participant, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, tournamentID, playerID)
	}
	return nil, tournament.ErrParticipantNotFound
}

func (m *mockParticipantRepo) ListByTournament(ctx context.Context, tournamentID shared.TournamentID) ([]*tournament.Participant, error) {
	if m.listByTournamentFunc != nil {
		return m.listByTournamentFunc(ctx, tournamentID)
	}
	return []*tournament.Participant{}, nil
}

func (m *mockParticipantRepo) Delete(ctx context.Context, tournamentID shared.TournamentID, playerID shared.PlayerID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, tournamentID, playerID)
	}
	return nil
}

type mockNakamaProvider struct {
	createFunc     func(ctx context.Context, params tournaments.CreateTournamentParams) error
	deleteFunc     func(ctx context.Context, id shared.TournamentID) error
	addAttemptFunc func(ctx context.Context, tournamentID shared.TournamentID, playerID shared.PlayerID, count int) error
}

func (m *mockNakamaProvider) CreateTournament(ctx context.Context, params tournaments.CreateTournamentParams) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, params)
	}
	return nil
}

func (m *mockNakamaProvider) DeleteTournament(ctx context.Context, id shared.TournamentID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockNakamaProvider) AddAttempt(ctx context.Context, tournamentID shared.TournamentID, playerID shared.PlayerID, count int) error {
	if m.addAttemptFunc != nil {
		return m.addAttemptFunc(ctx, tournamentID, playerID, count)
	}
	return nil
}

func TestService_CreateTournament(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name        string
		cmd         tournaments.CreateTournamentCommand
		saveErr     error
		providerErr error
		wantErr     bool
	}{
		{
			name: "successful tournament creation",
			cmd: tournaments.CreateTournamentCommand{
				ID:            "tournament-123",
				Title:         "Test Tournament",
				Description:   "Test Description",
				Category:      1,
				SortOrder:     tournament.SortOrderDescending,
				Operator:      tournament.OperatorBest,
				ResetSchedule: "",
				Authoritative: true,
				JoinRequired:  false,
				MaxSize:       100,
				MaxNumScore:   10,
				StartTime:     now.Add(1 * time.Hour),
				Duration:      24 * time.Hour,
			},
			wantErr: false,
		},
		{
			name: "empty tournament id",
			cmd: tournaments.CreateTournamentCommand{
				ID:        "",
				Title:     "Test Tournament",
				StartTime: now.Add(1 * time.Hour),
				Duration:  24 * time.Hour,
			},
			wantErr: true,
		},
		{
			name: "save failure",
			cmd: tournaments.CreateTournamentCommand{
				ID:        "tournament-123",
				Title:     "Test Tournament",
				Category:  1,
				StartTime: now.Add(1 * time.Hour),
				Duration:  24 * time.Hour,
			},
			saveErr: errors.New("save failed"),
			wantErr: true,
		},
		{
			name: "provider failure",
			cmd: tournaments.CreateTournamentCommand{
				ID:        "tournament-123",
				Title:     "Test Tournament",
				Category:  1,
				StartTime: now.Add(1 * time.Hour),
				Duration:  24 * time.Hour,
			},
			providerErr: errors.New("provider failed"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTournamentRepo{
				saveFunc: func(ctx context.Context, tour *tournament.Tournament) error {
					return tt.saveErr
				},
			}

			participantRepo := &mockParticipantRepo{}

			provider := &mockNakamaProvider{
				createFunc: func(ctx context.Context, params tournaments.CreateTournamentParams) error {
					return tt.providerErr
				},
			}

			service := tournaments.NewService(repo, participantRepo, provider)
			result, err := service.CreateTournament(ctx, tt.cmd)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTournament() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result.TournamentID != tt.cmd.ID {
				t.Errorf("Expected tournament ID %v, got %v", tt.cmd.ID, result.TournamentID)
			}
		})
	}
}

func TestService_DeleteTournament(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		cmd         tournaments.DeleteTournamentCommand
		deleteErr   error
		providerErr error
		wantErr     bool
	}{
		{
			name: "successful deletion",
			cmd: tournaments.DeleteTournamentCommand{
				TournamentID: "tournament-123",
			},
			wantErr: false,
		},
		{
			name: "empty tournament id",
			cmd: tournaments.DeleteTournamentCommand{
				TournamentID: "",
			},
			wantErr: true,
		},
		{
			name: "provider failure",
			cmd: tournaments.DeleteTournamentCommand{
				TournamentID: "tournament-123",
			},
			providerErr: errors.New("provider failed"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTournamentRepo{
				deleteFunc: func(ctx context.Context, id shared.TournamentID) error {
					return tt.deleteErr
				},
			}

			participantRepo := &mockParticipantRepo{}

			provider := &mockNakamaProvider{
				deleteFunc: func(ctx context.Context, id shared.TournamentID) error {
					return tt.providerErr
				},
			}

			service := tournaments.NewService(repo, participantRepo, provider)
			err := service.DeleteTournament(ctx, tt.cmd)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteTournament() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_AddAttempt(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name               string
		cmd                tournaments.AddAttemptCommand
		existingParticipant *tournament.Participant
		getErr             error
		providerErr        error
		wantErr            bool
	}{
		{
			name: "successful attempt addition for existing participant",
			cmd: tournaments.AddAttemptCommand{
				TournamentID: "tournament-123",
				PlayerID:     "player-456",
				Count:        5,
			},
			existingParticipant: &tournament.Participant{
				TournamentID: "tournament-123",
				PlayerID:     "player-456",
				Attempts:     10,
				JoinedAt:     now,
			},
			wantErr: false,
		},
		{
			name: "successful attempt addition for new participant",
			cmd: tournaments.AddAttemptCommand{
				TournamentID: "tournament-123",
				PlayerID:     "player-456",
				Count:        5,
			},
			getErr:  tournament.ErrParticipantNotFound,
			wantErr: false,
		},
		{
			name: "zero count",
			cmd: tournaments.AddAttemptCommand{
				TournamentID: "tournament-123",
				PlayerID:     "player-456",
				Count:        0,
			},
			wantErr: true,
		},
		{
			name: "negative count",
			cmd: tournaments.AddAttemptCommand{
				TournamentID: "tournament-123",
				PlayerID:     "player-456",
				Count:        -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTournamentRepo{}

			participantRepo := &mockParticipantRepo{
				getFunc: func(ctx context.Context, tournamentID shared.TournamentID, playerID shared.PlayerID) (*tournament.Participant, error) {
					if tt.getErr != nil {
						return nil, tt.getErr
					}
					return tt.existingParticipant, nil
				},
				saveFunc: func(ctx context.Context, p *tournament.Participant) error {
					return nil
				},
			}

			provider := &mockNakamaProvider{
				addAttemptFunc: func(ctx context.Context, tournamentID shared.TournamentID, playerID shared.PlayerID, count int) error {
					return tt.providerErr
				},
			}

			service := tournaments.NewService(repo, participantRepo, provider)
			err := service.AddAttempt(ctx, tt.cmd)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddAttempt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
