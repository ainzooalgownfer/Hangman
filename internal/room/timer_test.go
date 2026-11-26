package room

import (
	"testing"
	"time"
)

// timedDuel is a duel with a turn timer: Samy (0, conn 1) against Mohamed (1, conn 2).
func timedDuel(t *testing.T, limit time.Duration) (*Room, *fakePusher) {
	t.Helper()
	p := newFakePusher()
	r := NewRoom("R1", "Salle de Samy", "chat", 6, limit, p)
	r.Join("Samy", 1)
	r.Join("Mohamed", 2)
	return r, p
}

func TestTurnTimeLeftInView(t *testing.T) {
	r, _ := timedDuel(t, 30*time.Second)
	v := r.Status(0)
	if v.TurnTimeLeftMs <= 0 || v.TurnTimeLeftMs > 30000 {
		t.Errorf("expected 0 < turn_time_left_ms <= 30000, got %d", v.TurnTimeLeftMs)
	}
}

func TestNoTimerWhenLimitIsZero(t *testing.T) {
	r, _ := timedDuel(t, 0)
	if v := r.Status(0); v.TurnTimeLeftMs != 0 {
		t.Errorf("turn_time_limit 0 = no timer, got turn_time_left_ms %d", v.TurnTimeLeftMs)
	}
}

func TestTurnTimeoutPassesTurn(t *testing.T) {
	r, p := timedDuel(t, 100*time.Millisecond)
	time.Sleep(150 * time.Millisecond) // Samy doesn't play: one expiry at 100 ms, the next would be at 200 ms

	for _, conn := range []int64{1, 2} {
		pushed, ok := p.last(conn)
		if !ok || pushed.Turn == nil || *pushed.Turn != 1 {
			t.Fatalf("conn %d: expected a push with Mohamed's turn, got %+v", conn, pushed)
		}
	}
	v := r.Status(0)
	if v.Players[0].AttemptsLeft != 6 {
		t.Errorf("running out of time costs no life, got %d", v.Players[0].AttemptsLeft)
	}
	if v.Message == "" {
		t.Error("both players must be told why the turn passed")
	}
}

// TestStaleTimerDoesNothing: Samy plays before his time runs out. The timer of
// his turn must not fire afterwards and take Mohamed's turn away.
func TestStaleTimerDoesNothing(t *testing.T) {
	r, _ := timedDuel(t, 200*time.Millisecond)

	time.Sleep(100 * time.Millisecond)
	r.Letter(0, 1, 1, "c") // Mohamed's turn now, with a fresh 200 ms (until ~300 ms)

	time.Sleep(150 * time.Millisecond) // ~250 ms: Samy's old timer would have fired at ~200 ms
	if v := r.Status(0); v.Turn == nil || *v.Turn != 1 {
		t.Errorf("Samy's old timer must not pass Mohamed's turn, got turn %v", v.Turn)
	}
}

func TestTimerStopsWhenFinished(t *testing.T) {
	r, p := timedDuel(t, 50*time.Millisecond)
	r.Leave(0, 1) // forfeit: the duel is over
	pushes := p.count(2)

	time.Sleep(150 * time.Millisecond)
	if p.count(2) != pushes {
		t.Error("no turn timer may fire once the duel is over")
	}
}

func TestStopStopsTheTimer(t *testing.T) {
	r, p := timedDuel(t, 50*time.Millisecond)
	r.Stop() // the room is deleted
	pushes := p.count(1) + p.count(2)

	time.Sleep(150 * time.Millisecond)
	if p.count(1)+p.count(2) != pushes {
		t.Error("a stopped room must push nothing")
	}
}
