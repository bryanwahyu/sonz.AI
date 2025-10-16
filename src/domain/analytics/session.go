package analytics

import (
	"errors"
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// SessionState represents the lifecycle state of a user session.
type SessionState string

const (
	SessionStateActive SessionState = "active"
	SessionStateEnded  SessionState = "ended"
)

// Session aggregate tracks user session lifecycle.
type Session struct {
	UserID    shared.PlayerID
	State     SessionState
	Version   string
	Variant   string
	StartedAt time.Time
	EndedAt   *time.Time
}

// NewSession creates a new active session.
func NewSession(userID shared.PlayerID, version, variant string, startedAt time.Time) (*Session, error) {
	if err := userID.Validate(); err != nil {
		return nil, err
	}
	if version == "" {
		return nil, errors.New("version is required")
	}
	if startedAt.IsZero() {
		return nil, errors.New("start time is required")
	}
	return &Session{
		UserID:    userID,
		State:     SessionStateActive,
		Version:   version,
		Variant:   variant,
		StartedAt: startedAt,
	}, nil
}

// End marks the session as ended.
func (s *Session) End(endedAt time.Time) error {
	if s.State == SessionStateEnded {
		return errors.New("session already ended")
	}
	if endedAt.Before(s.StartedAt) {
		return errors.New("end time cannot be before start time")
	}
	s.State = SessionStateEnded
	s.EndedAt = &endedAt
	return nil
}

// IsActive checks if the session is currently active.
func (s *Session) IsActive() bool {
	return s.State == SessionStateActive
}

// Duration calculates the session duration.
func (s *Session) Duration() time.Duration {
	if s.EndedAt == nil {
		return time.Since(s.StartedAt)
	}
	return s.EndedAt.Sub(s.StartedAt)
}
