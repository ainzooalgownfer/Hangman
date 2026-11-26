package handlers

import (
	"hangman/internal/errs"
	"hangman/internal/protocol"
)

type QuitHandler struct {
	Sessions SessionStore
}

func (h *QuitHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
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

	return session.QuitOnce(req.RequestID, req.ConnID)

}
