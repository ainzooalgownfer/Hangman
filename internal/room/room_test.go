package room

import (
	"sync"
	"testing"

	"hangman/internal/protocol"
)

// fakePusher records what a room pushes, instead of writing to sockets.
type fakePusher struct {
	mu   sync.Mutex
	sent map[int64][]protocol.CommandResponse
}

func newFakePusher() *fakePusher {
	return &fakePusher{sent: make(map[int64][]protocol.CommandResponse)}
}

func (f *fakePusher) Push(connID int64, msg protocol.CommandResponse) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent[connID] = append(f.sent[connID], msg)
}

// last returns the last message pushed to connID.
func (f *fakePusher) last(connID int64) (protocol.CommandResponse, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	msgs := f.sent[connID]
	if len(msgs) == 0 {
		return protocol.CommandResponse{}, false
	}
	return msgs[len(msgs)-1], true
}

// count returns how many messages were pushed to connID.
func (f *fakePusher) count(connID int64) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sent[connID])
}

// intVal reads a *int field (You, Turn), failing the test if it is missing.
func intVal(t *testing.T, name string, p *int) int {
	t.Helper()
	if p == nil {
		t.Fatalf("%s missing", name)
	}
	return *p
}

func TestJoinStartsDuel(t *testing.T) {
	p := newFakePusher()
	r := NewRoom("K3J9", "Salle de Samy", "chat", 6, 0, p)

	first := r.Join("Samy", 1)
	if first.Type != "ROOM" || first.Status != "WAITING" {
		t.Fatalf("first join: expected ROOM WAITING, got %+v", first)
	}
	if first.RoomID != "K3J9" || first.RoomName != "Salle de Samy" {
		t.Errorf("first join: wrong room id or name: %+v", first)
	}
	if intVal(t, "you", first.You) != 0 || first.ResumeToken == "" {
		t.Errorf("first join: expected you=0 and a resume_token, got %+v", first)
	}
	if first.Turn != nil {
		t.Errorf("first join: no turn while waiting, got %d", *first.Turn)
	}
	if p.count(1) != 0 {
		t.Errorf("first join: nothing to push yet, got %d messages", p.count(1))
	}

	second := r.Join("Mohamed", 2)
	if second.Status != "PLAYING" || len(second.Players) != 2 {
		t.Fatalf("second join: expected PLAYING with 2 players, got %+v", second)
	}
	if intVal(t, "you", second.You) != 1 || intVal(t, "turn", second.Turn) != 0 {
		t.Errorf("second join: expected you=1 and turn=0 (the creator starts), got %+v", second)
	}
	if second.MaskedWord != "_ _ _ _" {
		t.Errorf("second join: expected masked word %q, got %q", "_ _ _ _", second.MaskedWord)
	}
	if second.Players[0].Name != "Samy" || second.Players[1].Name != "Mohamed" || second.Players[1].AttemptsLeft != 6 {
		t.Errorf("second join: wrong players: %+v", second.Players)
	}

	// The creator didn't ask for anything, but must learn the duel started.
	pushed, ok := p.last(1)
	if !ok {
		t.Fatal("the creator got no push when the opponent joined")
	}
	if pushed.Type != "ROOM" || pushed.Status != "PLAYING" || intVal(t, "you", pushed.You) != 0 {
		t.Errorf("push to the creator: expected ROOM PLAYING with you=0, got %+v", pushed)
	}
	if pushed.RequestID != 0 || pushed.ResumeToken != "" {
		t.Errorf("a push carries no request_id and never a token, got %+v", pushed)
	}
	if pushed.Version != second.Version {
		t.Errorf("both players must see the same version, got %d and %d", pushed.Version, second.Version)
	}
}

func TestJoinFullRoom(t *testing.T) {
	p := newFakePusher()
	r := NewRoom("K3J9", "Salle de Samy", "chat", 6, 0, p)
	r.Join("Samy", 1)
	r.Join("Mohamed", 2)
	pushesBefore := p.count(1) + p.count(2)

	third := r.Join("Intrus", 3)
	if third.Type != "ERROR" || third.Code != "PLAYER_NOT_ALLOWED" {
		t.Errorf("third join: expected PLAYER_NOT_ALLOWED, got %+v", third)
	}
	if s := r.Summary(); len(s.Players) != 2 {
		t.Errorf("third join must not change the room, players now %v", s.Players)
	}
	if p.count(1)+p.count(2) != pushesBefore || p.count(3) != 0 {
		t.Error("a refused join must push nothing")
	}
}

func TestTokensDiffer(t *testing.T) {
	r := NewRoom("K3J9", "Salle de Samy", "chat", 6, 0, newFakePusher())
	a := r.Join("Samy", 1).ResumeToken
	b := r.Join("Mohamed", 2).ResumeToken

	if a == "" || b == "" || a == b {
		t.Errorf("each player needs their own token, got %q and %q", a, b)
	}
	if r.Token(0) != a || r.Token(1) != b {
		t.Errorf("Token(i) must return the token given to player i")
	}
	if r.Token(2) != "" || r.Token(-1) != "" {
		t.Errorf("Token of a player who doesn't exist must be empty")
	}
}

func TestSummary(t *testing.T) {
	r := NewRoom("K3J9", "Salle de Samy", "chat", 6, 0, newFakePusher())
	r.Join("Samy", 1)

	s := r.Summary()
	if s.RoomID != "K3J9" || s.Name != "Salle de Samy" || s.Status != "WAITING" ||
		len(s.Players) != 1 || s.Players[0] != "Samy" {
		t.Errorf("unexpected summary: %+v", s)
	}
}
