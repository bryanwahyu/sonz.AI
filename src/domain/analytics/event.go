package analytics

import (
	"errors"
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// EventType defines the category of analytics event.
type EventType string

const (
	EventTypeIdentify EventType = "identify"
	EventTypeTrack    EventType = "track"
)

// EventName defines specific event names.
type EventName string

const (
	EventNameStart EventName = "start"
	EventNameEnd   EventName = "end"
)

// Context represents metadata attached to every event.
type Context struct {
	Direct  bool
	Library LibraryInfo
}

// LibraryInfo captures client library information.
type LibraryInfo struct {
	Name    string
	Version string
}

// AppInfo describes the application.
type AppInfo struct {
	Name    string
	Version string
}

// OSInfo describes the operating system.
type OSInfo struct {
	Name    string
	Version string
}

// Event is the domain aggregate for analytics events.
type Event struct {
	Type      EventType
	UserID    shared.PlayerID
	Name      EventName
	Context   Context
	App       *AppInfo
	OS        *OSInfo
	Timestamp time.Time
}

// NewIdentifyEvent creates an identity event.
func NewIdentifyEvent(userID shared.PlayerID, ctx Context, timestamp time.Time) (*Event, error) {
	if err := userID.Validate(); err != nil {
		return nil, err
	}
	if timestamp.IsZero() {
		return nil, errors.New("timestamp cannot be zero")
	}
	return &Event{
		Type:      EventTypeIdentify,
		UserID:    userID,
		Context:   ctx,
		Timestamp: timestamp,
	}, nil
}

// NewTrackEvent creates a tracking event with optional metadata.
func NewTrackEvent(userID shared.PlayerID, name EventName, ctx Context, timestamp time.Time) (*Event, error) {
	if err := userID.Validate(); err != nil {
		return nil, err
	}
	if name == "" {
		return nil, errors.New("event name cannot be empty")
	}
	if timestamp.IsZero() {
		return nil, errors.New("timestamp cannot be zero")
	}
	return &Event{
		Type:      EventTypeTrack,
		UserID:    userID,
		Name:      name,
		Context:   ctx,
		Timestamp: timestamp,
	}, nil
}

// WithAppInfo attaches application metadata.
func (e *Event) WithAppInfo(name, version string) *Event {
	e.App = &AppInfo{
		Name:    name,
		Version: version,
	}
	return e
}

// WithOSInfo attaches OS metadata.
func (e *Event) WithOSInfo(name, version string) *Event {
	e.OS = &OSInfo{
		Name:    name,
		Version: version,
	}
	return e
}

// Validate ensures the event is well-formed.
func (e *Event) Validate() error {
	if e.Type == "" {
		return errors.New("event type is required")
	}
	if err := e.UserID.Validate(); err != nil {
		return err
	}
	if e.Type == EventTypeTrack && e.Name == "" {
		return errors.New("track events require a name")
	}
	if e.Timestamp.IsZero() {
		return errors.New("timestamp is required")
	}
	return nil
}
