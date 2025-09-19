package group

import (
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type Membership struct {
	PlayerID shared.PlayerID
	Role     Role
	JoinedAt time.Time
}

// Group aggregate models membership and role policies.
type Group struct {
	ID          shared.GroupID
	Name        string
	Description string
	Members     map[shared.PlayerID]Membership
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewGroup(id shared.GroupID, name string, owner shared.PlayerID, now time.Time) (*Group, error) {
	if err := id.Validate(); err != nil {
		return nil, err
	}
	if err := owner.Validate(); err != nil {
		return nil, err
	}
	if name == "" {
		return nil, ErrNameRequired
	}
	g := &Group{
		ID:        id,
		Name:      name,
		Members:   make(map[shared.PlayerID]Membership),
		CreatedAt: now,
		UpdatedAt: now,
	}
	g.Members[owner] = Membership{PlayerID: owner, Role: RoleOwner, JoinedAt: now}
	return g, nil
}

func (g *Group) AssignRole(playerID shared.PlayerID, role Role, now time.Time) error {
	if _, ok := g.Members[playerID]; !ok {
		return ErrMemberNotFound
	}
	member := g.Members[playerID]
	member.Role = role
	member.JoinedAt = member.JoinedAt
	g.Members[playerID] = member
	g.UpdatedAt = now
	return nil
}
