package ratelimit

import (
	"sync"
	"time"
)

type loginAttempts struct {
	FailedCount      int
	AttemptsInWindow int
	BlockedUntil     time.Time
	WindowStarted    time.Time
}

var (
	attempts = make(map[string]*loginAttempts)
	mu       sync.RWMutex
)

const (
	maxFailedAttempts    = 5
	blockDuration        = 15 * time.Minute
	window               = time.Minute
	maxAttemptsPerWindow = 10
)

// CheckLoginBlocked reports whether login attempts are temporarily blocked.
func CheckLoginBlocked(identifier string) (bool, time.Time) {
	mu.Lock()
	defer mu.Unlock()

	entry, exists := attempts[identifier]
	if !exists {
		return false, time.Time{}
	}

	now := time.Now()
	if entry.BlockedUntil.After(now) {
		return true, entry.BlockedUntil
	}

	if !entry.BlockedUntil.IsZero() {
		entry.FailedCount = 0
		entry.BlockedUntil = time.Time{}
	}

	return false, time.Time{}
}

// CheckRateLimit allows no more than a fixed number of login requests per window.
func CheckRateLimit(identifier string) bool {
	mu.Lock()
	defer mu.Unlock()

	entry, exists := attempts[identifier]
	now := time.Now()

	if !exists {
		attempts[identifier] = &loginAttempts{
			AttemptsInWindow: 1,
			WindowStarted:    now,
		}
		return true
	}

	if now.Sub(entry.WindowStarted) >= window {
		entry.AttemptsInWindow = 0
		entry.WindowStarted = now
	}

	if entry.AttemptsInWindow >= maxAttemptsPerWindow {
		return false
	}

	entry.AttemptsInWindow++
	return true
}

// RecordFailed records an invalid login and may temporarily block it.
func RecordFailed(identifier string) {
	mu.Lock()
	defer mu.Unlock()

	entry, exists := attempts[identifier]
	if !exists {
		entry = &loginAttempts{}
		attempts[identifier] = entry
	}

	entry.FailedCount++
	if entry.FailedCount >= maxFailedAttempts {
		entry.BlockedUntil = time.Now().Add(blockDuration)
	}
}

// RecordSuccess clears failed-login blocking state.
func RecordSuccess(identifier string) {
	mu.Lock()
	defer mu.Unlock()
	if entry, exists := attempts[identifier]; exists {
		entry.FailedCount = 0
		entry.BlockedUntil = time.Time{}
	}
}
