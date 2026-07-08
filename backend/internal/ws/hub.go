package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn    *websocket.Conn
	Send    chan []byte
	VideoID uint
	UserID  uint
	Hub     *Hub
	mu      sync.Mutex
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		// Parse incoming danmaku JSON and broadcast
		var danmaku map[string]interface{}
		if err := json.Unmarshal(message, &danmaku); err == nil {
			// Re-broadcast to same room
			c.Hub.Broadcast <- &BroadcastMessage{
				VideoID: c.VideoID,
				Data:    message,
			}
		}
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()

	for msg := range c.Send {
		c.mu.Lock()
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		c.mu.Unlock()
		if err != nil {
			break
		}
	}
}

type BroadcastMessage struct {
	VideoID uint
	Data    []byte
}

type Hub struct {
	Rooms      map[uint]map[*Client]bool
	Broadcast  chan *BroadcastMessage
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

var DefaultHub *Hub

func InitHub() {
	DefaultHub = &Hub{
		Rooms:      make(map[uint]map[*Client]bool),
		Broadcast:  make(chan *BroadcastMessage, 256),
		Register:   make(chan *Client, 256),
		Unregister: make(chan *Client, 256),
	}
	go DefaultHub.Run()
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.Rooms[client.VideoID] == nil {
				h.Rooms[client.VideoID] = make(map[*Client]bool)
			}
			h.Rooms[client.VideoID][client] = true
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.Rooms[client.VideoID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
				}
				if len(clients) == 0 {
					delete(h.Rooms, client.VideoID)
				}
			}
			h.mu.Unlock()

		case msg := <-h.Broadcast:
			h.mu.RLock()
			if clients, ok := h.Rooms[msg.VideoID]; ok {
				for client := range clients {
					select {
					case client.Send <- msg.Data:
					default:
						go func(c *Client) {
							h.Unregister <- c
						}(client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}
