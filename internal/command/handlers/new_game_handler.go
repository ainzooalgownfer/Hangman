package handlers

import (
	"crypto/rand"
	"fmt"
	"strings"

	"hangman/internal/game"
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
			Code:    "INVALID_FORMAT",
			Message: "Le champ 'player' est obligatoire.",
		}
	}

	session := game.NewGameSession(rand.Text(), player, game.RandomWord(h.Words), h.MaxAttempts)
	h.Sessions.Add(session)

	resp := session.Snapshot()
	resp.Message = fmt.Sprintf("Partie lancée pour %s !", player)
	return resp
}
