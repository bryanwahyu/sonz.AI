package analytics

import "errors"

var (
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionAlreadyEnded = errors.New("session already ended")
	ErrInvalidEvent       = errors.New("invalid event")
	ErrDispatchFailed     = errors.New("failed to dispatch events")
)
