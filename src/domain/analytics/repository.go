package analytics

import (
	"context"

	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// EventDispatcher sends events to external analytics services.
type EventDispatcher interface {
	Dispatch(ctx context.Context, events []*Event) error
}

// SessionRepository manages session persistence.
type SessionRepository interface {
	Save(ctx context.Context, session *Session) error
	Get(ctx context.Context, userID shared.PlayerID) (*Session, error)
	Delete(ctx context.Context, userID shared.PlayerID) error
}
