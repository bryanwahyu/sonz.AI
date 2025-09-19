package group

import "context"

import "github.com/heroiclabs/nakama/v3/src/domain/shared"

type Repository interface {
	Get(ctx context.Context, id shared.GroupID) (*Group, error)
	Save(ctx context.Context, group *Group) error
	AddMember(ctx context.Context, groupID shared.GroupID, member Membership) error
}
