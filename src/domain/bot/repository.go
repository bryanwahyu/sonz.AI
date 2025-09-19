package bot

import "context"

import "github.com/heroiclabs/nakama/v3/src/domain/shared"

type Repository interface {
	ReserveCommand(ctx context.Context, key shared.IdempotencyKey) (*Command, error)
	Save(ctx context.Context, command *Command) error
	MarkProcessed(ctx context.Context, id shared.BotCommandID, state CommandState) error
}
