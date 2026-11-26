package handlers

import (
	"hangman/internal/errs"
	"hangman/internal/protocol"
)

type ResumeHandler struct {
	Sessions SessionStore
}

func (h *ResumeHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	if req.SessionID == "" {
		return errs.MissingSession()
	}
	session, ok := h.Sessions.Get(req.SessionID)

	if !ok {
		return errs.UnknownSession()
	}

	if !validSig(session.ResumeToken(), req) {
		return errs.InvalidSignature()
	}

	resp := session.ResumeOnce(req.RequestID, req.ConnID)
	if resp.Type == "ERROR" {
		return resp
	}
	resp.Message = "Session reprise avec succès."

	return resp

}
