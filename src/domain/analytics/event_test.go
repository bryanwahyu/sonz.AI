package analytics_test

import (
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/analytics"
	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

func TestNewIdentifyEvent(t *testing.T) {
	tests := []struct {
		name      string
		userID    shared.PlayerID
		timestamp time.Time
		wantErr   bool
	}{
		{
			name:      "valid identify event",
			userID:    "player-123",
			timestamp: time.Now(),
			wantErr:   false,
		},
		{
			name:      "empty user id",
			userID:    "",
			timestamp: time.Now(),
			wantErr:   true,
		},
		{
			name:      "zero timestamp",
			userID:    "player-123",
			timestamp: time.Time{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := analytics.Context{
				Direct: true,
				Library: analytics.LibraryInfo{
					Name:    "go",
					Version: "1.21",
				},
			}

			event, err := analytics.NewIdentifyEvent(tt.userID, ctx, tt.timestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewIdentifyEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if event.Type != analytics.EventTypeIdentify {
					t.Errorf("Expected type %v, got %v", analytics.EventTypeIdentify, event.Type)
				}
				if event.UserID != tt.userID {
					t.Errorf("Expected userID %v, got %v", tt.userID, event.UserID)
				}
			}
		})
	}
}

func TestNewTrackEvent(t *testing.T) {
	tests := []struct {
		name      string
		userID    shared.PlayerID
		eventName analytics.EventName
		timestamp time.Time
		wantErr   bool
	}{
		{
			name:      "valid track event",
			userID:    "player-123",
			eventName: analytics.EventNameStart,
			timestamp: time.Now(),
			wantErr:   false,
		},
		{
			name:      "empty event name",
			userID:    "player-123",
			eventName: "",
			timestamp: time.Now(),
			wantErr:   true,
		},
		{
			name:      "empty user id",
			userID:    "",
			eventName: analytics.EventNameStart,
			timestamp: time.Now(),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := analytics.Context{Direct: true}

			event, err := analytics.NewTrackEvent(tt.userID, tt.eventName, ctx, tt.timestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTrackEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if event.Type != analytics.EventTypeTrack {
					t.Errorf("Expected type %v, got %v", analytics.EventTypeTrack, event.Type)
				}
				if event.Name != tt.eventName {
					t.Errorf("Expected name %v, got %v", tt.eventName, event.Name)
				}
			}
		})
	}
}

func TestEvent_WithAppInfo(t *testing.T) {
	ctx := analytics.Context{Direct: true}
	event, _ := analytics.NewTrackEvent("player-123", analytics.EventNameStart, ctx, time.Now())

	event.WithAppInfo("MyApp", "1.0.0")

	if event.App == nil {
		t.Fatal("Expected App to be set")
	}
	if event.App.Name != "MyApp" {
		t.Errorf("Expected app name MyApp, got %v", event.App.Name)
	}
	if event.App.Version != "1.0.0" {
		t.Errorf("Expected app version 1.0.0, got %v", event.App.Version)
	}
}

func TestEvent_WithOSInfo(t *testing.T) {
	ctx := analytics.Context{Direct: true}
	event, _ := analytics.NewTrackEvent("player-123", analytics.EventNameStart, ctx, time.Now())

	event.WithOSInfo("linux", "5.15")

	if event.OS == nil {
		t.Fatal("Expected OS to be set")
	}
	if event.OS.Name != "linux" {
		t.Errorf("Expected OS name linux, got %v", event.OS.Name)
	}
	if event.OS.Version != "5.15" {
		t.Errorf("Expected OS version 5.15, got %v", event.OS.Version)
	}
}

func TestEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   *analytics.Event
		wantErr bool
	}{
		{
			name: "valid identify event",
			event: &analytics.Event{
				Type:      analytics.EventTypeIdentify,
				UserID:    "player-123",
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid track event",
			event: &analytics.Event{
				Type:      analytics.EventTypeTrack,
				UserID:    "player-123",
				Name:      analytics.EventNameStart,
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "track event without name",
			event: &analytics.Event{
				Type:      analytics.EventTypeTrack,
				UserID:    "player-123",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "event without type",
			event: &analytics.Event{
				UserID:    "player-123",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
