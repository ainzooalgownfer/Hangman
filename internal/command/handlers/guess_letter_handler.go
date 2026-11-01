package handlers

import (
	"strings"

	"hangman/internal/game/words"
	"hangman/internal/protocol"
)

type GuessLetterHandler struct {
	Sessions SessionStore
}

func (h *GuessLetterHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	if req.SessionID == "" {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "UNKNOWN_SESSION",
			Message: "Identifiant de session manquant.",
		}
	}

	trimmed := strings.TrimSpace(req.Letter)
	runes := []rune(trimmed)
	if len(runes) != 1 {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "INVALID_LETTER",
			Message: "Une seule lettre alphabétique est attendue.",
		}
	}

	if _, ok := words.BaseLetter(runes[0]); !ok {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "INVALID_LETTER",
			Message: "Une seule lettre alphabétique est attendue.",
		}
	}

	session, ok := h.Sessions.Get(req.SessionID)
	if !ok {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "UNKNOWN_SESSION",
			Message: "Session introuvable",
		}
	}

	return session.GuessOnce(req.RequestID, req.ConnID, trimmed)

}
