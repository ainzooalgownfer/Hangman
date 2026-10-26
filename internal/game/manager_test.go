package game

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestManagerAddGetRemove(t *testing.T) {
	m := NewSessionManager()

	if _, ok := m.Get("id1"); ok {
		t.Fatal("expected no session before Add")
	}

	s := NewGameSession("id1", "Alice", "chat", 6)
	m.Add(s)

	got, ok := m.Get("id1")
	if !ok {
		t.Fatal("expected session after Add")
	}
	if got != s {
		t.Error("Get must return the same pointer that was added")
	}

	m.Remove("id1")
	if _, ok := m.Get("id1"); ok {
		t.Error("expected no session after Remove")
	}
}

func TestManagerSharesSession(t *testing.T) {
	m := NewSessionManager()
	m.Add(NewGameSession("id1", "Alice", "chat", 6))

	first, _ := m.Get("id1")
	first.Guess("c")

	second, _ := m.Get("id1")
	if got := second.Snapshot().MaskedWord; got != "C _ _ _" {
		t.Errorf("guess on one Get must be visible on the next, got %q", got)
	}
}

func TestManagerConcurrentAdd(t *testing.T) {
	m := NewSessionManager()

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := fmt.Sprintf("id%d", i)
			m.Add(NewGameSession(id, "p", "chat", 6))
		}()
	}
	wg.Wait()

	for i := range 100 {
		if _, ok := m.Get(fmt.Sprintf("id%d", i)); !ok {
			t.Fatalf("session id%d missing after concurrent Add", i)
		}
	}
}

func TestRemoveExpired_RemoveIdle(t *testing.T) {
	m := NewSessionManager()
	s := NewGameSession("1", "med", "TEST", 10)

	m.Add(s)
	r := m.RemoveExpired(time.Now().Add(10*time.Minute), 5*time.Minute)

	_, ok := m.Get(s.id)
	if r != 1 || ok {
		t.Fatalf("expected 1 removed and session gone, got removed=%d found=%v", r, ok)

	}

}

func TestRemoveExpired_KeepRecents(t *testing.T) {
	m := NewSessionManager()
	s := NewGameSession("1", "med", "TEST", 10)

	m.Add(s)
	r := m.RemoveExpired(time.Now(), 5*time.Minute)

	_, ok := m.Get(s.id)
	if r != 0 || !ok {
		t.Fatalf("expected 0 removed and Session found, got removed=%d , found=%v", r, ok)
	}

}

func TestRemoveExpired_ActivityResetsTimer(t *testing.T) {
	m := NewSessionManager()
	s := NewGameSession("1", "med", "TEST", 10)

	m.Add(s)
	//simulate a old active session beacuse guess and the constructore both call now()time
	s.lastActive = time.Now().Add(-7 * time.Minute)
	s.Guess("T")

	r := m.RemoveExpired(time.Now().Add(3 * time.Minute), 5*time.Minute)

	_, ok := m.Get(s.id)
	if r != 0 || !ok {
		t.Fatalf("expected 0 and found cause Guess updates lastActive, got removed=%d, found=%v", r,ok)
	}
}
