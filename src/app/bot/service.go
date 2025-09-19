package bot

import (
	"context"
	"errors"
	"time"

	domain "github.com/heroiclabs/nakama/v3/src/domain/bot"
	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

type Repository interface {
	domain.Repository
}

// QueueProducer describes the worker enqueue/outbox publisher.
type QueueProducer interface {
	Enqueue(ctx context.Context, command *domain.Command) error
}

// Notifier sends immediate acknowledgement to players or external systems.
type Notifier interface {
	Notify(ctx context.Context, playerID shared.PlayerID, payload map[string]any) error
}

// Service manages bot command ingestion and acknowledgement.
type Service struct {
	Repo     Repository
	Producer QueueProducer
	Notifier Notifier
	Clock    func() time.Time
}

func NewService(repo Repository, producer QueueProducer, notifier Notifier) *Service {
	return &Service{
		Repo:     repo,
		Producer: producer,
		Notifier: notifier,
		Clock:    func() time.Time { return time.Now().UTC() },
	}
}

type CommandInput struct {
	CommandID      shared.BotCommandID
	Channel        string
	PlayerID       shared.PlayerID
	Payload        []byte
	IdempotencyKey shared.IdempotencyKey
}

type CommandResult struct {
	Accepted bool
}

func (s *Service) Handle(ctx context.Context, input CommandInput) (CommandResult, error) {
	now := s.Clock()
	if existing, err := s.Repo.ReserveCommand(ctx, input.IdempotencyKey); err == nil {
		if existing.State == domain.CommandStateCompleted {
			return CommandResult{Accepted: true}, nil
		}
		return CommandResult{}, shared.ErrDuplicate
	} else if !errors.Is(err, shared.ErrNotFound) {
		return CommandResult{}, err
	}

	cmd, err := domain.NewCommand(input.CommandID, input.Channel, input.Payload, input.IdempotencyKey, now)
	if err != nil {
		return CommandResult{}, err
	}
	if err := s.Repo.Save(ctx, cmd); err != nil {
		return CommandResult{}, err
	}
	if s.Producer != nil {
		if err := s.Producer.Enqueue(ctx, cmd); err != nil {
			cmd.MarkAttempt(now, err)
			_ = s.Repo.Save(ctx, cmd)
			return CommandResult{}, err
		}
	}
	if s.Notifier != nil && input.PlayerID != "" {
		_ = s.Notifier.Notify(ctx, input.PlayerID, map[string]any{"status": "accepted"})
	}
	return CommandResult{Accepted: true}, nil
}
