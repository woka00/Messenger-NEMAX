package users

import (
	"crypto/sha256"
	"encoding/base64"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// хранит логин и хеш
type User struct {
	Login        string
	PasswordHash string
}

var (
	store = make(map[string]User)
	mu    sync.RWMutex
)

// кидаем админов в память
func InitDefaults() {
	mu.Lock()
	defer mu.Unlock()

	// без повторного засева если уже есть
	if len(store) > 0 {
		return
	}

	hash1, _ := HashPassword("admin1")
	hash2, _ := HashPassword("admin2")
	store["admin1"] = User{Login: "admin1", PasswordHash: hash1}
	store["admin2"] = User{Login: "admin2", PasswordHash: hash2}
}

// sha256+bcrypt
func HashPassword(password string) (string, error) {
	h := sha256.Sum256([]byte(password))
	sha := base64.StdEncoding.EncodeToString(h[:])
	hash, err := bcrypt.GenerateFromPassword([]byte(sha), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// сверяем пару логин/пароль
func CheckCredentials(login, password string) bool {
	mu.RLock()
	u, ok := store[login]
	mu.RUnlock()
	if !ok {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

// пишет юзера с готовым хешем
func Add(login, passwordHash string) {
	mu.Lock()
	defer mu.Unlock()
	store[login] = User{Login: login, PasswordHash: passwordHash}
}

// проверка на наличие логина
func Exists(login string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := store[login]
	return ok
}

// отдает логины
func List() []string {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]string, 0, len(store))
	for login := range store {
		result = append(result, login)
	}
	return result
}
