package command_test

import (
	"testing"
	"time"

	"hangman/internal/auth"
	"hangman/internal/command"
	"hangman/internal/command/handlers"
	"hangman/internal/game"
	"hangman/internal/protocol"
)

func setupDispatcher() *command.CommandDispatcher {
	return setupDispatcherWithTimer(0) // 0 = no timer
}

func setupDispatcherWithTimer(limit time.Duration) *command.CommandDispatcher {
	store := game.NewSessionManager()
	d := command.NewCommandDispatcher()
	handlers.RegisterAll(d, store, 6, limit, []string{"chat", "pomme", "robot"})
	return d
}

// signed adds the signature a real client would send, made with the session's resume_token.
// req is a copy, so the caller's request isn't changed.
func signed(req protocol.CommandRequest, token string) protocol.CommandRequest {
	req.Sig = auth.Sign(token, auth.Message(req.SessionID, req.RequestID, req.Action, req.Letter))
	return req
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
	if resEmpty.Type != "ERROR" || resEmpty.Code != "INVALID_MESSAGE" {
		t.Errorf("expected INVALID_MESSAGE on missing player, got code=%s", resEmpty.Code)
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
	if resNoSession.Type != "ERROR" || resNoSession.Code != "UNKNOWN_SESSION" {
		t.Errorf("expected UNKNOWN_SESSION, got %s", resNoSession.Code)
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
	if resUnknown.Type != "ERROR" || resUnknown.Code != "UNKNOWN_SESSION" {
		t.Errorf("expected UNKNOWN_SESSION for unknown id, got %+v", resUnknown)
	}

	// Valid guess: create a real game first, then play on it
	newGame := d.Dispatch(protocol.CommandRequest{Action: "NEW_GAME", Player: "Alice"})
	if newGame.Type != "STATE" || newGame.SessionID == "" {
		t.Fatalf("NEW_GAME failed: %+v", newGame)
	}
	id := newGame.SessionID

	resValid := d.Dispatch(signed(protocol.CommandRequest{
		Action:    "LETTER",
		SessionID: id,
		Letter:    "e",
	}, newGame.ResumeToken))

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

	resDup := d.Dispatch(signed(protocol.CommandRequest{
		Action:    "LETTER",
		SessionID: id,
		Letter:    "E",
	}, newGame.ResumeToken))
	if resDup.Code != "ALREADY_GUESSED" {
		t.Errorf("expected ALREADY_GUESSED, got %+v", resDup)
	}
}

func TestQuitHandler(t *testing.T) {
	d := setupDispatcher()

	resUnknown := d.Dispatch(protocol.CommandRequest{Action: "QUIT", SessionID: "unknown-id"})
	if resUnknown.Type != "ERROR" || resUnknown.Code != "UNKNOWN_SESSION" {
		t.Errorf("expected UNKNOWN_SESSION for unknown id, got %+v", resUnknown)
	}

	newGame := d.Dispatch(protocol.CommandRequest{Action: "NEW_GAME", Player: "Alice"})
	if newGame.Type != "STATE" || newGame.SessionID == "" {
		t.Fatalf("NEW_GAME failed: %+v", newGame)
	}
	id := newGame.SessionID

	res := d.Dispatch(signed(protocol.CommandRequest{Action: "QUIT", SessionID: id}, newGame.ResumeToken))
	if res.Type != "BYE" || res.Status != "ABANDONED" {
		t.Fatalf("expected BYE status ABANDONED, got %+v", res)
	}
	if res.SessionID != id {
		t.Errorf("expected session %s, got %s", id, res.SessionID)
	}
	if res.SecretWord == "" {
		t.Error("expected the secret word to be revealed on quit")
	}

	// The game is kept as ABANDONED: RESUME still shows it, with the word.
	resAfter := d.Dispatch(signed(protocol.CommandRequest{Action: "RESUME", SessionID: id, RequestID: 1}, newGame.ResumeToken))
	if resAfter.Type != "STATE" || resAfter.Status != "ABANDONED" {
		t.Errorf("expected STATE ABANDONED on RESUME after QUIT, got %+v", resAfter)
	}
	if resAfter.SecretWord == "" {
		t.Error("expected RESUME on an abandoned game to reveal the word")
	}

	// A finished game can't be quit or played again.
	for _, action := range []string{"QUIT", "LETTER"} {
		r := d.Dispatch(signed(protocol.CommandRequest{Action: action, SessionID: id, Letter: "e"}, newGame.ResumeToken))
		if r.Code != "INVALID_STATE" {
			t.Errorf("expected INVALID_STATE for %s on an abandoned game, got %+v", action, r)
		}
	}
}

// TestSignatureRequired checks that LETTER, QUIT and RESUME are refused
// without a signature, with a signature made with the wrong token, and when
// the message was changed after being signed.
func TestSignatureRequired(t *testing.T) {
	d := setupDispatcher()

	newGame := d.Dispatch(protocol.CommandRequest{Action: "NEW_GAME", Player: "Alice"})
	id := newGame.SessionID

	for _, action := range []string{"LETTER", "QUIT", "RESUME"} {
		req := protocol.CommandRequest{Action: action, SessionID: id, Letter: "e"}

		if r := d.Dispatch(req); r.Code != "PLAYER_NOT_ALLOWED" {
			t.Errorf("%s without sig: expected PLAYER_NOT_ALLOWED, got %+v", action, r)
		}
		if r := d.Dispatch(signed(req, "wrong-token")); r.Code != "PLAYER_NOT_ALLOWED" {
			t.Errorf("%s signed with the wrong token: expected PLAYER_NOT_ALLOWED, got %+v", action, r)
		}

		tampered := signed(req, newGame.ResumeToken)
		tampered.Letter = "a"
		if r := d.Dispatch(tampered); r.Code != "PLAYER_NOT_ALLOWED" {
			t.Errorf("%s changed after signing: expected PLAYER_NOT_ALLOWED, got %+v", action, r)
		}
	}
}

// TestResumeReplayRejected checks that a RESUME is only accepted with a
// request_id higher than every one seen before, so a captured RESUME can't
// be sent again to take the game.
func TestResumeReplayRejected(t *testing.T) {
	d := setupDispatcher()

	newGame := d.Dispatch(protocol.CommandRequest{Action: "NEW_GAME", Player: "Alice"})
	resume := func(requestID int64) protocol.CommandResponse {
		return d.Dispatch(signed(protocol.CommandRequest{Action: "RESUME", SessionID: newGame.SessionID, RequestID: requestID}, newGame.ResumeToken))
	}

	if r := resume(5); r.Type != "STATE" {
		t.Fatalf("first RESUME: expected STATE, got %+v", r)
	}
	for _, id := range []int64{5, 3, 0} {
		if r := resume(id); r.Code != "PLAYER_NOT_ALLOWED" {
			t.Errorf("RESUME with request_id %d after 5: expected PLAYER_NOT_ALLOWED, got %+v", id, r)
		}
	}
	if r := resume(6); r.Type != "STATE" {
		t.Errorf("RESUME with a new request_id: expected STATE, got %+v", r)
	}
}

// TestStatusHandler checks STATUS: it needs a signature, returns the state
// with time_left_ms while the game runs, and gives the verdict (LOST,
// TIME_UP, the word) once the time is up.
func TestStatusHandler(t *testing.T) {
	d := setupDispatcherWithTimer(30 * time.Millisecond)

	newGame := d.Dispatch(protocol.CommandRequest{Action: "NEW_GAME", Player: "Alice"})
	if newGame.TimeLeftMs <= 0 {
		t.Fatalf("expected time_left_ms in the NEW_GAME response, got %+v", newGame)
	}
	status := protocol.CommandRequest{Action: "STATUS", SessionID: newGame.SessionID}

	if r := d.Dispatch(status); r.Code != "PLAYER_NOT_ALLOWED" {
		t.Errorf("STATUS without sig: expected PLAYER_NOT_ALLOWED, got %+v", r)
	}
	if r := d.Dispatch(signed(status, newGame.ResumeToken)); r.Type != "STATE" || r.Status != "IN_PROGRESS" || r.TimeLeftMs <= 0 {
		t.Errorf("signed STATUS: expected STATE IN_PROGRESS with time_left_ms, got %+v", r)
	}

	time.Sleep(50 * time.Millisecond)

	r := d.Dispatch(signed(status, newGame.ResumeToken))
	if r.Status != "LOST" || r.Reason != "TIME_UP" || r.SecretWord == "" {
		t.Errorf("STATUS after the deadline: expected LOST, TIME_UP and the word, got %+v", r)
	}
	letter := signed(protocol.CommandRequest{Action: "LETTER", SessionID: newGame.SessionID, Letter: "e"}, newGame.ResumeToken)
	if r := d.Dispatch(letter); r.Code != "INVALID_STATE" {
		t.Errorf("LETTER after the deadline: expected INVALID_STATE, got %+v", r)
	}
}
