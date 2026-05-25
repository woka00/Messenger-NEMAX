package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/woka00/Messenger-NEMAX/internal/sessions"
	"github.com/woka00/Messenger-NEMAX/internal/users"
	"github.com/woka00/Messenger-NEMAX/internal/ws"
)

func TestRegisterThenLoginWithPlainPassword(t *testing.T) {
	server := NewServer(ws.NewHub())
	login := fmt.Sprintf("user-%d", time.Now().UnixNano())
	body := []byte(fmt.Sprintf(`{"login":%q,"password":"secret"}`, login))

	registerReq := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	registerRes := httptest.NewRecorder()
	server.handleRegister(registerRes, registerReq)
	if registerRes.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d", registerRes.Code, http.StatusCreated)
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
	loginReq.RemoteAddr = "192.0.2.5:5000"
	loginRes := httptest.NewRecorder()
	server.handleLogin(loginRes, loginReq)
	if loginRes.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d", loginRes.Code, http.StatusOK)
	}
}

func TestRegisterDoesNotReplaceExistingUser(t *testing.T) {
	server := NewServer(ws.NewHub())
	login := fmt.Sprintf("same-user-%d", time.Now().UnixNano())

	first := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBufferString(
		fmt.Sprintf(`{"login":%q,"password":"first-password"}`, login),
	))
	firstRes := httptest.NewRecorder()
	server.handleRegister(firstRes, first)
	if firstRes.Code != http.StatusCreated {
		t.Fatalf("first registration status = %d, want %d", firstRes.Code, http.StatusCreated)
	}

	second := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBufferString(
		fmt.Sprintf(`{"login":%q,"password":"second-password"}`, login),
	))
	secondRes := httptest.NewRecorder()
	server.handleRegister(secondRes, second)
	if secondRes.Code != http.StatusConflict {
		t.Fatalf("second registration status = %d, want %d", secondRes.Code, http.StatusConflict)
	}
}

func TestGetClientIPIgnoresUntrustedProxyHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.8:4000"
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	req.Header.Set("X-Real-Ip", "203.0.113.2")

	if got := getClientIP(req); got != "192.0.2.8" {
		t.Fatalf("getClientIP() = %q, want RemoteAddr IP", got)
	}
}

func TestSendUsesAuthenticatedLoginAsSender(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	server := NewServer(hub)
	sender := fmt.Sprintf("sender-%d", time.Now().UnixNano())
	recipient := fmt.Sprintf("recipient-%d", time.Now().UnixNano())
	hash, err := users.HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if !users.AddIfAbsent(sender, hash) || !users.AddIfAbsent(recipient, hash) {
		t.Fatal("could not create test users")
	}

	token, err := sessions.Create(sender, "192.0.2.10")
	if err != nil {
		t.Fatalf("sessions.Create() error = %v", err)
	}

	sendReq := httptest.NewRequest(http.MethodPost, "/api/send", bytes.NewBufferString(
		fmt.Sprintf(`{"fromLogin":"forged","toLogin":%q,"text":"hello"}`, recipient),
	))
	sendReq.RemoteAddr = "192.0.2.10:4000"
	sendReq.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	sendRes := httptest.NewRecorder()
	server.handleSend(sendRes, sendReq)
	if sendRes.Code != http.StatusOK {
		t.Fatalf("send status = %d, want %d", sendRes.Code, http.StatusOK)
	}

	dialogReq := httptest.NewRequest(http.MethodPost, "/api/dialog", bytes.NewBufferString(
		fmt.Sprintf(`{"login":"forged","with":%q}`, recipient),
	))
	dialogReq.RemoteAddr = "192.0.2.10:4000"
	dialogReq.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	dialogRes := httptest.NewRecorder()
	server.handleDialog(dialogRes, dialogReq)
	if dialogRes.Code != http.StatusOK {
		t.Fatalf("dialog status = %d, want %d", dialogRes.Code, http.StatusOK)
	}

	var dialog []dialogMessage
	if err := json.NewDecoder(dialogRes.Body).Decode(&dialog); err != nil {
		t.Fatalf("decode dialog error = %v", err)
	}
	if len(dialog) != 1 || dialog[0].From != sender {
		t.Fatalf("dialog = %#v, want message from authenticated sender %q", dialog, sender)
	}
}
