package room

import (
	"crypto/rand"
	"log/slog"
	"slices"
	"sync"
	"time"

	"hangman/internal/game/words"
	"hangman/internal/protocol"
)

// Manager holds every room, like SessionManager holds solo games.
type Manager struct {
	mu          sync.Mutex
	rooms       map[string]*Room
	pusher      Pusher
	maxAttempts int
	turnLimit   time.Duration
	words       []string
}

func NewManager(p Pusher, maxAttempts int, turnLimit time.Duration, words []string) *Manager {
	return &Manager{
		rooms:       make(map[string]*Room),
		pusher:      p,
		maxAttempts: maxAttempts,
		turnLimit:   turnLimit,
		words:       words,
	}
}

// Create opens a room and makes player its creator.
func (m *Manager) Create(roomName, player string, connID int64) protocol.CommandResponse {
	if roomName == "" {
		roomName = "Salle de " + player
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	var id string
	for {
		id = rand.Text()[:6]
		if _, taken := m.rooms[id]; !taken {
			break
		}
	}
	r := NewRoom(id, roomName, words.RandomWord(m.words),
		m.maxAttempts, m.turnLimit, m.pusher)

	resp := r.Join(player, connID)
	m.rooms[id] = r
	return resp
}

func (m *Manager) Get(id string) (*Room, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rooms[id]
	return r, ok
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	r, ok := m.rooms[id]
	delete(m.rooms, id)
	m.mu.Unlock()

	if ok {
		r.Stop() // after unlocking: never call into a room under the manager's lock
	}
}

// List returns the rooms, the ones someone can join first.
func (m *Manager) List() []protocol.RoomSummary {
	m.mu.Lock()
	rooms := make([]*Room, 0, len(m.rooms))
	for _, r := range m.rooms {
		rooms = append(rooms, r)

	}
	m.mu.Unlock()
	summaries := make([]protocol.RoomSummary, 0, len(rooms))
	for _, r := range rooms {
		summaries = append(summaries, r.Summary())
	}
	rank := func(status string) int {
		if status == "WAITING" {
			return 0 // rooms someone can join,come first
		}
		return 1
	}
	slices.SortStableFunc(summaries, func(a, b protocol.RoomSummary) int {
		return rank(a.Status) - rank(b.Status)
	})
	return summaries
}

// Disconnected tells every room that connID is gone. Called by the server
// when a connection closes (server.OnDisconnect).
func (m *Manager) Disconnected(connID int64) {
	m.mu.Lock()
	rooms := make([]*Room, 0, len(m.rooms))
	for _, r := range m.rooms {
		rooms = append(rooms, r)
	}
	m.mu.Unlock()

	for _, r := range rooms {
		r.Disconnected(connID)
	}
}

// RemoveExpired deletes the rooms nobody has touched for ttl and returns how many.
func (m *Manager) RemoveExpired(now time.Time, ttl time.Duration) int {
	m.mu.Lock()
	var expired []*Room
	for id, r := range m.rooms {
		if r.IdleSince(now) > ttl {
			delete(m.rooms, id)
			expired = append(expired, r)
		}
	}
	m.mu.Unlock()

	for _, r := range expired {
		r.Stop()
	}
	return len(expired)
}

// StartJanitor removes idle rooms every interval, like SessionManager.StartJanitor.
func (m *Manager) StartJanitor(ttl, interval time.Duration) (stop func()) {
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case now := <-ticker.C:
				if removed := m.RemoveExpired(now, ttl); removed != 0 {
					slog.Info("removed rooms", "count", removed)
				}
			case <-done:
				return
			}
		}
	}()
	return func() { close(done) }
}
