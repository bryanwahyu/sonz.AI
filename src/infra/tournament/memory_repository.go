package tournament

import (
	"context"
	"sync"

	"github.com/heroiclabs/nakama/v3/src/domain/shared"
	"github.com/heroiclabs/nakama/v3/src/domain/tournament"
)

// MemoryRepository implements tournament.Repository using in-memory storage.
type MemoryRepository struct {
	mu          sync.RWMutex
	tournaments map[shared.TournamentID]*tournament.Tournament
}

// NewMemoryRepository creates a new in-memory tournament repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		tournaments: make(map[shared.TournamentID]*tournament.Tournament),
	}
}

// Save stores a tournament.
func (r *MemoryRepository) Save(ctx context.Context, t *tournament.Tournament) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tournaments[t.ID] = t
	return nil
}

// Get retrieves a tournament by ID.
func (r *MemoryRepository) Get(ctx context.Context, id shared.TournamentID) (*tournament.Tournament, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.tournaments[id]
	if !exists {
		return nil, tournament.ErrTournamentNotFound
	}

	return t, nil
}

// Delete removes a tournament.
func (r *MemoryRepository) Delete(ctx context.Context, id shared.TournamentID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.tournaments, id)
	return nil
}

// List retrieves a paginated list of tournaments.
func (r *MemoryRepository) List(ctx context.Context, limit, offset int) ([]*tournament.Tournament, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tournaments := make([]*tournament.Tournament, 0, len(r.tournaments))
	for _, t := range r.tournaments {
		tournaments = append(tournaments, t)
	}

	// Apply pagination
	start := offset
	if start > len(tournaments) {
		return []*tournament.Tournament{}, nil
	}

	end := start + limit
	if end > len(tournaments) {
		end = len(tournaments)
	}

	return tournaments[start:end], nil
}

// MemoryParticipantRepository implements ParticipantRepository using in-memory storage.
type MemoryParticipantRepository struct {
	mu           sync.RWMutex
	participants map[string]*tournament.Participant // key: "tournamentID:playerID"
}

// NewMemoryParticipantRepository creates a new in-memory participant repository.
func NewMemoryParticipantRepository() *MemoryParticipantRepository {
	return &MemoryParticipantRepository{
		participants: make(map[string]*tournament.Participant),
	}
}

func makeKey(tournamentID shared.TournamentID, playerID shared.PlayerID) string {
	return string(tournamentID) + ":" + string(playerID)
}

// Save stores a participant.
func (r *MemoryParticipantRepository) Save(ctx context.Context, p *tournament.Participant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := makeKey(p.TournamentID, p.PlayerID)
	r.participants[key] = p
	return nil
}

// Get retrieves a participant.
func (r *MemoryParticipantRepository) Get(ctx context.Context, tournamentID shared.TournamentID, playerID shared.PlayerID) (*tournament.Participant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := makeKey(tournamentID, playerID)
	p, exists := r.participants[key]
	if !exists {
		return nil, tournament.ErrParticipantNotFound
	}

	return p, nil
}

// ListByTournament retrieves all participants for a tournament.
func (r *MemoryParticipantRepository) ListByTournament(ctx context.Context, tournamentID shared.TournamentID) ([]*tournament.Participant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	participants := make([]*tournament.Participant, 0)
	for _, p := range r.participants {
		if p.TournamentID == tournamentID {
			participants = append(participants, p)
		}
	}

	return participants, nil
}

// Delete removes a participant.
func (r *MemoryParticipantRepository) Delete(ctx context.Context, tournamentID shared.TournamentID, playerID shared.PlayerID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := makeKey(tournamentID, playerID)
	delete(r.participants, key)
	return nil
}
