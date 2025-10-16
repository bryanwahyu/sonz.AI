package tournament

import "errors"

var (
	ErrTournamentNotFound      = errors.New("tournament not found")
	ErrTournamentAlreadyExists = errors.New("tournament already exists")
	ErrTournamentAlreadyEnded  = errors.New("tournament already ended")
	ErrParticipantNotFound     = errors.New("participant not found")
	ErrParticipantAlreadyJoined = errors.New("participant already joined")
	ErrTournamentFull          = errors.New("tournament is full")
	ErrInvalidAttemptCount     = errors.New("invalid attempt count")
)
