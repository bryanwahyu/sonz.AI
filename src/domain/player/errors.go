package player

import "errors"

var (
	ErrEmailRequired    = errors.New("player email is required")
	ErrAccountSuspended = errors.New("player account suspended")
	ErrDeviceInvalid    = errors.New("device fingerprint invalid")
)
