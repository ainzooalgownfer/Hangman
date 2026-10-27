package game

import (
	"testing"

	"hangman/internal/protocol"
)

// attempts safely reads resp.AttemptsLeft: it is a *int, so dereferencing nil would panic.
func attempts(t *testing.T, resp protocol.CommandResponse) int {
	t.Helper() // failures point at the caller's line, not this one
	if resp.AttemptsLeft == nil {
		t.Fatalf("attempts_left missing in response: %+v", resp)
	}
	return *resp.AttemptsLeft
}

func TestWinGame(t *testing.T) {
	s := NewGameSession("id1", "Alice", "chat", 6)

	for _, l := range []string{"c", "h", "a"} {
		resp := s.Guess(l)
		if resp.Status != "IN_PROGRESS" {
			t.Fatalf("after %q: expected IN_PROGRESS, got %s", l, resp.Status)
		}
	}

	resp := s.Guess("t")
	if resp.Status != "WON" {
		t.Fatalf("expected WON, got %s", resp.Status)
	}
	if resp.SecretWord != "CHAT" {
		t.Errorf("expected secret word CHAT, got %q", resp.SecretWord)
	}
	if got := attempts(t, resp); got != 6 {
		t.Errorf("correct guesses must not cost attempts: expected 6, got %d", got)
	}

	// A finished game rejects any further guess.
	resp = s.Guess("z")
	if resp.Type != "ERROR" || resp.Code != "GAME_ALREADY_OVER" {
		t.Errorf("expected ERROR GAME_ALREADY_OVER after win, got type=%s code=%s", resp.Type, resp.Code)
	}
}

func TestLoseGame(t *testing.T) {
	s := NewGameSession("id1", "Alice", "chat", 2)

	resp := s.Guess("x")
	if resp.Status != "IN_PROGRESS" {
		t.Fatalf("after 1st miss: expected IN_PROGRESS, got %s", resp.Status)
	}

	resp = s.Guess("y")
	if resp.Status != "LOST" {
		t.Fatalf("expected LOST, got %s", resp.Status)
	}
	if resp.SecretWord != "CHAT" {
		t.Errorf("expected secret word revealed on loss, got %q", resp.SecretWord)
	}
	if got := attempts(t, resp); got != 0 {
		t.Errorf("expected 0 attempts left, got %d", got)
	}

	resp = s.Guess("c")
	if resp.Type != "ERROR" || resp.Code != "GAME_ALREADY_OVER" {
		t.Errorf("expected ERROR GAME_ALREADY_OVER after loss, got type=%s code=%s", resp.Type, resp.Code)
	}
}

func TestWrongLetterCostsAttempt(t *testing.T) {
	s := NewGameSession("id1", "Alice", "chat", 6)

	resp := s.Guess("z")
	if got := attempts(t, resp); got != 5 {
		t.Errorf("expected 5 attempts after a miss, got %d", got)
	}
	if resp.Message != "Raté !" {
		t.Errorf("expected message %q, got %q", "Raté !", resp.Message)
	}
	if resp.MaskedWord != "_ _ _ _" {
		t.Errorf("a miss must not reveal anything, got %q", resp.MaskedWord)
	}
}

func TestCorrectLetterRevealsAllOccurrences(t *testing.T) {
	s := NewGameSession("id1", "Alice", "texte", 6)

	resp := s.Guess("e")
	if resp.MaskedWord != "_ E _ _ E" {
		t.Errorf("expected %q, got %q", "_ E _ _ E", resp.MaskedWord)
	}
	if got := attempts(t, resp); got != 6 {
		t.Errorf("expected 6 attempts after a hit, got %d", got)
	}
}

func TestDuplicateLetter(t *testing.T) {
	s := NewGameSession("id1", "Alice", "chat", 6)

	s.Guess("z") // a miss, so a wrongly charged duplicate would be visible
	resp := s.Guess("Z")

	if resp.Type != "ERROR" || resp.Code != "ALREADY_GUESSED" {
		t.Fatalf("expected ERROR ALREADY_GUESSED (case-insensitive), got type=%s code=%s", resp.Type, resp.Code)
	}

	// The duplicate must not cost an attempt or be recorded twice.
	snap := s.Snapshot()
	if got := attempts(t, snap); got != 5 {
		t.Errorf("duplicate must not cost an attempt: expected 5, got %d", got)
	}
	if len(snap.TriedLetters) != 1 {
		t.Errorf("expected 1 tried letter, got %v", snap.TriedLetters)
	}
}

func TestAccentedLetters(t *testing.T) {
	s := NewGameSession("id1", "Alice", "été", 6)

	if got := s.Snapshot().MaskedWord; got != "_ _ _" {
		t.Fatalf("expected one underscore per letter %q, got %q", "_ _ _", got)
	}

	resp := s.Guess("é")
	if resp.MaskedWord != "É _ É" {
		t.Errorf("expected %q, got %q", "É _ É", resp.MaskedWord)
	}
}

func TestNonLettersAreVisible(t *testing.T) {
	s := NewGameSession("id1", "Alice", "porte-clé", 6)

	if got := s.Snapshot().MaskedWord; got != "_ _ _ _ _ - _ _ _" {
		t.Errorf("expected hyphen visible from the start, got %q", got)
	}
}

func TestTriedLettersReturnsCopy(t *testing.T) {
	s := NewGameSession("id1", "Alice", "chat", 6)
	s.Guess("c")

	letters := s.TriedLetters()
	letters[0] = "Z"

	if got := s.TriedLetters()[0]; got != "C" {
		t.Errorf("TriedLetters leaked internal slice: expected C, got %s", got)
	}
}

func TestSnapshot(t *testing.T) {
	s := NewGameSession("id1", "Alice", "chat", 6)
	s.Guess("c")

	snap := s.Snapshot()
	if snap.Type != "STATE" || snap.SessionID != "id1" || snap.Status != "IN_PROGRESS" {
		t.Errorf("unexpected snapshot header: %+v", snap)
	}
	if snap.MaskedWord != "C _ _ _" {
		t.Errorf("expected %q, got %q", "C _ _ _", snap.MaskedWord)
	}
	if snap.SecretWord != "" {
		t.Errorf("snapshot of a game in progress must not leak the secret word, got %q", snap.SecretWord)
	}
}

func Test_SnapshotSecretWord( t * testing.T) {
	s := NewGameSession("id1", "Alice", "chat", 1)
	s.Guess("m")

	snap := s.Snapshot()
	if snap.Status != "LOST" || snap.SecretWord != "CHAT" {
		t.Fatalf("expected State Lost and return Secret Word, got status=%v, SW=%v",snap.Status, snap.SecretWord)
	}
}