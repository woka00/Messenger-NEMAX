package ratelimit

import (
	"testing"
	"time"
)

func TestCheckRateLimitCountsLoginRequests(t *testing.T) {
	resetAttempts(t)

	for i := 0; i < maxAttemptsPerWindow; i++ {
		if !CheckRateLimit("client") {
			t.Fatalf("CheckRateLimit() rejected request %d", i+1)
		}
	}
	if CheckRateLimit("client") {
		t.Fatal("CheckRateLimit() accepted a request above the per-window limit")
	}

	mu.Lock()
	attempts["client"].WindowStarted = time.Now().Add(-window)
	mu.Unlock()
	if !CheckRateLimit("client") {
		t.Fatal("CheckRateLimit() did not reset after the window expired")
	}
}

func TestCheckLoginBlockedExpires(t *testing.T) {
	resetAttempts(t)

	for i := 0; i < maxFailedAttempts; i++ {
		RecordFailed("client")
	}
	if blocked, _ := CheckLoginBlocked("client"); !blocked {
		t.Fatal("CheckLoginBlocked() = false after too many failures")
	}

	mu.Lock()
	attempts["client"].BlockedUntil = time.Now().Add(-time.Second)
	mu.Unlock()
	if blocked, _ := CheckLoginBlocked("client"); blocked {
		t.Fatal("CheckLoginBlocked() = true after blocking expired")
	}
}

func resetAttempts(t *testing.T) {
	t.Helper()
	mu.Lock()
	previous := attempts
	attempts = make(map[string]*loginAttempts)
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		attempts = previous
		mu.Unlock()
	})
}
