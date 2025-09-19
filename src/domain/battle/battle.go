package battle

import (
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// MatchState captures the authoritative state serialized to Nakama match handler.
type MatchState struct {
	Tick      int64
	Payload   []byte
	UpdatedAt time.Time
}

// PlayerSlot tracks participant placement in a battle.
type PlayerSlot struct {
	PlayerID shared.PlayerID
	JoinedAt time.Time
	Ready    bool
}

// Battle aggregate orchestrates match lifecycle around Nakama matches.
type Battle struct {
	ID             shared.BattleID
	Leader         shared.PlayerID
	Slots          []PlayerSlot
	StateSnapshot  MatchState
	CreatedAt      time.Time
	UpdatedAt      time.Time
	IdempotencyKey shared.IdempotencyKey
}

func NewBattle(id shared.BattleID, leader shared.PlayerID, key shared.IdempotencyKey, now time.Time) (*Battle, error) {
	if err := id.Validate(); err != nil {
		return nil, err
	}
	if err := leader.Validate(); err != nil {
		return nil, err
	}
	if err := key.Validate(); err != nil {
		return nil, err
	}
	return &Battle{
		ID:             id,
		Leader:         leader,
		Slots:          []PlayerSlot{{PlayerID: leader, JoinedAt: now, Ready: true}},
		CreatedAt:      now,
		UpdatedAt:      now,
		IdempotencyKey: key,
	}, nil
}

func (b *Battle) AddPlayer(player shared.PlayerID, now time.Time) error {
	for _, slot := range b.Slots {
		if slot.PlayerID == player {
			return ErrPlayerAlreadyJoined
		}
	}
	b.Slots = append(b.Slots, PlayerSlot{PlayerID: player, JoinedAt: now})
	b.UpdatedAt = now
	return nil
}

func (b *Battle) MarkReady(player shared.PlayerID, ready bool, now time.Time) error {
	for i, slot := range b.Slots {
		if slot.PlayerID == player {
			slot.Ready = ready
			b.Slots[i] = slot
			b.UpdatedAt = now
			return nil
		}
	}
	return ErrPlayerNotFound
}

func (b *Battle) UpdateSnapshot(state MatchState) {
	if state.UpdatedAt.IsZero() {
		state.UpdatedAt = time.Now().UTC()
	}
	b.StateSnapshot = state
	b.UpdatedAt = state.UpdatedAt
}
