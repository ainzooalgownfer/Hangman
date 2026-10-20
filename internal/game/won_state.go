package game

import "hangman/internal/protocol"

type WonState struct{}

func (s *WonState) Guess(session StateSession, letter string) protocol.CommandResponse {
	return protocol.CommandResponse{
		Type:    "ERROR",
		Code:    "GAME_OVER",
		Message: "La partie est terminée.",
	}
}

func (s *WonState) GetStatus() string {
	return "WON"
}
