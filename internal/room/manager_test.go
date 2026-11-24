package room

import (
	"testing"
	"time"
)

func newTestManager() (*Manager, *fakePusher) {
	p := newFakePusher()
	return NewManager(p, 6, 0, []string{"chat"}), p
}

func TestCreateAndList(t *testing.T) {
	m, _ := newTestManager()

	resp := m.Create("", "Samy", 1)
	if resp.Type != "ROOM" || resp.Status != "WAITING" || resp.ResumeToken == "" {
		t.Fatalf("create: expected ROOM WAITING with a token, got %+v", resp)
	}
	if resp.RoomName != "Salle de Samy" {
		t.Errorf("an empty name becomes %q, got %q", "Salle de Samy", resp.RoomName)
	}
	if len(resp.RoomID) != 6 {
		t.Errorf("expected a 6-character room id, got %q", resp.RoomID)
	}
	if named := m.Create("Duel du soir", "Mohamed", 2); named.RoomName != "Duel du soir" {
		t.Errorf("a given name is kept, got %q", named.RoomName)
	}

	rooms := m.List()
	if len(rooms) != 2 {
		t.Fatalf("expected 2 rooms in the list, got %d", len(rooms))
	}
}

func TestListWaitingFirst(t *testing.T) {
	m, _ := newTestManager()
	full := m.Create("Pleine", "Samy", 1)
	m.Create("Libre", "Mohamed", 2)

	r, _ := m.Get(full.RoomID)
	r.Join("Lina", 3) // "Pleine" is now PLAYING

	rooms := m.List()
	if len(rooms) != 2 || rooms[0].Status != "WAITING" || rooms[1].Status != "PLAYING" {
		t.Errorf("rooms someone can join come first, got %+v", rooms)
	}
}

func TestCreateIdsDiffer(t *testing.T) {
	m, _ := newTestManager()
	seen := make(map[string]bool)
	for range 50 {
		id := m.Create("", "Samy", 1).RoomID
		if seen[id] {
			t.Fatalf("room id %q given twice", id)
		}
		seen[id] = true
	}
}

func TestGetAndRemove(t *testing.T) {
	m, _ := newTestManager()
	id := m.Create("", "Samy", 1).RoomID

	if _, ok := m.Get(id); !ok {
		t.Fatal("a room just created must be found")
	}
	m.Remove(id)
	if _, ok := m.Get(id); ok {
		t.Error("a removed room must not be found")
	}
	if _, ok := m.Get("NOPE"); ok {
		t.Error("an unknown id must not be found")
	}
}

func TestRemoveExpiredRooms(t *testing.T) {
	m, _ := newTestManager()
	m.Create("", "Samy", 1)

	if n := m.RemoveExpired(time.Now(), time.Minute); n != 0 {
		t.Errorf("a room used just now must stay, removed %d", n)
	}
	if n := m.RemoveExpired(time.Now().Add(time.Hour), time.Minute); n != 1 {
		t.Errorf("a room idle for an hour must go, removed %d", n)
	}
	if len(m.List()) != 0 {
		t.Error("the expired room is still listed")
	}
}

func TestDisconnectedPushesToOpponent(t *testing.T) {
	m, p := newTestManager()
	id := m.Create("", "Samy", 1).RoomID
	r, _ := m.Get(id)
	r.Join("Mohamed", 2)
	before := p.count(2)

	m.Disconnected(1) // Samy's connection drops

	pushed, ok := p.last(2)
	if !ok || p.count(2) != before+1 {
		t.Fatal("Mohamed must be told that Samy is gone")
	}
	if pushed.Players[0].Connected || !pushed.Players[1].Connected {
		t.Errorf("expected Samy disconnected and Mohamed connected, got %+v", pushed.Players)
	}

	m.Disconnected(99) // a connection that isn't in any room
	if p.count(2) != before+1 {
		t.Error("a connection outside every room must push nothing")
	}
}

func TestDisconnectedWhileWaitingPushesNothing(t *testing.T) {
	m, p := newTestManager()
	m.Create("", "Samy", 1)

	m.Disconnected(1)
	if p.count(1) != 0 {
		t.Error("alone in the room: there is nobody to tell")
	}
}
