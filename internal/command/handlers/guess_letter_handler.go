package handlers

import (
	"strings"
	"unicode"

	"hangman/internal/protocol"
)

type GuessLetterHandler struct{}

func (h *GuessLetterHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	if req.SessionID == "" {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "SESSION_NOT_FOUND",
			Message: "Identifiant de session manquant.",
		}
	}

	trimmed := strings.TrimSpace(req.Letter)
	runes := []rune(trimmed)
	if len(runes) != 1 || !unicode.IsLetter(runes[0]) {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "INVALID_LETTER",
			Message: "Une seule lettre alphabétique est attendue.",
		}
	}

	return protocol.CommandResponse{
		Type:         "STATE",
		SessionID:    req.SessionID,
		MaskedWord:   "_ E _ _ E",
		AttemptsLeft: new(6), // FIX (bug 2): field is now *int; new(expr) needs Go 1.26+
		TriedLetters: []string{strings.ToUpper(string(runes[0]))},
		Status:       "IN_PROGRESS",
		Message:      "Bien vu !",
	}
}
