// Copyright 2024 The Nakama Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package se

import (
	"bytes"
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

// EventType defines a Segment event category.
type EventType string

const (
	defaultBaseURL = "https://api.segment.io/v1/batch"
)

const (
	EventTypeIdentify EventType = "identify"
	EventTypeTrack    EventType = "track"
)

const (
	EventStart = "start"
	EventEnd   = "end"
)

// Context represents the shared metadata sent with every Segment event.
type Context struct {
	Direct  bool        `json:"direct"`
	Library LibraryInfo `json:"library"`
}

// LibraryInfo captures the client library that produced the event.
type LibraryInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// AppInfo describes the application that produced the event.
type AppInfo struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// OSInfo describes the operating system that produced the event.
type OSInfo struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// Event models a Segment event payload.
type Event struct {
	Type    EventType `json:"type,omitempty"`
	UserID  string    `json:"userId,omitempty"`
	Name    string    `json:"event,omitempty"`
	Context Context   `json:"context,omitempty"`
	App     *AppInfo  `json:"app,omitempty"`
	OS      *OSInfo   `json:"os,omitempty"`
}

// Batch groups events in a single request.
type Batch struct {
	Events []Event `json:"batch,omitempty"`
}

// Tracker orchestrates Segment events and dispatches them through the Segment batch API.
type Tracker struct {
	key            string
	baseURL        string
	httpClient     *http.Client
	contextFactory func() Context
}

// TrackerOption configures a Tracker instance.
type TrackerOption func(*Tracker)

// NewTracker builds a Tracker with optional configuration.
func NewTracker(key string, opts ...TrackerOption) *Tracker {
	tracker := &Tracker{
		key:            key,
		baseURL:        defaultBaseURL,
		httpClient:     defaultHTTPClient(),
		contextFactory: defaultContext,
	}
	for _, opt := range opts {
		opt(tracker)
	}
	return tracker
}

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(client *http.Client) TrackerOption {
	return func(t *Tracker) {
		if client != nil {
			t.httpClient = client
		}
	}
}

// WithBaseURL overrides the default Segment batch endpoint.
func WithBaseURL(url string) TrackerOption {
	return func(t *Tracker) {
		if url != "" {
			t.baseURL = url
		}
	}
}

// WithContextFactory overrides the default context factory.
func WithContextFactory(factory func() Context) TrackerOption {
	return func(t *Tracker) {
		if factory != nil {
			t.contextFactory = factory
		}
	}
}

// StartSession records a user identity and a start track event in one batch.
func (t *Tracker) StartSession(userID, version, variant string) error {
	events := []Event{
		NewIdentifyEvent(userID, t.contextFactory()),
		NewTrackEvent(userID, EventStart, t.contextFactory(),
			WithAppInfo(variant, version),
			WithOSInfo(runtime.GOOS, runtime.GOARCH),
		),
	}
	return t.dispatch(events)
}

// EndSession records a track event indicating the end of a session.
func (t *Tracker) EndSession(userID string) error {
	events := []Event{
		NewTrackEvent(userID, EventEnd, t.contextFactory()),
	}
	return t.dispatch(events)
}

func (t *Tracker) dispatch(events []Event) error {
	batch := Batch{Events: events}
	body, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, t.baseURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(t.key, "")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		_ = resp.Body.Close()
	}

	return nil
}

// EventOption modifies a newly created event.
type EventOption func(*Event)

// NewIdentifyEvent constructs an identify event.
func NewIdentifyEvent(userID string, ctx Context) Event {
	return Event{
		Type:    EventTypeIdentify,
		UserID:  userID,
		Context: ctx,
	}
}

// NewTrackEvent constructs a track event with optional modifiers.
func NewTrackEvent(userID, name string, ctx Context, opts ...EventOption) Event {
	event := Event{
		Type:    EventTypeTrack,
		UserID:  userID,
		Name:    name,
		Context: ctx,
	}
	for _, opt := range opts {
		opt(&event)
	}
	return event
}

// WithAppInfo attaches application metadata to an event.
func WithAppInfo(name, version string) EventOption {
	return func(event *Event) {
		event.App = &AppInfo{
			Name:    name,
			Version: version,
		}
	}
}

// WithOSInfo attaches operating system metadata to an event.
func WithOSInfo(name, version string) EventOption {
	return func(event *Event) {
		event.OS = &OSInfo{
			Name:    name,
			Version: version,
		}
	}
}

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
	}
}

func defaultContext() Context {
	return Context{
		Direct: true,
		Library: LibraryInfo{
			Name:    "go",
			Version: runtime.Version(),
		},
	}
}

// Start is a convenience helper mirroring the previous API.
func Start(key, id, version, variant string) error {
	tracker := NewTracker(key)
	return tracker.StartSession(id, version, variant)
}

// End is a convenience helper mirroring the previous API.
func End(key, id string) error {
	tracker := NewTracker(key)
	return tracker.EndSession(id)
}
