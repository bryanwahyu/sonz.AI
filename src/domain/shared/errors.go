package shared

import "errors"

var (
	ErrDuplicate    = errors.New("duplicate operation")
	ErrNotFound     = errors.New("entity not found")
	ErrConflict     = errors.New("entity conflict")
	ErrInvalidState = errors.New("invalid state transition")
)
