package handlers

import (
	"hangman/internal/protocol"
)

type ResumeHandler struct {
	Sessions SessionStore
}

func (h *ResumeHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	if req.SessionID == "" {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "SESSION_NOT_FOUND",
			Message: "Identifiant de session manquant.",
		}
	}
	session, ok := h.Sessions.Get(req.SessionID)

	if !ok {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "SESSION_NOT_FOUND",
			Message: "Session introuvable.",
		}
	}

	resp := session.Snapshot()
	resp.Message = "Session reprise avec succès."

	return resp

}
