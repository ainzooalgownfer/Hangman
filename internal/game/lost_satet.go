package game

import (
	"hangman/internal/protocol"
)

type LostState struct{}

// with no reciever they r plain package level funcs

func (s *LostState) Guess(session StateSession, letter string) protocol.CommandResponse {
	return protocol.CommandResponse{
		Type:    "ERROR",
		Code:    "GAME_OVER",
		Message: "La partie est terminée.",
	}
}

func (s *LostState) GetStatus() string {
	return "LOST"
}
