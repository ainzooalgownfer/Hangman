package game

import (
	"sync"
)

type SessionManager struct {
	mu       sync.Mutex
	sessions map[string]*GameSession
}

func NewSessionManager() *SessionManager {
	return &SessionManager{sessions: make(map[string]*GameSession)}
}

func (m *SessionManager) Add(g *GameSession) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[g.ID()] = g
}

func (m *SessionManager) Get(id string) (*GameSession, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, ok := m.sessions[id]
	return v, ok
}

func (m *SessionManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, id)
}
