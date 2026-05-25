package users

import (
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// User contains credentials stored in memory.
type User struct {
	Login        string
	PasswordHash string
}

var (
	store = make(map[string]User)
	mu    sync.RWMutex
)

// InitDefaults initializes demo accounts used for local testing.
func InitDefaults() error {
	mu.Lock()
	defer mu.Unlock()

	if len(store) > 0 {
		return nil
	}

	hash1, err := HashPassword("admin1")
	if err != nil {
		return err
	}
	hash2, err := HashPassword("admin2")
	if err != nil {
		return err
	}
	store["admin1"] = User{Login: "admin1", PasswordHash: hash1}
	store["admin2"] = User{Login: "admin2", PasswordHash: hash2}
	return nil
}

// HashPassword hashes the plain-text password for storage.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckCredentials verifies a login and plain-text password.
func CheckCredentials(login, password string) bool {
	mu.RLock()
	u, ok := store[login]
	mu.RUnlock()
	if !ok {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

// AddIfAbsent creates a user only if the login is not already registered.
func AddIfAbsent(login, passwordHash string) bool {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := store[login]; exists {
		return false
	}
	store[login] = User{Login: login, PasswordHash: passwordHash}
	return true
}

// Exists reports whether the login is registered.
func Exists(login string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := store[login]
	return ok
}

// List returns all registered logins.
func List() []string {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]string, 0, len(store))
	for login := range store {
		result = append(result, login)
	}
	return result
}
