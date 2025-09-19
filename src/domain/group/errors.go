package group

import "errors"

var (
	ErrNameRequired   = errors.New("group name required")
	ErrMemberNotFound = errors.New("group member not found")
)
