package analytics

import (
	"context"
	"sync"

	"github.com/heroiclabs/nakama/v3/src/domain/analytics"
	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// MemorySessionRepository implements SessionRepository using in-memory storage.
type MemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[shared.PlayerID]*analytics.Session
}

// NewMemorySessionRepository creates a new in-memory session repository.
func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{
		sessions: make(map[shared.PlayerID]*analytics.Session),
	}
}

// Save stores a session.
func (r *MemorySessionRepository) Save(ctx context.Context, session *analytics.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[session.UserID] = session
	return nil
}

// Get retrieves a session by user ID.
func (r *MemorySessionRepository) Get(ctx context.Context, userID shared.PlayerID) (*analytics.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[userID]
	if !exists {
		return nil, analytics.ErrSessionNotFound
	}

	return session, nil
}

// Delete removes a session.
func (r *MemorySessionRepository) Delete(ctx context.Context, userID shared.PlayerID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.sessions, userID)
	return nil
}
