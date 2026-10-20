package handlers

import (
	"hangman/internal/protocol"
)

type QuitHandler struct{}

func (h *QuitHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	if req.SessionID == "" {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "SESSION_NOT_FOUND",
			Message: "Identifiant de session manquant.",
		}
	}

	return protocol.CommandResponse{
		Type:       "BYE",
		SessionID:  req.SessionID,
		Status:     "ABANDONED",
		SecretWord: "TEXTE",
		Message:    "Partie abandonnée.",
	}
}
