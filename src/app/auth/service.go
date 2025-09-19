package auth

import (
	"context"
	"errors"
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/player"
	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// AuthResult wraps Nakama session tokens.
type AuthResult struct {
	UserID       shared.PlayerID
	SessionToken string
	RefreshToken string
	Username     string
}

// AuthProvider describes the Nakama authentication integration.
type AuthProvider interface {
	AuthenticateDevice(ctx context.Context, deviceID, username string, vars map[string]string) (AuthResult, error)
	AuthenticateEmail(ctx context.Context, email, password string, vars map[string]string) (AuthResult, error)
}

// PlayerRepository defines the persistence contract needed by the service.
type PlayerRepository interface {
	player.Repository
}

// Clock abstracts time for deterministic testing.
type Clock func() time.Time

// Service orchestrates player authentication flows on top of Nakama.
type Service struct {
	Repo  PlayerRepository
	Auth  AuthProvider
	Clock Clock
}

func NewService(repo PlayerRepository, authProvider AuthProvider) *Service {
	return &Service{
		Repo:  repo,
		Auth:  authProvider,
		Clock: func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) AuthenticateDevice(ctx context.Context, deviceID, username string, vars map[string]string) (AuthResult, error) {
	result, err := s.Auth.AuthenticateDevice(ctx, deviceID, username, vars)
	if err != nil {
		return AuthResult{}, err
	}
	now := s.Clock()
	account, err := s.Repo.GetByID(ctx, result.UserID)
	if err != nil {
		if !errors.Is(err, shared.ErrNotFound) {
			return AuthResult{}, err
		}
		account, err = player.NewPlayerAccount(result.UserID, vars["email"], username, now)
		if err != nil {
			return AuthResult{}, err
		}
	}
	_ = account.RegisterDevice(player.DeviceFingerprint{ID: deviceID, Platform: vars["platform"], LastSeen: now})
	account.RecordSession(player.SessionMetadata{SessionID: result.SessionToken, IssuedAt: now})
	if err := s.Repo.Save(ctx, account); err != nil {
		return AuthResult{}, err
	}
	return result, nil
}

func (s *Service) AuthenticateEmail(ctx context.Context, email, password string, vars map[string]string) (AuthResult, error) {
	result, err := s.Auth.AuthenticateEmail(ctx, email, password, vars)
	if err != nil {
		return AuthResult{}, err
	}
	now := s.Clock()
	account, err := s.Repo.GetByID(ctx, result.UserID)
	if err != nil {
		if !errors.Is(err, shared.ErrNotFound) {
			return AuthResult{}, err
		}
		account, err = player.NewPlayerAccount(result.UserID, email, result.Username, now)
		if err != nil {
			return AuthResult{}, err
		}
	}
	account.RecordSession(player.SessionMetadata{SessionID: result.SessionToken, IssuedAt: now})
	if err := s.Repo.Save(ctx, account); err != nil {
		return AuthResult{}, err
	}
	return result, nil
}
