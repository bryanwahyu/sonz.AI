package tournament_test

import (
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/shared"
	"github.com/heroiclabs/nakama/v3/src/domain/tournament"
)

func TestNewTournament(t *testing.T) {
	now := time.Now()
	startTime := now.Add(1 * time.Hour)

	tests := []struct {
		name      string
		id        shared.TournamentID
		title     string
		category  int
		maxSize   int
		startTime time.Time
		duration  time.Duration
		wantErr   bool
	}{
		{
			name:      "valid tournament",
			id:        "tournament-123",
			title:     "Test Tournament",
			category:  1,
			maxSize:   100,
			startTime: startTime,
			duration:  24 * time.Hour,
			wantErr:   false,
		},
		{
			name:      "empty id",
			id:        "",
			title:     "Test Tournament",
			category:  1,
			maxSize:   100,
			startTime: startTime,
			duration:  24 * time.Hour,
			wantErr:   true,
		},
		{
			name:      "empty title",
			id:        "tournament-123",
			title:     "",
			category:  1,
			maxSize:   100,
			startTime: startTime,
			duration:  24 * time.Hour,
			wantErr:   true,
		},
		{
			name:      "negative category",
			id:        "tournament-123",
			title:     "Test Tournament",
			category:  -1,
			maxSize:   100,
			startTime: startTime,
			duration:  24 * time.Hour,
			wantErr:   true,
		},
		{
			name:      "negative max size",
			id:        "tournament-123",
			title:     "Test Tournament",
			category:  1,
			maxSize:   -1,
			startTime: startTime,
			duration:  24 * time.Hour,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tour, err := tournament.NewTournament(
				tt.id,
				tt.title,
				"Test Description",
				tt.category,
				tournament.SortOrderDescending,
				tournament.OperatorBest,
				"",
				true,
				false,
				tt.maxSize,
				10,
				tt.startTime,
				tt.duration,
				now,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewTournament() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tour.State != tournament.StateActive {
					t.Errorf("Expected state %v, got %v", tournament.StateActive, tour.State)
				}
				if !tour.IsActive() {
					t.Error("Expected tournament to be active")
				}
			}
		})
	}
}

func TestTournament_End(t *testing.T) {
	now := time.Now()
	startTime := now.Add(1 * time.Hour)

	tests := []struct {
		name    string
		endTime time.Time
		wantErr bool
	}{
		{
			name:    "valid end time",
			endTime: startTime.Add(2 * time.Hour),
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
			tour, _ := tournament.NewTournament(
				"tournament-123",
				"Test Tournament",
				"Description",
				1,
				tournament.SortOrderDescending,
				tournament.OperatorBest,
				"",
				true,
				false,
				100,
				10,
				startTime,
				24*time.Hour,
				now,
			)

			err := tour.End(tt.endTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("End() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tour.State != tournament.StateEnded {
					t.Errorf("Expected state %v, got %v", tournament.StateEnded, tour.State)
				}
				if tour.IsActive() {
					t.Error("Expected tournament to not be active")
				}
			}
		})
	}
}

func TestTournament_EndTwice(t *testing.T) {
	now := time.Now()
	startTime := now.Add(1 * time.Hour)
	tour, _ := tournament.NewTournament(
		"tournament-123",
		"Test Tournament",
		"Description",
		1,
		tournament.SortOrderDescending,
		tournament.OperatorBest,
		"",
		true,
		false,
		100,
		10,
		startTime,
		24*time.Hour,
		now,
	)

	err := tour.End(startTime.Add(2 * time.Hour))
	if err != nil {
		t.Fatalf("First End() failed: %v", err)
	}

	err = tour.End(startTime.Add(3 * time.Hour))
	if err != tournament.ErrTournamentAlreadyEnded {
		t.Errorf("Expected ErrTournamentAlreadyEnded, got %v", err)
	}
}

func TestTournament_Reset(t *testing.T) {
	now := time.Now()
	startTime := now.Add(1 * time.Hour)
	tour, _ := tournament.NewTournament(
		"tournament-123",
		"Test Tournament",
		"Description",
		1,
		tournament.SortOrderDescending,
		tournament.OperatorBest,
		"",
		true,
		false,
		100,
		10,
		startTime,
		24*time.Hour,
		now,
	)

	err := tour.Reset(now.Add(2 * time.Hour))
	if err != nil {
		t.Errorf("Reset() failed: %v", err)
	}

	if tour.State != tournament.StateReset {
		t.Errorf("Expected state %v, got %v", tournament.StateReset, tour.State)
	}
}

func TestTournament_CalculateEndTime(t *testing.T) {
	now := time.Now()
	startTime := now.Add(1 * time.Hour)
	duration := 24 * time.Hour

	tour, _ := tournament.NewTournament(
		"tournament-123",
		"Test Tournament",
		"Description",
		1,
		tournament.SortOrderDescending,
		tournament.OperatorBest,
		"",
		true,
		false,
		100,
		10,
		startTime,
		duration,
		now,
	)

	endTime := tour.CalculateEndTime()
	expected := startTime.Add(duration)

	if !endTime.Equal(expected) {
		t.Errorf("Expected end time %v, got %v", expected, endTime)
	}
}
