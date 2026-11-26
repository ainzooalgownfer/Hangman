package state

import (
	"hangman/internal/errs"
	"hangman/internal/protocol"
)

type AbandonedState struct{}

func (s *AbandonedState) GetStatus() string {
	return "ABANDONED"
}

func (s *AbandonedState) Quit(session StateSession) protocol.CommandResponse {
	return errs.GameOver()
}

func (s *AbandonedState) Guess(st StateSession, l string) protocol.CommandResponse {
	return errs.GameOver()
}
