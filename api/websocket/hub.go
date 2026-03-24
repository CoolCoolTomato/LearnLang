package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte
}

type Hub struct {
	clients    map[int64][]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	mu         sync.RWMutex
}

type Message struct {
	UserID  int64
	Content []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64][]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = append(h.clients[client.UserID], client)
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				for i, c := range clients {
					if c == client {
						h.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
						close(client.Send)
						break
					}
				}
				if len(h.clients[client.UserID]) == 0 {
					delete(h.clients, client.UserID)
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			clients, ok := h.clients[message.UserID]
			h.mu.RUnlock()

			if ok {
				for _, client := range clients {
					select {
					case client.Send <- message.Content:
					default:

					}
				}
			}
		}
	}
}

func (h *Hub) SendToUser(userID int64, message []byte) {
	h.broadcast <- &Message{
		UserID:  userID,
		Content: message,
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}
