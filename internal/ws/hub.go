package ws

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

// Upgrader uses gorilla/websocket's same-origin validation.
var Upgrader = websocket.Upgrader{}

type Client struct {
	conn  *websocket.Conn
	login string
	send  chan []byte
	hub   *Hub
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

// Message is a notification sent to browser clients.
type Message struct {
	Type     string          `json:"type"`
	From     string          `json:"from,omitempty"`
	To       string          `json:"to,omitempty"`
	Text     string          `json:"text,omitempty"`
	Time     int64           `json:"time,omitempty"`
	Dialog   string          `json:"dialog,omitempty"`
	Messages json.RawMessage `json:"messages,omitempty"`
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			var payload Message
			_ = json.Unmarshal(message, &payload)
			for client := range h.clients {
				if payload.Type == "new_message" {
					if client.login != payload.From && client.login != payload.To {
						continue
					}
				}
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// Broadcast sends a payload through the hub.
func (h *Hub) Broadcast(msg []byte) {
	h.broadcast <- msg
}

// Register attaches an authenticated WebSocket connection to the hub.
func (h *Hub) Register(conn *websocket.Conn, login string) {
	client := &Client{
		conn:  conn,
		login: login,
		send:  make(chan []byte, 256),
		hub:   h,
	}
	h.register <- client
	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}

}
