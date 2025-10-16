package tournaments

import (
	"context"
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/shared"
	"github.com/heroiclabs/nakama/v3/src/domain/tournament"
)

// NakamaProvider abstracts Nakama tournament operations.
type NakamaProvider interface {
	CreateTournament(ctx context.Context, params CreateTournamentParams) error
	DeleteTournament(ctx context.Context, id shared.TournamentID) error
	AddAttempt(ctx context.Context, tournamentID shared.TournamentID, playerID shared.PlayerID, count int) error
}

// CreateTournamentParams encapsulates Nakama tournament creation parameters.
type CreateTournamentParams struct {
	ID            string
	Authoritative bool
	SortOrder     string
	Operator      string
	ResetSchedule string
	Title         string
	Description   string
	Category      int
	StartTime     int
	EndTime       int
	Duration      int
	MaxSize       int
	MaxNumScore   int
	JoinRequired  bool
}

// Service coordinates tournament operations.
type Service struct {
	Repo         tournament.Repository
	Participants tournament.ParticipantRepository
	Provider     NakamaProvider
	Clock        func() time.Time
}

// NewService creates a new tournament service.
func NewService(repo tournament.Repository, participants tournament.ParticipantRepository, provider NakamaProvider) *Service {
	return &Service{
		Repo:         repo,
		Participants: participants,
		Provider:     provider,
		Clock:        func() time.Time { return time.Now().UTC() },
	}
}

// CreateTournamentCommand contains parameters for creating a tournament.
type CreateTournamentCommand struct {
	ID            shared.TournamentID
	Title         string
	Description   string
	Category      int
	SortOrder     tournament.SortOrder
	Operator      tournament.Operator
	ResetSchedule string
	Authoritative bool
	JoinRequired  bool
	MaxSize       int
	MaxNumScore   int
	StartTime     time.Time
	EndTime       *time.Time
	Duration      time.Duration
}

// CreateTournamentResult contains the created tournament ID.
type CreateTournamentResult struct {
	TournamentID shared.TournamentID
}

// CreateTournament creates a new tournament.
func (s *Service) CreateTournament(ctx context.Context, cmd CreateTournamentCommand) (CreateTournamentResult, error) {
	now := s.Clock()
	
	// Create domain aggregate
	t, err := tournament.NewTournament(
		cmd.ID,
		cmd.Title,
		cmd.Description,
		cmd.Category,
		cmd.SortOrder,
		cmd.Operator,
		cmd.ResetSchedule,
		cmd.Authoritative,
		cmd.JoinRequired,
		cmd.MaxSize,
		cmd.MaxNumScore,
		cmd.StartTime,
		cmd.Duration,
		now,
	)
	if err != nil {
		return CreateTournamentResult{}, err
	}

	// Save to repository
	if err := s.Repo.Save(ctx, t); err != nil {
		return CreateTournamentResult{}, err
	}

	// Create in Nakama
	params := CreateTournamentParams{
		ID:            string(cmd.ID),
		Authoritative: cmd.Authoritative,
		SortOrder:     string(cmd.SortOrder),
		Operator:      string(cmd.Operator),
		ResetSchedule: cmd.ResetSchedule,
		Title:         cmd.Title,
		Description:   cmd.Description,
		Category:      cmd.Category,
		StartTime:     int(cmd.StartTime.Unix()),
		Duration:      int(cmd.Duration.Seconds()),
		MaxSize:       cmd.MaxSize,
		MaxNumScore:   cmd.MaxNumScore,
		JoinRequired:  cmd.JoinRequired,
	}
	if cmd.EndTime != nil {
		params.EndTime = int(cmd.EndTime.Unix())
	}

	if err := s.Provider.CreateTournament(ctx, params); err != nil {
		return CreateTournamentResult{}, err
	}

	return CreateTournamentResult{TournamentID: t.ID}, nil
}

// DeleteTournamentCommand contains parameters for deleting a tournament.
type DeleteTournamentCommand struct {
	TournamentID shared.TournamentID
}

// DeleteTournament removes a tournament.
func (s *Service) DeleteTournament(ctx context.Context, cmd DeleteTournamentCommand) error {
	if err := cmd.TournamentID.Validate(); err != nil {
		return err
	}

	// Delete from Nakama
	if err := s.Provider.DeleteTournament(ctx, cmd.TournamentID); err != nil {
		return err
	}

	// Delete from repository
	if err := s.Repo.Delete(ctx, cmd.TournamentID); err != nil {
		return err
	}

	return nil
}

// AddAttemptCommand contains parameters for adding tournament attempts.
type AddAttemptCommand struct {
	TournamentID shared.TournamentID
	PlayerID     shared.PlayerID
	Count        int
}

// AddAttempt adds attempts for a player in a tournament.
func (s *Service) AddAttempt(ctx context.Context, cmd AddAttemptCommand) error {
	if err := cmd.TournamentID.Validate(); err != nil {
		return err
	}
	if err := cmd.PlayerID.Validate(); err != nil {
		return err
	}
	if cmd.Count <= 0 {
		return tournament.ErrInvalidAttemptCount
	}

	now := s.Clock()

	// Get or create participant
	participant, err := s.Participants.Get(ctx, cmd.TournamentID, cmd.PlayerID)
	if err != nil {
		if err == tournament.ErrParticipantNotFound {
			participant, err = tournament.NewParticipant(cmd.TournamentID, cmd.PlayerID, now)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Add attempts
	if err := participant.AddAttempts(cmd.Count, now); err != nil {
		return err
	}

	// Save participant
	if err := s.Participants.Save(ctx, participant); err != nil {
		return err
	}

	// Update in Nakama
	if err := s.Provider.AddAttempt(ctx, cmd.TournamentID, cmd.PlayerID, cmd.Count); err != nil {
		return err
	}

	return nil
}

// GetTournamentQuery contains parameters for retrieving a tournament.
type GetTournamentQuery struct {
	TournamentID shared.TournamentID
}

// GetTournament retrieves a tournament by ID.
func (s *Service) GetTournament(ctx context.Context, query GetTournamentQuery) (*tournament.Tournament, error) {
	if err := query.TournamentID.Validate(); err != nil {
		return nil, err
	}

	return s.Repo.Get(ctx, query.TournamentID)
}

// ListTournamentsQuery contains parameters for listing tournaments.
type ListTournamentsQuery struct {
	Limit  int
	Offset int
}

// ListTournaments retrieves a paginated list of tournaments.
func (s *Service) ListTournaments(ctx context.Context, query ListTournamentsQuery) ([]*tournament.Tournament, error) {
	if query.Limit <= 0 {
		query.Limit = 10
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	return s.Repo.List(ctx, query.Limit, query.Offset)
}
