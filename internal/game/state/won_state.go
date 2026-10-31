package state

import "hangman/internal/protocol"

type WonState struct{}

func (s *WonState) Guess(session StateSession, letter string) protocol.CommandResponse {
	return gameOverError()
}

func (s *WonState) GetStatus() string {
	return "WON"
}

func (s *WonState) Quit(session StateSession) protocol.CommandResponse {
	return gameOverError()
}
