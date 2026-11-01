package handlers

import (
	"crypto/rand"
	"fmt"
	"strings"

	"hangman/internal/game"
	"hangman/internal/game/words"
	"hangman/internal/protocol"
)

type NewGameHandler struct {
	Sessions    SessionStore
	MaxAttempts int
	Words       []string
}

func (h *NewGameHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	player := strings.TrimSpace(req.Player)
	if player == "" {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "INVALID_MESSAGE",
			Message: "Le champ 'player' est obligatoire.",
		}
	}

	session := game.NewGameSession(rand.Text(), player, words.RandomWord(h.Words), h.MaxAttempts)
	session.Claim(req.ConnID)
	h.Sessions.Add(session)

	resp := session.Snapshot()
	resp.Message = fmt.Sprintf("Partie lancée pour %s !", player)
	// The token is added only here, not in Snapshot(): RESUME also uses Snapshot,
	// and only the player who created the game must ever receive it.
	resp.ResumeToken = session.ResumeToken()
	return resp
}
