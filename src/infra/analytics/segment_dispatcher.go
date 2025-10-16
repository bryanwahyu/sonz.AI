package analytics

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/analytics"
)

// SegmentDispatcher implements EventDispatcher for Segment.io.
type SegmentDispatcher struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// NewSegmentDispatcher creates a new Segment dispatcher.
func NewSegmentDispatcher(apiKey, baseURL string) *SegmentDispatcher {
	if baseURL == "" {
		baseURL = "https://api.segment.io/v1/batch"
	}
	return &SegmentDispatcher{
		APIKey:  apiKey,
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// WithHTTPClient sets a custom HTTP client.
func (d *SegmentDispatcher) WithHTTPClient(client *http.Client) *SegmentDispatcher {
	d.HTTPClient = client
	return d
}

// segmentEvent represents the Segment API event format.
type segmentEvent struct {
	Type    string                 `json:"type"`
	UserID  string                 `json:"userId"`
	Event   string                 `json:"event,omitempty"`
	Context map[string]interface{} `json:"context"`
	App     *segmentApp            `json:"app,omitempty"`
	OS      *segmentOS             `json:"os,omitempty"`
}

type segmentApp struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type segmentOS struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type segmentBatch struct {
	Batch []segmentEvent `json:"batch"`
}

// Dispatch sends events to Segment.
func (d *SegmentDispatcher) Dispatch(ctx context.Context, events []*analytics.Event) error {
	if len(events) == 0 {
		return nil
	}

	segmentEvents := make([]segmentEvent, 0, len(events))
	for _, event := range events {
		if err := event.Validate(); err != nil {
			return err
		}

		se := segmentEvent{
			Type:   string(event.Type),
			UserID: string(event.UserID),
			Context: map[string]interface{}{
				"direct": event.Context.Direct,
				"library": map[string]string{
					"name":    event.Context.Library.Name,
					"version": event.Context.Library.Version,
				},
			},
		}

		if event.Type == analytics.EventTypeTrack {
			se.Event = string(event.Name)
		}

		if event.App != nil {
			se.App = &segmentApp{
				Name:    event.App.Name,
				Version: event.App.Version,
			}
		}

		if event.OS != nil {
			se.OS = &segmentOS{
				Name:    event.OS.Name,
				Version: event.OS.Version,
			}
		}

		segmentEvents = append(segmentEvents, se)
	}

	batch := segmentBatch{Batch: segmentEvents}
	body, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.BaseURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(d.APIKey, "")

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return analytics.ErrDispatchFailed
	}

	return nil
}
