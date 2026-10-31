package state

import "hangman/internal/protocol"

type StateSession interface {
	ID() string
	MaskedWord() string
	AttemptsLeft() int
	TriedLetters() []string
	SecretWord() string
	SetState(state GameState)
	DecrementAttempts()
	AddTriedLetter(letter string)
	UpdateMaskedWord()
	IsWordComplete() bool
}

type GameState interface {
	Guess(session StateSession, letter string) protocol.CommandResponse
	GetStatus() string
	Quit(session StateSession) protocol.CommandResponse
}

// StateResponse builds a STATE response from the session.
func StateResponse(session StateSession, status, message string) protocol.CommandResponse {
	return protocol.CommandResponse{
		Type:         "STATE",
		SessionID:    session.ID(),
		MaskedWord:   session.MaskedWord(),
		AttemptsLeft: new(session.AttemptsLeft()),
		TriedLetters: session.TriedLetters(),
		Status:       status,
		Message:      message,
	}
}

// gameOverError is the answer to any action on a game that is already finished.
func gameOverError() protocol.CommandResponse {
	return protocol.CommandResponse{
		Type:    "ERROR",
		Code:    "GAME_ALREADY_OVER",
		Message: "La partie est terminée.",
	}
}
