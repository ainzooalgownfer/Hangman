package room

import (
	"slices"
	"strings"
	"time"

	"hangman/internal/game/words"
	"hangman/internal/protocol"
)

// Letter plays a letter for player i. Same pattern as Join: lock, check, change,
// build the messages, unlock, then push to the opponent.
func (r *Room) Letter(i int, requestID, connID int64, letter string) protocol.CommandResponse {
	r.mu.Lock()
	if i < 0 || i >= len(r.players) || connID != r.players[i].connID {
		r.mu.Unlock()
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "PLAYER_NOT_ALLOWED",
			Message: "Cette partie est jouée depuis une autre connexion.",
		}
	}

	if requestID != 0 && requestID == r.players[i].lastRequestID {
		reply := r.players[i].lastReply // a copy
		r.mu.Unlock()
		reply.Replayed = true
		return reply
	}

	if r.status != "PLAYING" {
		r.mu.Unlock()
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "INVALID_STATE",
			Message: "La partie n'est pas en cours.",
		}
	}

	if r.turn != i {
		r.mu.Unlock()
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "INVALID_STATE",
			Message: "Ce n'est pas ton tour.",
		}
	}

	runes := []rune(letter)
	if len(runes) == 0 {
		r.mu.Unlock()
		return invalidLetter()
	}
	base, ok := words.BaseLetter(runes[0])
	if !ok {
		r.mu.Unlock()
		return invalidLetter()
	}

	if slices.Contains(r.tried, string(base)) {
		r.mu.Unlock()
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "ALREADY_GUESSED",
			Message: "Lettre déjà proposée.",
		}
	}

	// The move is valid.
	r.tried = append(r.tried, string(base))
	name := r.players[i].name
	if n := words.CountBase(r.word, base); n > 0 {
		r.players[i].score += n
		r.message = name + " a trouvé " + string(base) + "."
	} else {
		r.players[i].lives--
		r.message = name + " s'est trompé : " + string(base) + "."
	}

	switch {
	case !strings.Contains(words.Mask(r.word, r.tried), "_"):
		r.status = "FINISHED"
		s0, s1 := r.players[0].score, r.players[1].score
		switch {
		case s0 > s1:
			r.winner = 0
		case s1 > s0:
			r.winner = 1
		default:
			r.winner = -1
		}
	case r.players[i].lives == 0:
		r.status = "FINISHED"
		r.winner = 1 - i // the other player wins
	default:
		r.turn = 1 - i
	}
	r.version++
	r.lastActive = time.Now()

	reply := r.view(i)
	if requestID != 0 {
		r.players[i].lastRequestID = requestID
		r.players[i].lastReply = reply
		r.players[i].maxRequestID = max(r.players[i].maxRequestID, requestID)
	}

	toOther := r.view(1 - i)
	otherConn := r.players[1-i].connID
	r.mu.Unlock()

	r.pusher.Push(otherConn, toOther)
	return reply
}

func invalidLetter() protocol.CommandResponse {
	return protocol.CommandResponse{
		Type:    "ERROR",
		Code:    "INVALID_LETTER",
		Message: "Une seule lettre alphabétique est attendue.",
	}
}

func (r *Room) Leave(i int, connID int64) (protocol.CommandResponse, bool) {
	r.mu.Lock()
	if i < 0 || i >= len(r.players) || connID != r.players[i].connID {
		r.mu.Unlock()
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "PLAYER_NOT_ALLOWED",
			Message: "Cette partie est jouée depuis une autre connexion.",
		}, false
	}

	r.players[i].gone = true
	var push bool
	var toOther protocol.CommandResponse
	var otherConn int64

	if r.status == "PLAYING" { // forfeit
		r.status = "FINISHED"
		r.winner = 1 - i
		r.message = r.players[i].name + " a quitté la partie."
		r.version++
		r.lastActive = time.Now()
		toOther = r.view(1 - i)
		otherConn = r.players[1-i].connID
		push = true
	}
	empty := true
	for _, p := range r.players {
		if !p.gone {
			empty = false
			break
		}
	}

	r.mu.Unlock()
	if push {
		r.pusher.Push(otherConn, toOther)
	}
	return protocol.CommandResponse{
		Type:    "BYE",
		RoomID:  r.id, // set once at creation and never changed: safe to read without the lock
		Message: "Tu as quitté la salle.",
	}, empty
}
