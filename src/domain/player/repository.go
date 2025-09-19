package player

import "context"

import "github.com/heroiclabs/nakama/v3/src/domain/shared"

type Repository interface {
	GetByID(ctx context.Context, id shared.PlayerID) (*PlayerAccount, error)
	Save(ctx context.Context, account *PlayerAccount) error
	AppendSession(ctx context.Context, id shared.PlayerID, session SessionMetadata) error
}
