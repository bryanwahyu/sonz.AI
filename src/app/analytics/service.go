package analytics

import (
	"context"
	"runtime"
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/analytics"
	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// Service coordinates analytics operations.
type Service struct {
	Dispatcher analytics.EventDispatcher
	Sessions   analytics.SessionRepository
	Clock      func() time.Time
	ContextFactory func() analytics.Context
}

// NewService creates a new analytics service.
func NewService(dispatcher analytics.EventDispatcher, sessions analytics.SessionRepository) *Service {
	return &Service{
		Dispatcher: dispatcher,
		Sessions:   sessions,
		Clock:      func() time.Time { return time.Now().UTC() },
		ContextFactory: defaultContextFactory,
	}
}

// StartSessionCommand contains parameters for starting a session.
type StartSessionCommand struct {
	UserID  shared.PlayerID
	Version string
	Variant string
}

// StartSession initiates a user session and dispatches tracking events.
func (s *Service) StartSession(ctx context.Context, cmd StartSessionCommand) error {
	if err := cmd.UserID.Validate(); err != nil {
		return err
	}
	if cmd.Version == "" {
		return analytics.ErrInvalidEvent
	}

	now := s.Clock()
	session, err := analytics.NewSession(cmd.UserID, cmd.Version, cmd.Variant, now)
	if err != nil {
		return err
	}

	// Save session
	if err := s.Sessions.Save(ctx, session); err != nil {
		return err
	}

	// Create events
	context := s.ContextFactory()
	identifyEvent, err := analytics.NewIdentifyEvent(cmd.UserID, context, now)
	if err != nil {
		return err
	}

	trackEvent, err := analytics.NewTrackEvent(cmd.UserID, analytics.EventNameStart, context, now)
	if err != nil {
		return err
	}
	trackEvent.WithAppInfo(cmd.Variant, cmd.Version).WithOSInfo(runtime.GOOS, runtime.GOARCH)

	// Dispatch events
	events := []*analytics.Event{identifyEvent, trackEvent}
	if err := s.Dispatcher.Dispatch(ctx, events); err != nil {
		return analytics.ErrDispatchFailed
	}

	return nil
}

// EndSessionCommand contains parameters for ending a session.
type EndSessionCommand struct {
	UserID shared.PlayerID
}

// EndSession terminates a user session and dispatches tracking event.
func (s *Service) EndSession(ctx context.Context, cmd EndSessionCommand) error {
	if err := cmd.UserID.Validate(); err != nil {
		return err
	}

	// Get existing session
	session, err := s.Sessions.Get(ctx, cmd.UserID)
	if err != nil {
		return err
	}

	now := s.Clock()
	if err := session.End(now); err != nil {
		return err
	}

	// Update session
	if err := s.Sessions.Save(ctx, session); err != nil {
		return err
	}

	// Create end event
	context := s.ContextFactory()
	trackEvent, err := analytics.NewTrackEvent(cmd.UserID, analytics.EventNameEnd, context, now)
	if err != nil {
		return err
	}

	// Dispatch event
	events := []*analytics.Event{trackEvent}
	if err := s.Dispatcher.Dispatch(ctx, events); err != nil {
		return analytics.ErrDispatchFailed
	}

	// Clean up session
	_ = s.Sessions.Delete(ctx, cmd.UserID)

	return nil
}

// TrackEventCommand contains parameters for tracking custom events.
type TrackEventCommand struct {
	UserID  shared.PlayerID
	Name    analytics.EventName
	AppName string
	AppVersion string
	OSName  string
	OSVersion string
}

// TrackEvent dispatches a custom tracking event.
func (s *Service) TrackEvent(ctx context.Context, cmd TrackEventCommand) error {
	if err := cmd.UserID.Validate(); err != nil {
		return err
	}

	now := s.Clock()
	context := s.ContextFactory()
	
	event, err := analytics.NewTrackEvent(cmd.UserID, cmd.Name, context, now)
	if err != nil {
		return err
	}

	if cmd.AppName != "" || cmd.AppVersion != "" {
		event.WithAppInfo(cmd.AppName, cmd.AppVersion)
	}
	if cmd.OSName != "" || cmd.OSVersion != "" {
		event.WithOSInfo(cmd.OSName, cmd.OSVersion)
	}

	events := []*analytics.Event{event}
	if err := s.Dispatcher.Dispatch(ctx, events); err != nil {
		return analytics.ErrDispatchFailed
	}

	return nil
}

func defaultContextFactory() analytics.Context {
	return analytics.Context{
		Direct: true,
		Library: analytics.LibraryInfo{
			Name:    "go",
			Version: runtime.Version(),
		},
	}
}
