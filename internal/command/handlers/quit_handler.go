package handlers

import (
	"hangman/internal/protocol"
)

type QuitHandler struct {
	Sessions SessionStore
}

func (h *QuitHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	if req.SessionID == "" {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "UNKNOWN_SESSION",
			Message: "Identifiant de session manquant.",
		}
	}

	session, ok := h.Sessions.Get(req.SessionID)
	if !ok {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "UNKNOWN_SESSION",
			Message: "Session introuvable.",
		}
	}

	return session.QuitOnce(req.RequestID, req.ConnID)

}
