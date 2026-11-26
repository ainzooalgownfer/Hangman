package state

import (
	"hangman/internal/errs"
	"hangman/internal/protocol"
)

type LostState struct{}

func (s *LostState) Guess(session StateSession, letter string) protocol.CommandResponse {
	return errs.GameOver()
}

func (s *LostState) GetStatus() string {
	return "LOST"
}

func (s *LostState) Quit(session StateSession) protocol.CommandResponse {
	return errs.GameOver()
}
