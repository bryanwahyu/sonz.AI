package battle

import "errors"

var (
	ErrPlayerAlreadyJoined = errors.New("player already joined battle")
	ErrPlayerNotFound      = errors.New("player not in battle")
)
