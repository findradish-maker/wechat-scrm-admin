package realtime

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*websocket.Conn]struct{}),
	}
}

func (h *Hub) Register(wxid string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[wxid]; !ok {
		h.clients[wxid] = make(map[*websocket.Conn]struct{})
	}
	h.clients[wxid][conn] = struct{}{}
}

func (h *Hub) Unregister(wxid string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	group, ok := h.clients[wxid]
	if !ok {
		return
	}
	delete(group, conn)
	if len(group) == 0 {
		delete(h.clients, wxid)
	}
}

func (h *Hub) Broadcast(wxid string, payload []byte) {
	h.mu.RLock()
	group := h.clients[wxid]
	if len(group) == 0 {
		h.mu.RUnlock()
		return
	}

	conns := make([]*websocket.Conn, 0, len(group))
	for conn := range group {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()

	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			_ = conn.Close()
			h.Unregister(wxid, conn)
		}
	}
}

func (h *Hub) Count(wxid string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.clients[wxid])
}
