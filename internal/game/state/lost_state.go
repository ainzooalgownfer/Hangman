package state

import (
	"hangman/internal/protocol"
)

type LostState struct{}

func (s *LostState) Guess(session StateSession, letter string) protocol.CommandResponse {
	return gameOverError()
}

func (s *LostState) GetStatus() string {
	return "LOST"
}

func (s *LostState) Quit(session StateSession) protocol.CommandResponse {
	return gameOverError()
}
