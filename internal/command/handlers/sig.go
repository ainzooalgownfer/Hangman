package handlers

import (
	"hangman/internal/auth"
	"hangman/internal/protocol"
)

// helper
func validSig(token string, req protocol.CommandRequest) bool {
	id := req.SessionID
	if req.RoomID != "" {
		id = req.RoomID // room actions sign room_id|request_id|action|letter
	}
	payload := req.Letter
	if req.Action == "CHAT" {
		payload = req.Text // the chat text can't be changed in transit either
	}
	return auth.Verify(token, auth.Message(id, req.RequestID, req.Action, payload), req.Sig)
}

func inValidSig() protocol.CommandResponse {
	return protocol.CommandResponse{
		Type:    "ERROR",
		Code:    "PLAYER_NOT_ALLOWED",
		Message: "Invalid signateur",
	}
}
