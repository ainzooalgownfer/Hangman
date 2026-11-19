package handlers

import "hangman/internal/protocol"

// StatusHandler returns the current state of a game without playing anything.
// The client sends it when its countdown reaches 0, to get the server's verdict.
type StatusHandler struct {
	Sessions SessionStore
}

func (h *StatusHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
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

	return session.Snapshot()

}
