package handlers

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"hangman/internal/errs"
	"hangman/internal/game"
	"hangman/internal/game/words"
	"hangman/internal/protocol"
)

type NewGameHandler struct {
	Sessions    SessionStore
	MaxAttempts int
	TimeLimit   time.Duration
	Words       []string
}

func (h *NewGameHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	player := strings.TrimSpace(req.Player)
	if player == "" {
		return errs.MissingPlayer()
	}

	session := game.NewGameSession(rand.Text(), player, words.RandomWord(h.Words), h.MaxAttempts)
	session.StartTimer(h.TimeLimit)
	session.Claim(req.ConnID)
	h.Sessions.Add(session)

	resp := session.Snapshot()
	resp.Message = fmt.Sprintf("Partie lancée pour %s !", player)
	// The token is added only here, not in Snapshot(): RESUME also uses Snapshot,
	// and only the player who created the game must ever receive it.
	resp.ResumeToken = session.ResumeToken()
	return resp
}
