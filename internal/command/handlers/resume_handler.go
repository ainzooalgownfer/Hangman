package handlers

import (
	"hangman/internal/protocol"
)

type ResumeHandler struct{}

func (h *ResumeHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	if req.SessionID == "" {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "SESSION_NOT_FOUND",
			Message: "Identifiant de session manquant.",
		}
	}

	return protocol.CommandResponse{
		Type:         "STATE",
		SessionID:    req.SessionID,
		MaskedWord:   "_ E _ _ E",
		AttemptsLeft: new(6), // FIX (bug 2): field is now *int; new(expr) needs Go 1.26+
		TriedLetters: []string{"E"},
		Status:       "IN_PROGRESS",
		Message:      "Session reprise avec succès.",
	}
}
