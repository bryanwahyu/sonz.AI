package groups

import (
	"context"
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/group"
	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// Provider defines the Nakama group API contract the application layer needs.
type Provider interface {
	CreateGroup(ctx context.Context, payload CreateGroupPayload) (CreateGroupResult, error)
	UpdateMetadata(ctx context.Context, groupID shared.GroupID, metadata map[string]any) error
}

type Repository interface {
	group.Repository
}

// CreateGroupPayload maps to Nakama group creation request.
type CreateGroupPayload struct {
	Name        string
	Description string
	CreatorID   shared.PlayerID
	AvatarURL   string
	LangTag     string
	Open        bool
}

// CreateGroupResult captures relevant Nakama response fields.
type CreateGroupResult struct {
	GroupID shared.GroupID
	Handle  string
}

// Service coordinates group creation with domain invariants.
type Service struct {
	Repo     Repository
	Provider Provider
	Clock    func() time.Time
}

func NewService(repo Repository, provider Provider) *Service {
	return &Service{
		Repo:     repo,
		Provider: provider,
		Clock:    func() time.Time { return time.Now().UTC() },
	}
}

type CreateInput struct {
	CreatorID   shared.PlayerID
	Name        string
	Description string
	Open        bool
	AvatarURL   string
	LangTag     string
}

type CreateOutput struct {
	GroupID shared.GroupID
	Handle  string
}

func (s *Service) CreateGroup(ctx context.Context, cmd CreateInput) (CreateOutput, error) {
	if err := cmd.CreatorID.Validate(); err != nil {
		return CreateOutput{}, err
	}
	now := s.Clock()
	payload := CreateGroupPayload{
		Name:        cmd.Name,
		Description: cmd.Description,
		CreatorID:   cmd.CreatorID,
		AvatarURL:   cmd.AvatarURL,
		LangTag:     cmd.LangTag,
		Open:        cmd.Open,
	}
	result, err := s.Provider.CreateGroup(ctx, payload)
	if err != nil {
		return CreateOutput{}, err
	}
	aggregate, err := group.NewGroup(result.GroupID, cmd.Name, cmd.CreatorID, now)
	if err != nil {
		return CreateOutput{}, err
	}
	aggregate.Description = cmd.Description
	if err := s.Repo.Save(ctx, aggregate); err != nil {
		return CreateOutput{}, err
	}
	return CreateOutput{GroupID: result.GroupID, Handle: result.Handle}, nil
}
