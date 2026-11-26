package room

import (
	"slices"
	"sync"
	"testing"
)

// duel returns a room where Samy (player 0, connID 1) and Mohamed (player 1,
// connID 2) have both joined: the duel is on and it is Samy's turn.
func duel(t *testing.T, word string, lives int) (*Room, *fakePusher) {
	t.Helper()
	p := newFakePusher()
	r := NewRoom("R1", "Salle de Samy", word, lives, 0, p)
	r.Join("Samy", 1)
	r.Join("Mohamed", 2)
	return r, p
}

func TestLetterWrongConnection(t *testing.T) {
	r, _ := duel(t, "chat", 6)
	if resp := r.Letter(0, 1, 99, "c"); resp.Code != "PLAYER_NOT_ALLOWED" {
		t.Errorf("a letter from another connection: expected PLAYER_NOT_ALLOWED, got %+v", resp)
	}
}

func TestLetterNotYourTurn(t *testing.T) {
	r, _ := duel(t, "chat", 6)
	if resp := r.Letter(1, 1, 2, "c"); resp.Code != "INVALID_STATE" {
		t.Errorf("Mohamed plays on Samy's turn: expected INVALID_STATE, got %+v", resp)
	}
}

func TestLetterScoresAndPassesTurn(t *testing.T) {
	r, p := duel(t, "chat", 6)

	resp := r.Letter(0, 1, 1, "c")
	if resp.Type != "ROOM" || resp.Players[0].Score != 1 || intVal(t, "turn", resp.Turn) != 1 {
		t.Fatalf("Samy finds C: expected score 1 and Mohamed's turn, got %+v", resp)
	}
	if resp.MaskedWord != "C _ _ _" || !slices.Equal(resp.TriedLetters, []string{"C"}) {
		t.Errorf("expected masked %q and tried [C], got %q %v", "C _ _ _", resp.MaskedWord, resp.TriedLetters)
	}

	pushed, ok := p.last(2)
	if !ok {
		t.Fatal("Mohamed got no push after Samy's letter")
	}
	if pushed.MaskedWord != "C _ _ _" || intVal(t, "you", pushed.You) != 1 || intVal(t, "turn", pushed.Turn) != 1 {
		t.Errorf("push to Mohamed: expected the same word, you=1, turn=1, got %+v", pushed)
	}
	if pushed.Version != resp.Version {
		t.Errorf("both players must see the same version, got %d and %d", resp.Version, pushed.Version)
	}
}

func TestLetterMissCostsLife(t *testing.T) {
	r, _ := duel(t, "chat", 6)
	resp := r.Letter(0, 1, 1, "z")
	if resp.Players[0].AttemptsLeft != 5 || resp.Players[0].Score != 0 || intVal(t, "turn", resp.Turn) != 1 {
		t.Errorf("Samy misses: expected 5 lives, score 0, Mohamed's turn, got %+v", resp.Players)
	}
}

func TestAlreadyGuessedKeepsTurn(t *testing.T) {
	r, _ := duel(t, "chat", 6)
	r.Letter(0, 1, 1, "c")

	if resp := r.Letter(1, 1, 2, "c"); resp.Code != "ALREADY_GUESSED" {
		t.Fatalf("C again: expected ALREADY_GUESSED, got %+v", resp)
	}
	if resp := r.Letter(1, 2, 2, "h"); resp.Type != "ROOM" {
		t.Errorf("after a repeated letter it is still Mohamed's turn, got %+v", resp)
	}
}

func TestAccentedLetterScores(t *testing.T) {
	r, _ := duel(t, "élève", 6)
	resp := r.Letter(0, 1, 1, "e")
	if resp.Players[0].Score != 3 || resp.MaskedWord != "É _ È _ E" {
		t.Errorf("E reveals É, È and E: expected score 3 and %q, got %d %q", "É _ È _ E", resp.Players[0].Score, resp.MaskedWord)
	}
}

func TestWordCompleteDraw(t *testing.T) {
	r, p := duel(t, "chat", 6)
	r.Letter(0, 1, 1, "c")
	r.Letter(1, 1, 2, "h")
	r.Letter(0, 2, 1, "a")
	resp := r.Letter(1, 2, 2, "t") // 2 points each

	if resp.Status != "FINISHED" || resp.Result != "DRAW" || resp.SecretWord != "CHAT" {
		t.Errorf("word complete at 2-2: expected FINISHED, DRAW, CHAT, got %+v", resp)
	}
	if resp.Turn != nil {
		t.Errorf("no turn once the duel is over, got %d", *resp.Turn)
	}
	if pushed, _ := p.last(1); pushed.Result != "DRAW" {
		t.Errorf("Samy must also learn it's a draw, got %+v", pushed)
	}
}

func TestWordCompleteHigherScoreWins(t *testing.T) {
	r, p := duel(t, "chats", 6)
	r.Letter(0, 1, 1, "c")
	r.Letter(1, 1, 2, "h")
	r.Letter(0, 2, 1, "a")
	r.Letter(1, 2, 2, "t")
	resp := r.Letter(0, 3, 1, "s") // Samy 3, Mohamed 2

	if resp.Status != "FINISHED" || resp.Result != "WIN" {
		t.Errorf("Samy finishes the word 3-2: expected WIN, got %+v", resp)
	}
	if pushed, _ := p.last(2); pushed.Result != "LOSE" || pushed.SecretWord != "CHATS" {
		t.Errorf("Mohamed must get LOSE and the word, got %+v", pushed)
	}
}

func TestNoLivesLeftOtherWins(t *testing.T) {
	r, p := duel(t, "chat", 2)
	r.Letter(0, 1, 1, "z")
	r.Letter(1, 1, 2, "c")
	resp := r.Letter(0, 2, 1, "x") // Samy's 2nd miss: no lives left

	if resp.Status != "FINISHED" || resp.Result != "LOSE" || resp.SecretWord != "CHAT" {
		t.Errorf("Samy has no lives left: expected FINISHED, LOSE, CHAT, got %+v", resp)
	}
	if pushed, _ := p.last(2); pushed.Result != "WIN" {
		t.Errorf("Mohamed must get WIN, got %+v", pushed)
	}
}

func TestLetterAfterFinished(t *testing.T) {
	r, _ := duel(t, "chat", 1)
	r.Letter(0, 1, 1, "z") // Samy loses at once
	if resp := r.Letter(1, 1, 2, "c"); resp.Code != "INVALID_STATE" {
		t.Errorf("a letter after the end: expected INVALID_STATE, got %+v", resp)
	}
}

func TestLetterReplay(t *testing.T) {
	r, p := duel(t, "chat", 6)
	first := r.Letter(0, 7, 1, "c")
	pushes := p.count(2)

	again := r.Letter(0, 7, 1, "c") // same request_id: the answer was lost
	if !again.Replayed || again.Players[0].Score != 1 || again.Version != first.Version {
		t.Errorf("replay: expected the same answer marked replayed, score still 1, got %+v", again)
	}
	if p.count(2) != pushes {
		t.Error("a replay must not push again to the opponent")
	}
}

func TestLeaveWhileWaitingEmptiesRoom(t *testing.T) {
	r := NewRoom("R1", "Salle de Samy", "chat", 6, 0, newFakePusher())
	r.Join("Samy", 1)

	resp, empty := r.Leave(0, 1)
	if resp.Type != "BYE" || !empty {
		t.Errorf("the only player leaves: expected BYE and an empty room, got %+v empty=%v", resp, empty)
	}
}

func TestLeaveDuringDuelIsForfeit(t *testing.T) {
	r, p := duel(t, "chat", 6)

	resp, empty := r.Leave(0, 1)
	if resp.Type != "BYE" || empty {
		t.Fatalf("Samy leaves mid-duel: expected BYE, room not empty yet, got %+v empty=%v", resp, empty)
	}
	pushed, ok := p.last(2)
	if !ok || pushed.Status != "FINISHED" || pushed.Result != "WIN" {
		t.Errorf("Mohamed must win by forfeit, got %+v", pushed)
	}

	if _, empty := r.Leave(1, 2); !empty {
		t.Error("once both players have left, the room is empty")
	}
}

func TestLeaveWrongConnection(t *testing.T) {
	r, _ := duel(t, "chat", 6)
	if resp, _ := r.Leave(0, 99); resp.Code != "PLAYER_NOT_ALLOWED" {
		t.Errorf("leave from another connection: expected PLAYER_NOT_ALLOWED, got %+v", resp)
	}
}

// TestConcurrentLetters sends two letters from Samy at the same moment, many
// times: exactly one must be played, the other refused (it's no longer his turn).
func TestConcurrentLetters(t *testing.T) {
	for round := 0; round < 50; round++ {
		r, _ := duel(t, "chat", 6)

		var wg sync.WaitGroup
		results := make(chan string, 2)
		for i, l := range []string{"c", "h"} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				results <- r.Letter(0, int64(i+1), 1, l).Type
			}()
		}
		wg.Wait()
		close(results)

		played := 0
		for typ := range results {
			if typ == "ROOM" {
				played++
			}
		}
		if played != 1 {
			t.Fatalf("round %d: expected exactly 1 letter played, got %d", round, played)
		}
	}
}

func TestResumeMovesConnection(t *testing.T) {
	r, p := duel(t, "chat", 6)
	r.Disconnected(1) // Samy's connection drops

	resp := r.Resume(0, 5, 3) // he comes back on connection 3
	if resp.Type != "ROOM" || intVal(t, "you", resp.You) != 0 || !resp.Players[0].Connected {
		t.Fatalf("resume: expected ROOM, you=0, Samy connected, got %+v", resp)
	}
	if pushed, _ := p.last(2); !pushed.Players[0].Connected {
		t.Errorf("Mohamed must be told Samy is back, got %+v", pushed.Players)
	}
	if r := r.Letter(0, 6, 1, "c"); r.Code != "PLAYER_NOT_ALLOWED" {
		t.Errorf("the old connection must be refused, got %+v", r)
	}
	if r := r.Letter(0, 6, 3, "c"); r.Type != "ROOM" {
		t.Errorf("the new connection must be able to play, got %+v", r)
	}
}

func TestResumeReplayRefused(t *testing.T) {
	r, _ := duel(t, "chat", 6)

	if resp := r.Resume(0, 5, 3); resp.Type != "ROOM" {
		t.Fatalf("first resume: expected ROOM, got %+v", resp)
	}
	for _, id := range []int64{5, 4, 0} {
		if resp := r.Resume(0, id, 4); resp.Code != "PLAYER_NOT_ALLOWED" {
			t.Errorf("resume with request_id %d after 5: expected PLAYER_NOT_ALLOWED, got %+v", id, resp)
		}
	}

	r.Letter(0, 10, 3, "c") // a letter also raises the highest request_id seen
	if resp := r.Resume(0, 9, 4); resp.Code != "PLAYER_NOT_ALLOWED" {
		t.Errorf("resume with request_id 9 after a letter with 10: expected PLAYER_NOT_ALLOWED, got %+v", resp)
	}
}

func TestResumeAfterLeaveRefused(t *testing.T) {
	r, _ := duel(t, "chat", 6)
	r.Leave(0, 1)
	if resp := r.Resume(0, 5, 3); resp.Code != "PLAYER_NOT_ALLOWED" {
		t.Errorf("a player who left can't come back, got %+v", resp)
	}
}

func TestStatusChangesNothing(t *testing.T) {
	r, p := duel(t, "chat", 6)
	before := r.Status(1)
	pushes := p.count(1)

	again := r.Status(1)
	if again.Type != "ROOM" || again.Version != before.Version || intVal(t, "you", again.You) != 1 {
		t.Errorf("status: expected the same view for player 1, got %+v", again)
	}
	if p.count(1) != pushes {
		t.Error("status must push nothing")
	}
}
