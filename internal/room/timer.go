package room

import "time"

// startTurnTimer gives the player whose turn it is turnLimit to play.
// Caller holds r.mu.
func (r *Room) startTurnTimer() {
	if r.turnLimit == 0 {
		return
	}
	if r.timer != nil {
		r.timer.Stop()
	}
	seq := r.turnSeq // copied: this timer belongs to this turn only
	r.turnDeadline = time.Now().Add(r.turnLimit)
	r.timer = time.AfterFunc(r.turnLimit, func() { r.turnTimeout(seq) })
}

// turnTimeout runs on the timer's own goroutine when a turn runs out: the turn
// passes to the other player, with no life lost, and both are told.
func (r *Room) turnTimeout(seq int) {
	r.mu.Lock()

	// A stale timer: the player played in time (the turn already passed), or
	// the duel is over. Its turn is gone, so it does nothing.
	if r.status != "PLAYING" || seq != r.turnSeq {
		r.mu.Unlock()
		return
	}

	r.turn = 1 - r.turn
	r.turnSeq++
	r.version++
	r.lastActive = time.Now()
	r.message = "Temps écoulé : au tour de " + r.players[r.turn].name + "."
	r.startTurnTimer() // the new player gets their full time

	// Built under the lock: both see exactly the same state.
	view0, view1 := r.view(0), r.view(1)
	conn0, conn1 := r.players[0].connID, r.players[1].connID
	r.mu.Unlock()

	r.pusher.Push(conn0, view0)
	r.pusher.Push(conn1, view1)
}

func (r *Room) stopTimer() {
	if r.timer != nil {
		r.timer.Stop()
	}
	r.turnDeadline = time.Time{}
}

