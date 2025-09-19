package shared

import (
	"errors"
	"strings"
)

// ID types keep domain entities distinct while remaining simple strings at runtime.
type (
	PlayerID       string
	GroupID        string
	BattleID       string
	SeasonID       string
	BotCommandID   string
	IdempotencyKey string
)

// Validate ensures IDs are not blank and normalized.
func (id PlayerID) Validate() error {
	if strings.TrimSpace(string(id)) == "" {
		return errors.New("player id is required")
	}
	return nil
}

func (id GroupID) Validate() error {
	if strings.TrimSpace(string(id)) == "" {
		return errors.New("group id is required")
	}
	return nil
}

func (id BattleID) Validate() error {
	if strings.TrimSpace(string(id)) == "" {
		return errors.New("battle id is required")
	}
	return nil
}

func (id SeasonID) Validate() error {
	if strings.TrimSpace(string(id)) == "" {
		return errors.New("season id is required")
	}
	return nil
}

func (id BotCommandID) Validate() error {
	if strings.TrimSpace(string(id)) == "" {
		return errors.New("bot command id is required")
	}
	return nil
}

func (key IdempotencyKey) Validate() error {
	if strings.TrimSpace(string(key)) == "" {
		return errors.New("idempotency key is required")
	}
	return nil
}
