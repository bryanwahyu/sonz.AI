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
	"context"

	"github.com/heroiclabs/nakama/v3/src/app/analytics"
	"github.com/heroiclabs/nakama/v3/src/domain/shared"
	infraAnalytics "github.com/heroiclabs/nakama/v3/src/infra/analytics"
)

// TrackerAdapter adapts the new DDD structure to the legacy Tracker interface.
type TrackerAdapter struct {
	service *analytics.Service
}

// NewTrackerAdapter creates a new adapter using the DDD service layer.
func NewTrackerAdapter(key string, opts ...TrackerOption) *TrackerAdapter {
	// Create infrastructure dependencies
	dispatcher := infraAnalytics.NewSegmentDispatcher(key, defaultBaseURL)
	sessionRepo := infraAnalytics.NewMemorySessionRepository()

	// Apply options to dispatcher
	tracker := &Tracker{
		key:        key,
		baseURL:    defaultBaseURL,
		httpClient: defaultHTTPClient(),
	}
	for _, opt := range opts {
		opt(tracker)
	}

	if tracker.baseURL != defaultBaseURL {
		dispatcher.BaseURL = tracker.baseURL
	}
	if tracker.httpClient != nil {
		dispatcher.HTTPClient = tracker.httpClient
	}

	// Create service
	service := analytics.NewService(dispatcher, sessionRepo)

	return &TrackerAdapter{
		service: service,
	}
}

// StartSession starts a user session using the DDD service.
func (a *TrackerAdapter) StartSession(userID, version, variant string) error {
	ctx := context.Background()
	cmd := analytics.StartSessionCommand{
		UserID:  shared.PlayerID(userID),
		Version: version,
		Variant: variant,
	}
	return a.service.StartSession(ctx, cmd)
}

// EndSession ends a user session using the DDD service.
func (a *TrackerAdapter) EndSession(userID string) error {
	ctx := context.Background()
	cmd := analytics.EndSessionCommand{
		UserID: shared.PlayerID(userID),
	}
	return a.service.EndSession(ctx, cmd)
}

// StartWithAdapter is a convenience function using the adapter.
func StartWithAdapter(key, id, version, variant string) error {
	adapter := NewTrackerAdapter(key)
	return adapter.StartSession(id, version, variant)
}

// EndWithAdapter is a convenience function using the adapter.
func EndWithAdapter(key, id string) error {
	adapter := NewTrackerAdapter(key)
	return adapter.EndSession(id)
}
