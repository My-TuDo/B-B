// Package ws 提供基于 WebSocket 的弹幕实时通信功能。
// Hub 管理按 videoID 分组的房间，每个房间内的客户端通过 channel 收发弹幕消息。
package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

// Client 表示一个 WebSocket 连接的客户端。
// 每个客户端绑定到一个视频房间（VideoID）和一个用户（UserID），
// 通过 Hub 进行注册/注销和消息广播。
type Client struct {
	Conn    *websocket.Conn // WebSocket 连接
	Send    chan []byte     // 发送消息的缓冲 channel
	VideoID uint            // 所在的视频房间 ID
	UserID  uint            // 用户 ID
	Hub     *Hub            // 所属的 Hub
	mu      sync.Mutex      // 保护 Conn 的并发写
}

// ReadPump 是客户端的读取协程，持续从 WebSocket 连接读取消息。
// 收到弹幕 JSON 后通过 Hub 广播给同房间的所有客户端。
// 连接断开时自动注销并关闭连接。
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c // 通知 Hub 注销
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break // 连接断开或出错，退出循环
		}

		// 解析收到的弹幕 JSON 并广播到同一房间
		var danmaku map[string]interface{}
		if err := json.Unmarshal(message, &danmaku); err == nil {
			// 转发到同一房间
			c.Hub.Broadcast <- &BroadcastMessage{
				VideoID: c.VideoID,
				Data:    message,
			}
		}
	}
}

// WritePump 是客户端的写入协程，持续从 Send channel 取出消息并写入 WebSocket。
// Send channel 关闭时退出。
func (c *Client) WritePump() {
	defer c.Conn.Close()

	for msg := range c.Send {
		c.mu.Lock()
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		c.mu.Unlock()
		if err != nil {
			break // 写入失败，退出
		}
	}
}

// BroadcastMessage 是一条待广播的弹幕消息，包含目标房间和消息数据。
type BroadcastMessage struct {
	VideoID uint   // 目标视频房间
	Data    []byte // 消息内容（JSON 字节）
}

// Hub 是 WebSocket 连接管理中心。
// 维护按 VideoID 分组的房间映射，通过 channel 处理注册、注销和广播。
type Hub struct {
	Rooms      map[uint]map[*Client]bool // videoID → 客户端集合
	Broadcast  chan *BroadcastMessage    // 广播消息 channel
	Register   chan *Client              // 客户端注册 channel
	Unregister chan *Client              // 客户端注销 channel
	mu         sync.RWMutex              // 保护 Rooms 的并发访问
}

// DefaultHub 是全局唯一的 Hub 实例。
var DefaultHub *Hub

// InitHub 初始化全局 Hub 并启动其事件循环。
func InitHub() {
	DefaultHub = &Hub{
		Rooms:      make(map[uint]map[*Client]bool),
		Broadcast:  make(chan *BroadcastMessage, 256),
		Register:   make(chan *Client, 256),
		Unregister: make(chan *Client, 256),
	}
	go DefaultHub.Run() // 在后台 goroutine 中运行事件循环
}

// Run 是 Hub 的主事件循环，在单独的 goroutine 中运行。
// 通过 select 处理三种事件：客户端注册、客户端注销和消息广播。
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// 注册：将客户端加入对应房间
			h.mu.Lock()
			if h.Rooms[client.VideoID] == nil {
				h.Rooms[client.VideoID] = make(map[*Client]bool)
			}
			h.Rooms[client.VideoID][client] = true
			h.mu.Unlock()

		case client := <-h.Unregister:
			// 注销：从房间移除客户端，关闭 Send channel
			h.mu.Lock()
			if clients, ok := h.Rooms[client.VideoID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send) // 关闭 channel 以终止 WritePump
				}
				// 房间为空则清理
				if len(clients) == 0 {
					delete(h.Rooms, client.VideoID)
				}
			}
			h.mu.Unlock()

		case msg := <-h.Broadcast:
			// 广播：将消息发送给目标房间的所有客户端
			h.mu.RLock()
			if clients, ok := h.Rooms[msg.VideoID]; ok {
				for client := range clients {
					select {
					case client.Send <- msg.Data:
					default:
						// 客户端发送缓冲区满，视为断开，异步注销
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
