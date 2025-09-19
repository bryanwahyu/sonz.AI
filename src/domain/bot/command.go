package bot

import (
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

type CommandState string

const (
	CommandStatePending   CommandState = "pending"
	CommandStateCompleted CommandState = "completed"
	CommandStateFailed    CommandState = "failed"
)

// Command aggregate tracks dedupe and retry policies for bot automation.
type Command struct {
	ID             shared.BotCommandID
	Channel        string
	Payload        []byte
	IdempotencyKey shared.IdempotencyKey
	State          CommandState
	AttemptedAt    time.Time
	CompletedAt    time.Time
	RetryCount     int
	LastError      string
	CreatedAt      time.Time
}

func NewCommand(id shared.BotCommandID, channel string, payload []byte, key shared.IdempotencyKey, now time.Time) (*Command, error) {
	if err := id.Validate(); err != nil {
		return nil, err
	}
	if err := key.Validate(); err != nil {
		return nil, err
	}
	return &Command{
		ID:             id,
		Channel:        channel,
		Payload:        payload,
		IdempotencyKey: key,
		State:          CommandStatePending,
		CreatedAt:      now,
	}, nil
}

func (c *Command) MarkAttempt(now time.Time, err error) {
	c.AttemptedAt = now
	if err != nil {
		c.RetryCount++
		c.State = CommandStateFailed
		c.LastError = err.Error()
		return
	}
	c.State = CommandStateCompleted
	c.CompletedAt = now
	c.LastError = ""
}
