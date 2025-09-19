package player

import (
	"time"

	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

// DeviceFingerprint captures device claims trusted during authentication.
type DeviceFingerprint struct {
	ID       string
	Platform string
	LastSeen time.Time
}

type SessionMetadata struct {
	SessionID string
	IpAddress string
	UserAgent string
	IssuedAt  time.Time
}

// PlayerAccount is the aggregate root for authentication and session policies.
type PlayerAccount struct {
	ID            shared.PlayerID
	Email         string
	DisplayName   string
	Devices       map[string]DeviceFingerprint
	Sessions      []SessionMetadata
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Suspended     bool
	SuspensionMsg string
}

func NewPlayerAccount(id shared.PlayerID, email, displayName string, now time.Time) (*PlayerAccount, error) {
	if err := id.Validate(); err != nil {
		return nil, err
	}
	if email == "" {
		return nil, ErrEmailRequired
	}
	acct := &PlayerAccount{
		ID:          id,
		Email:       email,
		DisplayName: displayName,
		Devices:     make(map[string]DeviceFingerprint),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return acct, nil
}

func (p *PlayerAccount) RegisterDevice(device DeviceFingerprint) error {
	if p.Suspended {
		return ErrAccountSuspended
	}
	if device.ID == "" {
		return ErrDeviceInvalid
	}
	if device.LastSeen.IsZero() {
		device.LastSeen = time.Now().UTC()
	}
	p.Devices[device.ID] = device
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *PlayerAccount) RecordSession(session SessionMetadata) {
	if session.SessionID == "" {
		return
	}
	if session.IssuedAt.IsZero() {
		session.IssuedAt = time.Now().UTC()
	}
	p.Sessions = append(p.Sessions, session)
	p.UpdatedAt = time.Now().UTC()
}

func (p *PlayerAccount) Suspend(message string) {
	p.Suspended = true
	p.SuspensionMsg = message
	p.UpdatedAt = time.Now().UTC()
}

func (p *PlayerAccount) Reinstate() {
	p.Suspended = false
	p.SuspensionMsg = ""
	p.UpdatedAt = time.Now().UTC()
}

func (p *PlayerAccount) CanStartBattle(key shared.IdempotencyKey) error {
	if p.Suspended {
		return ErrAccountSuspended
	}
	if err := key.Validate(); err != nil {
		return err
	}
	return nil
}
