package sessions

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

// Session contains authentication state stored in memory.
type Session struct {
	Token     string
	Login     string
	IP        string
	CreatedAt time.Time
	ExpiresAt time.Time
}

const Duration = 24 * time.Hour

var (
	store = make(map[string]*Session)
	mu    sync.RWMutex
)

// Create generates a session tied to the client's IP address.
func Create(login, ip string) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}

	now := time.Now()
	session := &Session{
		Token:     token,
		Login:     login,
		IP:        ip,
		CreatedAt: now,
		ExpiresAt: now.Add(Duration),
	}

	mu.Lock()
	store[token] = session
	mu.Unlock()

	return token, nil
}

// Validate checks that a token exists, is not expired, and matches the IP address.
func Validate(token, clientIP string) (string, bool) {
	mu.RLock()
	session, exists := store[token]
	mu.RUnlock()
	if !exists {
		return "", false
	}

	if time.Now().After(session.ExpiresAt) {
		return "", false
	}
	if session.IP != clientIP {
		return "", false
	}
	return session.Login, true
}

// Delete removes a session token.
func Delete(token string) {
	mu.Lock()
	defer mu.Unlock()
	delete(store, token)
}

// StartCleanup periodically removes expired sessions.
func StartCleanup() {
	ticker := time.NewTicker(time.Hour)
	go func() {
		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for token, session := range store {
				if now.After(session.ExpiresAt) {
					delete(store, token)
				}
			}
			mu.Unlock()
		}
	}()
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
