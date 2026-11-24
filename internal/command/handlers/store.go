package handlers

import (
	"hangman/internal/game"
	"hangman/internal/protocol"
	"hangman/internal/room"
)

type SessionStore interface {
	Add(s *game.GameSession)
	Get(id string) (*game.GameSession, bool)
	Remove(id string)
}

// RoomStore is what the room handlers need from the room manager.
type RoomStore interface {
	Create(roomName, player string, connID int64) protocol.CommandResponse
	Get(id string) (*room.Room, bool)
	List() []protocol.RoomSummary
	Remove(id string)
}
