package command_test

import (
	"testing"

	"hangman/internal/command"
	"hangman/internal/command/handlers"
	"hangman/internal/game"
	"hangman/internal/protocol"
)

func setupDispatcher() *command.CommandDispatcher {
	store := game.NewSessionManager()
	d := command.NewCommandDispatcher()
	handlers.RegisterAll(d, store, 6, []string{"chat", "pomme", "robot"})
	return d
}

func TestUnknownAction(t *testing.T) {
	d := setupDispatcher()
	req := protocol.CommandRequest{Action: "INVALID_CMD"}
	res := d.Dispatch(req)

	if res.Type != "ERROR" || res.Code != "UNKNOWN_ACTION" {
		t.Fatalf("expected ERROR with code UNKNOWN_ACTION, got type=%s, code=%s", res.Type, res.Code)
	}
}

func TestNewGameHandler(t *testing.T) {
	d := setupDispatcher()

	// Missing player
	resEmpty := d.Dispatch(protocol.CommandRequest{Action: "NEW_GAME"})
	if resEmpty.Type != "ERROR" || resEmpty.Code != "INVALID_FORMAT" {
		t.Errorf("expected INVALID_FORMAT on missing player, got code=%s", resEmpty.Code)
	}

	// Valid player
	resValid := d.Dispatch(protocol.CommandRequest{Action: "NEW_GAME", Player: "Alice"})
	if resValid.Type != "STATE" || resValid.Status != "IN_PROGRESS" {
		t.Errorf("expected STATE with status IN_PROGRESS, got status=%s", resValid.Status)
	}
}

func TestGuessLetterHandler(t *testing.T) {
	d := setupDispatcher()

	// Missing session_id
	resNoSession := d.Dispatch(protocol.CommandRequest{Action: "LETTER", Letter: "E"})
	if resNoSession.Type != "ERROR" || resNoSession.Code != "SESSION_NOT_FOUND" {
		t.Errorf("expected SESSION_NOT_FOUND, got %s", resNoSession.Code)
	}

	// Invalid character (checked before the lookup, so the ID doesn't matter)
	resInvalidChar := d.Dispatch(protocol.CommandRequest{
		Action:    "LETTER",
		SessionID: "test-session",
		Letter:    "12",
	})
	if resInvalidChar.Type != "ERROR" || resInvalidChar.Code != "INVALID_LETTER" {
		t.Errorf("expected INVALID_LETTER, got %s", resInvalidChar.Code)
	}

	// Unknown session_id
	resUnknown := d.Dispatch(protocol.CommandRequest{
		Action:    "LETTER",
		SessionID: "unknown-id",
		Letter:    "e",
	})
	if resUnknown.Type != "ERROR" || resUnknown.Code != "SESSION_NOT_FOUND" {
		t.Errorf("expected SESSION_NOT_FOUND for unknown id, got %+v", resUnknown)
	}

	// Valid guess: create a real game first, then play on it
	newGame := d.Dispatch(protocol.CommandRequest{Action: "NEW_GAME", Player: "Alice"})
	if newGame.Type != "STATE" || newGame.SessionID == "" {
		t.Fatalf("NEW_GAME failed: %+v", newGame)
	}
	id := newGame.SessionID

	resValid := d.Dispatch(protocol.CommandRequest{
		Action:    "LETTER",
		SessionID: id,
		Letter:    "e",
	})

	// Only check what is true whatever the random word is.
	if resValid.Type != "STATE" {
		t.Fatalf("expected STATE, got %+v", resValid)
	}
	if resValid.SessionID != id {
		t.Errorf("expected session %s, got %s", id, resValid.SessionID)
	}
	if len(resValid.TriedLetters) != 1 || resValid.TriedLetters[0] != "E" {
		t.Errorf("expected tried letters [E], got %v", resValid.TriedLetters)
	}
	if resValid.Status != "IN_PROGRESS" {
		t.Errorf("expected IN_PROGRESS after one guess, got %s", resValid.Status)
	}
	if resValid.AttemptsLeft == nil || (*resValid.AttemptsLeft != 6 && *resValid.AttemptsLeft != 5) {
		t.Errorf("expected 6 or 5 attempts left, got %v", resValid.AttemptsLeft)
	}

	resDup := d.Dispatch(protocol.CommandRequest{
		Action:    "LETTER",
		SessionID: id,
		Letter:    "E",
	})
	if resDup.Code != "ALREADY_GUESSED" {
		t.Errorf("expected ALREADY_GUESSED, got %+v", resDup)
	}
}

func TestQuitHandler(t *testing.T) {
	d := setupDispatcher()

	resUnknown := d.Dispatch(protocol.CommandRequest{Action: "QUIT", SessionID: "unknown-id"})
	if resUnknown.Type != "ERROR" || resUnknown.Code != "SESSION_NOT_FOUND" {
		t.Errorf("expected SESSION_NOT_FOUND for unknown id, got %+v", resUnknown)
	}

	newGame := d.Dispatch(protocol.CommandRequest{Action: "NEW_GAME", Player: "Alice"})
	if newGame.Type != "STATE" || newGame.SessionID == "" {
		t.Fatalf("NEW_GAME failed: %+v", newGame)
	}
	id := newGame.SessionID

	res := d.Dispatch(protocol.CommandRequest{Action: "QUIT", SessionID: id})
	if res.Type != "BYE" || res.Status != "ABANDONED" {
		t.Fatalf("expected BYE status ABANDONED, got %+v", res)
	}
	if res.SessionID != id {
		t.Errorf("expected session %s, got %s", id, res.SessionID)
	}
	if res.SecretWord == "" {
		t.Error("expected the secret word to be revealed on quit")
	}

	// The game must be gone: RESUME on the same ID no longer finds it.
	resAfter := d.Dispatch(protocol.CommandRequest{Action: "RESUME", SessionID: id})
	if resAfter.Code != "SESSION_NOT_FOUND" {
		t.Errorf("expected session removed after QUIT, got %+v", resAfter)
	}
}
