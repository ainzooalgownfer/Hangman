package handlers

import (
	"crypto/subtle"
	"hangman/internal/protocol"
)

type ResumeHandler struct {
	Sessions SessionStore
}

func (h *ResumeHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
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

	validToken := subtle.ConstantTimeCompare([]byte(req.ResumeToken), []byte(session.ResumeToken()))
	if validToken != 1 {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "PLAYER_NOT_ALLOWED",
			Message: "Jeton de reprise invalide.",
		}
	}

	session.Claim(req.ConnID)
	resp := session.Snapshot()
	resp.Message = "Session reprise avec succès."

	return resp

}
