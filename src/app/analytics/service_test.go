package analytics_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/src/app/analytics"
	domainAnalytics "github.com/heroiclabs/nakama/v3/src/domain/analytics"
	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// Mock implementations
type mockDispatcher struct {
	dispatchFunc func(ctx context.Context, events []*domainAnalytics.Event) error
}

func (m *mockDispatcher) Dispatch(ctx context.Context, events []*domainAnalytics.Event) error {
	if m.dispatchFunc != nil {
		return m.dispatchFunc(ctx, events)
	}
	return nil
}

type mockSessionRepo struct {
	saveFunc   func(ctx context.Context, session *domainAnalytics.Session) error
	getFunc    func(ctx context.Context, userID shared.PlayerID) (*domainAnalytics.Session, error)
	deleteFunc func(ctx context.Context, userID shared.PlayerID) error
}

func (m *mockSessionRepo) Save(ctx context.Context, session *domainAnalytics.Session) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, session)
	}
	return nil
}

func (m *mockSessionRepo) Get(ctx context.Context, userID shared.PlayerID) (*domainAnalytics.Session, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, userID)
	}
	return nil, domainAnalytics.ErrSessionNotFound
}

func (m *mockSessionRepo) Delete(ctx context.Context, userID shared.PlayerID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, userID)
	}
	return nil
}

func TestService_StartSession(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		cmd            analytics.StartSessionCommand
		dispatchErr    error
		saveErr        error
		wantErr        bool
		wantEventCount int
	}{
		{
			name: "successful session start",
			cmd: analytics.StartSessionCommand{
				UserID:  "player-123",
				Version: "1.0.0",
				Variant: "production",
			},
			wantErr:        false,
			wantEventCount: 2, // identify + track
		},
		{
			name: "empty user id",
			cmd: analytics.StartSessionCommand{
				UserID:  "",
				Version: "1.0.0",
				Variant: "production",
			},
			wantErr: true,
		},
		{
			name: "empty version",
			cmd: analytics.StartSessionCommand{
				UserID:  "player-123",
				Version: "",
				Variant: "production",
			},
			wantErr: true,
		},
		{
			name: "dispatch failure",
			cmd: analytics.StartSessionCommand{
				UserID:  "player-123",
				Version: "1.0.0",
				Variant: "production",
			},
			dispatchErr: errors.New("dispatch failed"),
			wantErr:     true,
		},
		{
			name: "save failure",
			cmd: analytics.StartSessionCommand{
				UserID:  "player-123",
				Version: "1.0.0",
				Variant: "production",
			},
			saveErr: errors.New("save failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedEvents []*domainAnalytics.Event

			dispatcher := &mockDispatcher{
				dispatchFunc: func(ctx context.Context, events []*domainAnalytics.Event) error {
					capturedEvents = events
					return tt.dispatchErr
				},
			}

			sessionRepo := &mockSessionRepo{
				saveFunc: func(ctx context.Context, session *domainAnalytics.Session) error {
					return tt.saveErr
				},
			}

			service := analytics.NewService(dispatcher, sessionRepo)
			err := service.StartSession(ctx, tt.cmd)

			if (err != nil) != tt.wantErr {
				t.Errorf("StartSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.wantEventCount > 0 {
				if len(capturedEvents) != tt.wantEventCount {
					t.Errorf("Expected %d events, got %d", tt.wantEventCount, len(capturedEvents))
				}
			}
		})
	}
}

func TestService_EndSession(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name        string
		cmd         analytics.EndSessionCommand
		existingSession *domainAnalytics.Session
		getErr      error
		dispatchErr error
		wantErr     bool
	}{
		{
			name: "successful session end",
			cmd: analytics.EndSessionCommand{
				UserID: "player-123",
			},
			existingSession: &domainAnalytics.Session{
				UserID:    "player-123",
				State:     domainAnalytics.SessionStateActive,
				Version:   "1.0.0",
				Variant:   "production",
				StartedAt: now,
			},
			wantErr: false,
		},
		{
			name: "session not found",
			cmd: analytics.EndSessionCommand{
				UserID: "player-123",
			},
			getErr:  domainAnalytics.ErrSessionNotFound,
			wantErr: true,
		},
		{
			name: "empty user id",
			cmd: analytics.EndSessionCommand{
				UserID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispatcher := &mockDispatcher{
				dispatchFunc: func(ctx context.Context, events []*domainAnalytics.Event) error {
					return tt.dispatchErr
				},
			}

			sessionRepo := &mockSessionRepo{
				getFunc: func(ctx context.Context, userID shared.PlayerID) (*domainAnalytics.Session, error) {
					if tt.getErr != nil {
						return nil, tt.getErr
					}
					return tt.existingSession, nil
				},
				saveFunc: func(ctx context.Context, session *domainAnalytics.Session) error {
					return nil
				},
				deleteFunc: func(ctx context.Context, userID shared.PlayerID) error {
					return nil
				},
			}

			service := analytics.NewService(dispatcher, sessionRepo)
			err := service.EndSession(ctx, tt.cmd)

			if (err != nil) != tt.wantErr {
				t.Errorf("EndSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_TrackEvent(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		cmd         analytics.TrackEventCommand
		dispatchErr error
		wantErr     bool
	}{
		{
			name: "successful event tracking",
			cmd: analytics.TrackEventCommand{
				UserID:     "player-123",
				Name:       domainAnalytics.EventNameStart,
				AppName:    "MyApp",
				AppVersion: "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "empty user id",
			cmd: analytics.TrackEventCommand{
				UserID: "",
				Name:   domainAnalytics.EventNameStart,
			},
			wantErr: true,
		},
		{
			name: "dispatch failure",
			cmd: analytics.TrackEventCommand{
				UserID: "player-123",
				Name:   domainAnalytics.EventNameStart,
			},
			dispatchErr: errors.New("dispatch failed"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispatcher := &mockDispatcher{
				dispatchFunc: func(ctx context.Context, events []*domainAnalytics.Event) error {
					return tt.dispatchErr
				},
			}

			sessionRepo := &mockSessionRepo{}
			service := analytics.NewService(dispatcher, sessionRepo)

			err := service.TrackEvent(ctx, tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("TrackEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
