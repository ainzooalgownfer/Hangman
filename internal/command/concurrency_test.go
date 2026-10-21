package command_test

import (
	"slices"
	"sync"
	"testing"

	"hangman/internal/protocol"
)

// Run with the race detector to catch unsynchronized access:
//   CGO_ENABLED=1 go test -race ./internal/command/ -run Concurrent -v

// TestConcurrentPlayers runs many independent games at once through one
// dispatcher, so every goroutine shares the same session store.
func TestConcurrentPlayers(t *testing.T) {
	d := setupDispatcher()
	const players = 100

	var wg sync.WaitGroup
	results := make(chan string, players)

	for range players {
		wg.Add(1)
		go func() {
			defer wg.Done()

			newGame := d.Dispatch(protocol.CommandRequest{Action: "NEW_GAME", Player: "bot"})
			id := newGame.SessionID
			if id == "" {
				results <- "no session id"
				return
			}

			for _, l := range []string{"e", "a", "r", "o", "i", "t", "n", "s", "l", "c", "u", "d", "p", "m", "h", "g", "b", "f", "y", "w", "k", "v", "x", "z", "j", "q"} {
				resp := d.Dispatch(protocol.CommandRequest{Action: "LETTER", SessionID: id, Letter: l})
				if resp.Type == "ERROR" {
					results <- "unexpected error: " + resp.Code
					return
				}
				if resp.Status == "WON" || resp.Status == "LOST" {
					break
				}
			}

			quit := d.Dispatch(protocol.CommandRequest{Action: "QUIT", SessionID: id})
			if quit.Type != "BYE" {
				results <- "quit failed: " + quit.Code
				return
			}
			results <- "ok"
		}()
	}
	wg.Wait()
	close(results)

	for r := range results {
		if r != "ok" {
			t.Error(r)
		}
	}
}

// TestConcurrentGuessesSameSession hammers ONE game from many goroutines.
// This is the case GameSession.mu exists for.
func TestConcurrentGuessesSameSession(t *testing.T) {
	d := setupDispatcher()

	newGame := d.Dispatch(protocol.CommandRequest{Action: "NEW_GAME", Player: "Alice"})
	id := newGame.SessionID

	letters := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
		"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

	var wg sync.WaitGroup
	responses := make(chan protocol.CommandResponse, len(letters)*2)

	// Each letter is sent twice, concurrently: exactly one must be accepted.
	for _, l := range letters {
		for range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				responses <- d.Dispatch(protocol.CommandRequest{Action: "LETTER", SessionID: id, Letter: l})
			}()
		}
	}

	// RESUME readers running at the same time as the guesses.
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.Dispatch(protocol.CommandRequest{Action: "RESUME", SessionID: id})
		}()
	}
	wg.Wait()
	close(responses)

	accepted := 0
	for resp := range responses {
		switch {
		case resp.Type == "STATE":
			accepted++
		case resp.Code == "ALREADY_GUESSED", resp.Code == "GAME_ALREADY_OVER":
			// expected under contention
		default:
			t.Errorf("unexpected response: %+v", resp)
		}
	}

	final := d.Dispatch(protocol.CommandRequest{Action: "RESUME", SessionID: id})
	tried := final.TriedLetters

	// Every accepted guess recorded exactly once, never twice.
	if accepted != len(tried) {
		t.Errorf("accepted %d guesses but session recorded %d letters: %v", accepted, len(tried), tried)
	}
	sorted := slices.Clone(tried)
	slices.Sort(sorted)
	if len(slices.Compact(sorted)) != len(tried) {
		t.Errorf("a letter was recorded twice: %v", tried)
	}
	if final.Status == "IN_PROGRESS" {
		t.Errorf("all 26 letters were offered, the game should be over, got %+v", final)
	}
}
