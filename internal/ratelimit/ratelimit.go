package ratelimit

import (
	"sync"
	"time"
)

type loginAttempts struct {
	FailedCount  int
	BlockedUntil time.Time
	LastAttempt  time.Time
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

// проверка бана
func CheckLoginBlocked(identifier string) (bool, time.Time) {
	mu.RLock()
	defer mu.RUnlock()

	entry, exists := attempts[identifier]
	if !exists {
		return false, time.Time{}
	}

	if entry.BlockedUntil.After(time.Now()) {
		return true, entry.BlockedUntil
	}

	if entry.BlockedUntil.Before(time.Now()) && entry.BlockedUntil != (time.Time{}) {
		// снимаем блок и обнуляем
		mu.RUnlock()
		mu.Lock()
		entry.FailedCount = 0
		entry.BlockedUntil = time.Time{}
		mu.Unlock()
		mu.RLock()
	}

	return false, time.Time{}
}

// лимит в окне
func CheckRateLimit(identifier string) bool {
	mu.Lock()
	defer mu.Unlock()

	entry, exists := attempts[identifier]
	now := time.Now()

	if !exists {
		attempts[identifier] = &loginAttempts{LastAttempt: now}
		return true
	}

	if now.Sub(entry.LastAttempt) > window {
		entry.LastAttempt = now
		return true
	}

	if entry.FailedCount >= maxAttemptsPerWindow {
		return false
	}

	entry.LastAttempt = now
	return true
}

// фиксирует фейл и можем блокнуть быдло
func RecordFailed(identifier string) {
	mu.Lock()
	defer mu.Unlock()

	entry, exists := attempts[identifier]
	if !exists {
		entry = &loginAttempts{}
		attempts[identifier] = entry
	}

	entry.FailedCount++
	entry.LastAttempt = time.Now()
	if entry.FailedCount >= maxFailedAttempts {
		entry.BlockedUntil = time.Now().Add(blockDuration)
	}
}

// RecordSuccess чистим счетчики
func RecordSuccess(identifier string) {
	mu.Lock()
	defer mu.Unlock()
	if entry, exists := attempts[identifier]; exists {
		entry.FailedCount = 0
		entry.BlockedUntil = time.Time{}
	}
}
