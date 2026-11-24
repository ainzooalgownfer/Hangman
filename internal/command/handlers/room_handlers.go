package handlers

import (
	"strings"

	"hangman/internal/game/words"
	"hangman/internal/protocol"
	"hangman/internal/room"
)

// roomPlayer finds who sent req: the player whose token signed it.
// The client never says "I'm player 0": the signature proves it.
func roomPlayer(r *room.Room, req protocol.CommandRequest) (int, bool) {
	for i := 0; i < 2; i++ {
		if t := r.Token(i); t != "" && validSig(t, req) {
			return i, true
		}
	}
	return -1, false
}

func unknownRoom() protocol.CommandResponse {
	return protocol.CommandResponse{Type: "ERROR", Code: "UNKNOWN_ROOM", Message: "Salle introuvable."}
}

func missingPlayer() protocol.CommandResponse {
	return protocol.CommandResponse{
		Type:    "ERROR",
		Code:    "INVALID_MESSAGE",
		Message: "Le champ 'player' est obligatoire.",
	}
}


type ListRoomsHandler struct {
	Rooms RoomStore
}

func (h *ListRoomsHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	return protocol.CommandResponse{Type: "ROOMS", Rooms: h.Rooms.List()}
}

// CreateRoomHandler answers CREATE_ROOM. No signature: the player gets their token in the reply.
type CreateRoomHandler struct {
	Rooms RoomStore
}

func (h *CreateRoomHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	
	player := strings.TrimSpace(req.Player)
	if player == "" {
		return missingPlayer()
	}
	return h.Rooms.Create(strings.TrimSpace(req.RoomName), player, req.ConnID)
}

// JoinRoomHandler answers JOIN_ROOM. No signature: like CREATE_ROOM, the token comes back in the reply.
type JoinRoomHandler struct {
	Rooms RoomStore
}

func (h *JoinRoomHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	player := strings.TrimSpace(req.Player)
	if player == "" {
		return missingPlayer()
	}
	r, ok := h.Rooms.Get(req.RoomID)
	if !ok {
		return unknownRoom()
	}
	return r.Join(player, req.ConnID) // a full room answers PLAYER_NOT_ALLOWED itself
}

// RoomLetterHandler answers ROOM_LETTER.
type RoomLetterHandler struct {
	Rooms RoomStore
}

func (h *RoomLetterHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	r, ok := h.Rooms.Get(req.RoomID)
	if !ok {
		return unknownRoom()
	}
	
	i, ok := roomPlayer(r, req)
	if !ok {
		return inValidSig()
	}
	trimmed := strings.TrimSpace(req.Letter)
	runes := []rune(trimmed)
	if len(runes) != 1 {
		return invalidLetterResponse()
	}
	if _, ok := words.BaseLetter(runes[0]); !ok {
		return invalidLetterResponse()
	}
	return r.Letter(i, req.RequestID, req.ConnID, trimmed)
}

func invalidLetterResponse() protocol.CommandResponse {
	return protocol.CommandResponse{
		Type:    "ERROR",
		Code:    "INVALID_LETTER",
		Message: "Une seule lettre alphabétique est attendue.",
	}
}

// LeaveRoomHandler answers LEAVE_ROOM, and deletes the room once nobody is left.
type LeaveRoomHandler struct {
	Rooms RoomStore
}

func (h *LeaveRoomHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	r, ok := h.Rooms.Get(req.RoomID)
	if !ok {
		return unknownRoom()
	}
	i, ok := roomPlayer(r, req)
	if !ok {
		return inValidSig()
	}
	resp, empty := r.Leave(i, req.ConnID)
	if empty {
		h.Rooms.Remove(req.RoomID)
	}
	return resp
}
