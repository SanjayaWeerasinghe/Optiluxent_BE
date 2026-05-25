package security

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	lockoutKeyPrefix = "lockout:"
	lockoutTTL       = 24 * time.Hour
)

type lockoutState struct {
	Attempts    int       `json:"attempts"`
	LockedUntil time.Time `json:"locked_until,omitempty"`
}

// AccountLockout tracks failed login attempts and locks accounts.
type AccountLockout struct {
	client *redis.Client
}

func NewAccountLockout(client *redis.Client) *AccountLockout {
	return &AccountLockout{client: client}
}

// IsLocked returns true if the email is currently locked.
func (l *AccountLockout) IsLocked(ctx context.Context, email string) (bool, error) {
	state, err := l.getState(ctx, email)
	if err != nil {
		return false, err
	}
	if state == nil {
		return false, nil
	}
	return !state.LockedUntil.IsZero() && time.Now().Before(state.LockedUntil), nil
}

// RecordFailure increments failed attempts and applies lockout if thresholds are crossed.
func (l *AccountLockout) RecordFailure(ctx context.Context, email string) error {
	state, err := l.getState(ctx, email)
	if err != nil {
		return err
	}
	if state == nil {
		state = &lockoutState{}
	}

	state.Attempts++

	switch {
	case state.Attempts >= 20:
		state.LockedUntil = time.Now().Add(24 * time.Hour)
	case state.Attempts >= 10:
		state.LockedUntil = time.Now().Add(time.Hour)
	case state.Attempts >= 5:
		state.LockedUntil = time.Now().Add(15 * time.Minute)
	}

	return l.setState(ctx, email, state)
}

// Reset clears the lockout state on successful login.
func (l *AccountLockout) Reset(ctx context.Context, email string) error {
	return l.client.Del(ctx, lockoutKeyPrefix+email).Err()
}

// Attempts returns the current failure count.
func (l *AccountLockout) Attempts(ctx context.Context, email string) (int, error) {
	state, err := l.getState(ctx, email)
	if err != nil {
		return 0, err
	}
	if state == nil {
		return 0, nil
	}
	return state.Attempts, nil
}

func (l *AccountLockout) getState(ctx context.Context, email string) (*lockoutState, error) {
	data, err := l.client.Get(ctx, lockoutKeyPrefix+email).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("lockout: get state: %w", err)
	}
	var state lockoutState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("lockout: unmarshal: %w", err)
	}
	return &state, nil
}

func (l *AccountLockout) setState(ctx context.Context, email string, state *lockoutState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("lockout: marshal: %w", err)
	}
	return l.client.Set(ctx, lockoutKeyPrefix+email, data, lockoutTTL).Err()
}
