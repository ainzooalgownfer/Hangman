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
	d.Register("NEW_GAME", &handlers.NewGameHandler{Sessions: store})
	d.Register("LETTER", &handlers.GuessLetterHandler{})
	d.Register("RESUME", &handlers.ResumeHandler{})
	d.Register("QUIT", &handlers.QuitHandler{})
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

	// Invalid character
	resInvalidChar := d.Dispatch(protocol.CommandRequest{
		Action:    "LETTER",
		SessionID: "test-session",
		Letter:    "12",
	})
	if resInvalidChar.Type != "ERROR" || resInvalidChar.Code != "INVALID_LETTER" {
		t.Errorf("expected INVALID_LETTER, got %s", resInvalidChar.Code)
	}

	// Valid guess
	resValid := d.Dispatch(protocol.CommandRequest{
		Action:    "LETTER",
		SessionID: "test-session",
		Letter:    "e",
	})
	if resValid.Type != "STATE" {
		t.Errorf("expected STATE, got %s", resValid.Type)
	}
}

func TestQuitHandler(t *testing.T) {
	d := setupDispatcher()
	res := d.Dispatch(protocol.CommandRequest{
		Action:    "QUIT",
		SessionID: "test-session",
	})

	if res.Type != "BYE" || res.Status != "ABANDONED" {
		t.Errorf("expected BYE status ABANDONED, got type=%s status=%s", res.Type, res.Status)
	}
}
