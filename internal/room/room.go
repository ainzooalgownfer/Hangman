package room

import (
	"crypto/rand"
	"slices"
	"strings"
	"sync"
	"time"

	"hangman/internal/game/words"
	"hangman/internal/protocol"
)

// Pusher sends a message to a connection that didn't ask for it.
// The network Hub implements it; the room never touches a socket.
type Pusher interface {
	Push(connID int64, msg protocol.CommandResponse)
}

type player struct {
	name      string
	token     string // HMAC key for this player's signed requests
	connID    int64
	connected bool
	score     int
	lives     int
}

// Room is one 1 vs 1 duel. Everything below mu is protected by it.
type Room struct {
	mu          sync.Mutex
	id, name    string
	word        string // in capitals
	maxAttempts int
	turnLimit   time.Duration
	pusher      Pusher

	players []*player // 1 while WAITING, then 2
	status  string    // WAITING, PLAYING, FINISHED
	turn    int       // index of the player whose turn it is
	tried   []string  // base letters already played, by both players
	version int       // +1 on every change
	message string    // last event, shown to both players
}

func NewRoom(id, name, word string, maxAttempts int, turnLimit time.Duration, p Pusher) *Room {
	return &Room{id: id,
		name:        name,
		word:        strings.ToUpper(word),
		maxAttempts: maxAttempts,
		turnLimit:   turnLimit,
		pusher:      p,
		status:      "WAITING",
		tried:       make([]string, 0)}
}

func (r *Room) Token(i int) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if i < 0 || i >= len(r.players) {
		return ""
	}
	return r.players[i].token
}

func (r *Room) Summary() protocol.RoomSummary {
	r.mu.Lock()
	defer r.mu.Unlock()

	names := make([]string, 0, len(r.players))
	for _, p := range r.players {
		names = append(names, p.name)
	}
	return protocol.RoomSummary{
		RoomID:  r.id,
		Name:    r.name,
		Players: names,
		Status:  r.status,
	}
}

// view builds the ROOM message as player i sees it. Caller holds r.mu.
// view builds the ROOM message as player i sees it. Caller holds r.mu.
func (r *Room) view(i int) protocol.CommandResponse {
	players := make([]protocol.PlayerView, 0, len(r.players))
	for _, p := range r.players {
		players = append(players, protocol.PlayerView{
			Name:         p.name,
			Score:        p.score,
			AttemptsLeft: p.lives,
			Connected:    p.connected,
		})
	}

	resp := protocol.CommandResponse{
		Type:         "ROOM",
		RoomID:       r.id,
		RoomName:     r.name,
		Status:       r.status,
		Version:      r.version,
		Message:      r.message,
		MaskedWord:   words.Mask(r.word, r.tried),
		TriedLetters: slices.Clone(r.tried),
		Players:      players,
		You:          &i, // i is this call's own copy: safe to point to
	}
	if r.status == "PLAYING" {
		t := r.turn // a copy: never hand out a pointer into the room
		resp.Turn = &t
	}
	return resp
}

// Join adds a player. The first one waits; the second starts the duel, and the
// creator, who asked for nothing, gets the new state pushed.
func (r *Room) Join(name string, connID int64) protocol.CommandResponse {
	r.mu.Lock() // no defer: the push below must happen after unlocking

	if len(r.players) == 2 {
		r.mu.Unlock()
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "PLAYER_NOT_ALLOWED",
			Message: "Salle complète.",
		}
	}

	p := &player{
		name:      name,
		token:     rand.Text(),
		connID:    connID,
		connected: true,
		lives:     r.maxAttempts,
	}
	r.players = append(r.players, p)
	me := len(r.players) - 1
	r.version++

	starts := len(r.players) == 2
	if starts {
		r.status = "PLAYING"
		r.turn = 0 // the creator plays first
		r.message = p.name + " a rejoint la salle."
	}

	reply := r.view(me)
	reply.ResumeToken = p.token // only in this player's reply, never in a push

	// Built under the lock, so it matches the reply exactly...
	var toCreator protocol.CommandResponse
	var creatorConn int64
	if starts {
		toCreator = r.view(0)
		creatorConn = r.players[0].connID
	}
	r.mu.Unlock()

	// ...and sent after it, so a slow client never blocks the room.
	if starts {
		r.pusher.Push(creatorConn, toCreator)
	}
	return reply
}
