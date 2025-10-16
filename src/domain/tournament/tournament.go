package tournament

import (
	"errors"
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// SortOrder defines how tournament scores are sorted.
type SortOrder string

const (
	SortOrderAscending  SortOrder = "asc"
	SortOrderDescending SortOrder = "desc"
)

// Operator defines score comparison logic.
type Operator string

const (
	OperatorBest       Operator = "best"
	OperatorSet        Operator = "set"
	OperatorIncrement  Operator = "incr"
	OperatorDecrement  Operator = "decr"
)

// TournamentState represents the lifecycle state.
type TournamentState string

const (
	StateActive   TournamentState = "active"
	StateEnded    TournamentState = "ended"
	StateReset    TournamentState = "reset"
)

// Tournament aggregate represents a competitive event.
type Tournament struct {
	ID            shared.TournamentID
	Title         string
	Description   string
	Category      int
	SortOrder     SortOrder
	Operator      Operator
	ResetSchedule string
	Authoritative bool
	JoinRequired  bool
	MaxSize       int
	MaxNumScore   int
	StartTime     time.Time
	EndTime       *time.Time
	Duration      time.Duration
	State         TournamentState
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewTournament creates a new tournament aggregate.
func NewTournament(
	id shared.TournamentID,
	title, description string,
	category int,
	sortOrder SortOrder,
	operator Operator,
	resetSchedule string,
	authoritative, joinRequired bool,
	maxSize, maxNumScore int,
	startTime time.Time,
	duration time.Duration,
	now time.Time,
) (*Tournament, error) {
	if err := id.Validate(); err != nil {
		return nil, err
	}
	if title == "" {
		return nil, errors.New("title is required")
	}
	if category < 0 {
		return nil, errors.New("category must be non-negative")
	}
	if maxSize < 0 {
		return nil, errors.New("max size must be non-negative")
	}
	if maxNumScore < 0 {
		return nil, errors.New("max num score must be non-negative")
	}
	if startTime.IsZero() {
		return nil, errors.New("start time is required")
	}
	if duration < 0 {
		return nil, errors.New("duration must be non-negative")
	}

	return &Tournament{
		ID:            id,
		Title:         title,
		Description:   description,
		Category:      category,
		SortOrder:     sortOrder,
		Operator:      operator,
		ResetSchedule: resetSchedule,
		Authoritative: authoritative,
		JoinRequired:  joinRequired,
		MaxSize:       maxSize,
		MaxNumScore:   maxNumScore,
		StartTime:     startTime,
		Duration:      duration,
		State:         StateActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// End marks the tournament as ended.
func (t *Tournament) End(endTime time.Time) error {
	if t.State == StateEnded {
		return ErrTournamentAlreadyEnded
	}
	if endTime.Before(t.StartTime) {
		return errors.New("end time cannot be before start time")
	}
	t.State = StateEnded
	t.EndTime = &endTime
	t.UpdatedAt = endTime
	return nil
}

// Reset marks the tournament as reset.
func (t *Tournament) Reset(resetTime time.Time) error {
	if t.State != StateActive && t.State != StateEnded {
		return errors.New("can only reset active or ended tournaments")
	}
	t.State = StateReset
	t.UpdatedAt = resetTime
	return nil
}

// IsActive checks if the tournament is currently active.
func (t *Tournament) IsActive() bool {
	return t.State == StateActive
}

// CalculateEndTime computes the end time based on start time and duration.
func (t *Tournament) CalculateEndTime() time.Time {
	if t.Duration > 0 {
		return t.StartTime.Add(t.Duration)
	}
	if t.EndTime != nil {
		return *t.EndTime
	}
	return time.Time{}
}

// Validate ensures the tournament is well-formed.
func (t *Tournament) Validate() error {
	if err := t.ID.Validate(); err != nil {
		return err
	}
	if t.Title == "" {
		return errors.New("title is required")
	}
	if t.StartTime.IsZero() {
		return errors.New("start time is required")
	}
	return nil
}
