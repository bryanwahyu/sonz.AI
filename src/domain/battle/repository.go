package battle

import "context"

import "github.com/heroiclabs/nakama/v3/src/domain/shared"

type Repository interface {
	Get(ctx context.Context, id shared.BattleID) (*Battle, error)
	Save(ctx context.Context, battle *Battle) error
	StoreSnapshot(ctx context.Context, id shared.BattleID, state MatchState) error
}
