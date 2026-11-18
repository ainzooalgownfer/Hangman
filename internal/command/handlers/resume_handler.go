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

	if !validSig(session.ResumeToken(), req) {
		return inValidSig()
	}

	resp := session.ResumeOnce(req.RequestID, req.ConnID)
	if resp.Type == "ERROR" {
		return resp
	}
	resp.Message = "Session reprise avec succès."

	return resp

}
