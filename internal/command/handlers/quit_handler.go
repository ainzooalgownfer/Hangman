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
	resp.Type = "BYE"
	resp.Status = "ABANDONED"
	resp.SecretWord = session.SecretWord()
	resp.Message = "Partie abandonnée."

	h.Sessions.Remove(req.SessionID)

	return resp

}
