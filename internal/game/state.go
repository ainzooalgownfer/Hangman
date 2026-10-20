package game

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
}

func stateResponse(session StateSession, status, message string) protocol.CommandResponse {
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
