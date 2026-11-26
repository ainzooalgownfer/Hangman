package handlers

import (
	"hangman/internal/errs"
	"hangman/internal/protocol"
)

// StatusHandler returns the current state of a game without playing anything.
// The client sends it when its countdown reaches 0, to get the server's verdict.
type StatusHandler struct {
	Sessions SessionStore
}

func (h *StatusHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
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

	return session.Snapshot()

}
