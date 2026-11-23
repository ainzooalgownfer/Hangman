package network

import (
	"log/slog"
	"sync"

	"hangman/internal/protocol"
)

// Hub knows every open connection by its id, so the server can send a message
// to a player who didn't ask for one (the opponent's move).
type Hub struct {
	mu    sync.Mutex
	conns map[int64]*ClientConnection
}

func NewHub() *Hub {
	
	return &Hub{conns: make(map[int64]*ClientConnection)}
}

func (h *Hub) Add(connID int64, c *ClientConnection) {

	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[connID] = c
}

func (h *Hub) Remove(connID int64) {

	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, connID)
}

func (h *Hub) GetConn(connID int64) (*ClientConnection, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	v, ok := h.conns[connID]
	return v, ok
}

// Push sends msg to the connection connID, if it is still open.
func (h *Hub) Push(connID int64, msg protocol.CommandResponse) {

	conn, ok := h.GetConn(connID)
	if !ok {
		return // never added, or already gone
	}
	err := conn.SendResponse(msg)
	if err != nil {
		slog.Warn("push failed", "connID", connID, "err", err)
	}
}
