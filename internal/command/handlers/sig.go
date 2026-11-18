package handlers

import (
	"hangman/internal/auth"
	"hangman/internal/protocol"
)

// helper
func validSig(token string, req protocol.CommandRequest) bool {
	msg := auth.Message(req.SessionID, req.RequestID, req.Action, req.Letter)
	return auth.Verify(token, msg, req.Sig)
}

func inValidSig() protocol.CommandResponse {
	return protocol.CommandResponse{
		Type:    "ERROR",
		Code:    "PLAYER_NOT_ALLOWED",
		Message: "Invalid signateur",
	}
}
