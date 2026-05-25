package users

import "testing"

func TestPasswordHashAndAddIfAbsent(t *testing.T) {
	resetStore(t)

	hash, err := HashPassword("secret-password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if !AddIfAbsent("alice", hash) {
		t.Fatal("AddIfAbsent() = false, want true for a new login")
	}
	if !CheckCredentials("alice", "secret-password") {
		t.Fatal("CheckCredentials() = false for the original password")
	}
	if AddIfAbsent("alice", hash) {
		t.Fatal("AddIfAbsent() = true for an existing login")
	}
}

func TestInitDefaults(t *testing.T) {
	resetStore(t)

	if err := InitDefaults(); err != nil {
		t.Fatalf("InitDefaults() error = %v", err)
	}
	if !CheckCredentials("admin1", "admin1") || !CheckCredentials("admin2", "admin2") {
		t.Fatal("default users cannot log in with documented credentials")
	}
}

func resetStore(t *testing.T) {
	t.Helper()
	mu.Lock()
	previous := store
	store = make(map[string]User)
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		store = previous
		mu.Unlock()
	})
}
