package analytics_test

import (
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/analytics"
	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

func TestNewSession(t *testing.T) {
	tests := []struct {
		name      string
		userID    shared.PlayerID
		version   string
		variant   string
		startedAt time.Time
		wantErr   bool
	}{
		{
			name:      "valid session",
			userID:    "player-123",
			version:   "1.0.0",
			variant:   "production",
			startedAt: time.Now(),
			wantErr:   false,
		},
		{
			name:      "empty user id",
			userID:    "",
			version:   "1.0.0",
			variant:   "production",
			startedAt: time.Now(),
			wantErr:   true,
		},
		{
			name:      "empty version",
			userID:    "player-123",
			version:   "",
			variant:   "production",
			startedAt: time.Now(),
			wantErr:   true,
		},
		{
			name:      "zero start time",
			userID:    "player-123",
			version:   "1.0.0",
			variant:   "production",
			startedAt: time.Time{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := analytics.NewSession(tt.userID, tt.version, tt.variant, tt.startedAt)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if session.State != analytics.SessionStateActive {
					t.Errorf("Expected state %v, got %v", analytics.SessionStateActive, session.State)
				}
				if !session.IsActive() {
					t.Error("Expected session to be active")
				}
			}
		})
	}
}

func TestSession_End(t *testing.T) {
	startTime := time.Now()

	tests := []struct {
		name    string
		endTime time.Time
		wantErr bool
	}{
		{
			name:    "valid end time",
			endTime: startTime.Add(1 * time.Hour),
			wantErr: false,
		},
		{
			name:    "end time before start",
			endTime: startTime.Add(-1 * time.Hour),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := analytics.NewSession("player-123", "1.0.0", "prod", startTime)
			err := s.End(tt.endTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("End() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if s.State != analytics.SessionStateEnded {
					t.Errorf("Expected state %v, got %v", analytics.SessionStateEnded, s.State)
				}
				if s.IsActive() {
					t.Error("Expected session to not be active")
				}
			}
		})
	}
}

func TestSession_EndTwice(t *testing.T) {
	startTime := time.Now()
	session, _ := analytics.NewSession("player-123", "1.0.0", "prod", startTime)

	err := session.End(startTime.Add(1 * time.Hour))
	if err != nil {
		t.Fatalf("First End() failed: %v", err)
	}

	err = session.End(startTime.Add(2 * time.Hour))
	if err == nil {
		t.Error("Expected error when ending session twice")
	}
}

func TestSession_Duration(t *testing.T) {
	startTime := time.Now()
	session, _ := analytics.NewSession("player-123", "1.0.0", "prod", startTime)

	// Test duration for active session
	duration := session.Duration()
	if duration < 0 {
		t.Error("Duration should be non-negative for active session")
	}

	// Test duration for ended session
	endTime := startTime.Add(2 * time.Hour)
	session.End(endTime)
	duration = session.Duration()
	expected := 2 * time.Hour
	if duration != expected {
		t.Errorf("Expected duration %v, got %v", expected, duration)
	}
}
