package httpapi

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sort"
	"time"

	"github.com/woka00/Messenger-NEMAX/internal/messages"
	"github.com/woka00/Messenger-NEMAX/internal/ratelimit"
	"github.com/woka00/Messenger-NEMAX/internal/sessions"
	"github.com/woka00/Messenger-NEMAX/internal/users"
	"github.com/woka00/Messenger-NEMAX/internal/ws"
)

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type sendRequest struct {
	ToLogin string `json:"toLogin"`
	Text    string `json:"text"`
}

type inboxMessage struct {
	From string `json:"from"`
	Text string `json:"text"`
}

type dialogRequest struct {
	With string `json:"with"`
}

type dialogMessage struct {
	From string `json:"from"`
	Text string `json:"text"`
	Time int64  `json:"time"`
}

type usersResponse struct {
	Users []string `json:"users"`
}

type Server struct {
	hub *ws.Hub
}

func NewServer(hub *ws.Hub) *Server {
	return &Server{hub: hub}
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/register", s.handleRegisterPage)
	mux.HandleFunc("/app", s.handleAppPage)
	mux.HandleFunc("/api/login", s.handleLogin)
	mux.HandleFunc("/api/logout", s.handleLogout)
	mux.HandleFunc("/api/register", s.handleRegister)
	mux.HandleFunc("/api/users", s.handleUsers)
	mux.HandleFunc("/api/send", s.handleSend)
	mux.HandleFunc("/api/inbox", s.handleInbox)
	mux.HandleFunc("/api/dialog", s.handleDialog)
	mux.HandleFunc("/api/me", s.handleMe)
	mux.HandleFunc("/ws", s.handleWebSocket)

	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(loginHTML))
}

func (s *Server) handleRegisterPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(registerHTML))
}

func (s *Server) handleAppPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(appHTML))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clientIP := getClientIP(r)

	if !ratelimit.CheckRateLimit(clientIP) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Превышен лимит попыток. Попробуйте позже.",
		})
		return
	}

	if blocked, until := ratelimit.CheckLoginBlocked(clientIP); blocked {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		remaining := time.Until(until).Round(time.Second)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("IP временно заблокирован. Попробуйте через %v", remaining),
		})
		return
	}

	if !users.CheckCredentials(req.Login, req.Password) {
		ratelimit.RecordFailed(clientIP)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ratelimit.RecordSuccess(clientIP)

	token, err := sessions.Create(req.Login, clientIP)
	if err != nil {
		http.Error(w, "Не удалось создать сессию", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(sessions.Duration.Seconds()),
	})

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"login": req.Login})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err == nil {
		sessions.Delete(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	login, ok := s.getSessionFromRequest(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]string{"login": login})
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	login, ok := s.getSessionFromRequest(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	conn, err := ws.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.hub.Register(conn, login)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(req.Login) < 3 || len(req.Password) < 3 {
		http.Error(w, "Логин и пароль должны быть не короче 3 символов", http.StatusBadRequest)
		return
	}

	passwordHash, err := users.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Не удалось захешировать пароль", http.StatusInternalServerError)
		return
	}
	if !users.AddIfAbsent(req.Login, passwordHash) {
		http.Error(w, "Пользователь уже существует", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if _, ok := s.getSessionFromRequest(r); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	list := users.List()
	sort.Strings(list)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(usersResponse{Users: list})
}

func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	login, ok := s.getSessionFromRequest(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !users.Exists(req.ToLogin) {
		http.Error(w, "unknown recipient", http.StatusBadRequest)
		return
	}

	msg := messages.Message{
		From: login,
		To:   req.ToLogin,
		Text: req.Text,
		Time: time.Now().UnixNano(),
	}
	messages.Add(req.ToLogin, msg)

	wsMsg := ws.Message{
		Type: "new_message",
		From: login,
		To:   req.ToLogin,
		Text: req.Text,
		Time: msg.Time,
	}
	if data, err := json.Marshal(wsMsg); err == nil {
		s.hub.Broadcast(data)
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleInbox(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	login, ok := s.getSessionFromRequest(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	msgs := messages.Inbox(login)
	result := make([]inboxMessage, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, inboxMessage{From: m.From, Text: m.Text})
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(result)
}

func (s *Server) handleDialog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req dialogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	login, ok := s.getSessionFromRequest(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !users.Exists(req.With) {
		http.Error(w, "unknown peer", http.StatusBadRequest)
		return
	}

	all := messages.Dialog(login, req.With)
	result := make([]dialogMessage, 0, len(all))
	for _, m := range all {
		result = append(result, dialogMessage{
			From: m.From,
			Text: m.Text,
			Time: m.Time,
		})
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(result)
}

func (s *Server) getSessionFromRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return "", false
	}
	clientIP := getClientIP(r)
	return sessions.Validate(cookie.Value, clientIP)
}

func getClientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
