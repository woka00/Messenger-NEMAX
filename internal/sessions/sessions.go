package sessions

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

// держит состояние авторизации
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

// выдает новый токен для логина+ip
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

// сверяет токен и IP
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

// убирает сессию по токену
func Delete(token string) {
	mu.Lock()
	defer mu.Unlock()
	delete(store, token)
}

// чистит протухшие сессии по таймеру
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
