package handlers

import (
	"crypto/rand"
	"fmt"
	"strings"

	"hangman/internal/game"
	"hangman/internal/protocol"
)

const maxAttempts = 6

type NewGameHandler struct {
	Sessions SessionStore
}

func (h *NewGameHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	player := strings.TrimSpace(req.Player)
	if player == "" {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "INVALID_FORMAT",
			Message: "Le champ 'player' est obligatoire.",
		}
	}

	session := game.NewGameSession(rand.Text(), player, game.RandomWord(), maxAttempts)
	h.Sessions.Add(session)

	resp := session.Snapshot()
	resp.Message = fmt.Sprintf("Partie lancée pour %s !", player)
	return resp
}
