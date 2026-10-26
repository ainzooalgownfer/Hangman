package game

import (
	"log/slog"
	"sync"
	"time"
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

func (m *SessionManager) RemoveExpired(now time.Time, ttl time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	removed := 0
	for id, s := range m.sessions {
		if s.IdleSince(now) > ttl {
			delete(m.sessions, id)
			removed++
		}
	}
	return removed
}

func (m *SessionManager) StartJanitor(ttl, interval time.Duration) (stop func()) {
	// chan is a typed pipe that go routines uses it to pass data or read safely
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case now := <-ticker.C:
				removed := m.RemoveExpired(now, ttl)
				if removed != 0 {
					slog.Info("removed Sessions", "count", removed)
				}
			case <-done:
				return
			}
		}
	}()

	return func() { close(done) }
}
