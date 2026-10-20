package game

import (
	"fmt"
	"sync"
	"testing"
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
