package state

import "hangman/internal/protocol"

type AbandonedState struct{}

func (s *AbandonedState) GetStatus() string {
	return "ABANDONED"
}

func (s *AbandonedState) Quit(session StateSession) protocol.CommandResponse {
	return gameOverError()
}

func (s *AbandonedState) Guess(st StateSession, l string) protocol.CommandResponse {
	return gameOverError()
}
