package game

import (
	"hangman/internal/protocol"
	"slices"
	"strings"
)

type InProgressState struct{}

func (s *InProgressState) Guess(session StateSession, letter string) protocol.CommandResponse {

	upper := strings.ToUpper(letter)
	found := strings.Contains(session.SecretWord(), upper)
	message := "Bien vu !"

	if slices.Contains(session.TriedLetters(), upper) {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "ALREADY_GUESSED",
			Message: "Lettre déjà proposée.",
		}
	}

	session.AddTriedLetter(letter)

	if !found {
		session.DecrementAttempts()
		message = "Raté !"
	}

	session.UpdateMaskedWord()

	if session.IsWordComplete() {
		session.SetState(&WonState{})
		resp := stateResponse(session, "WON", "bravo !")
		resp.SecretWord = session.SecretWord()
		return resp
	}

	if session.AttemptsLeft() == 0 {
		session.SetState(&LostState{})
		resp := stateResponse(session, "LOST", "game over")
		resp.SecretWord = session.SecretWord()
		return resp
	}

	return stateResponse(session, "IN_PROGRESS", message)

}

func (s *InProgressState) GetStatus() string {
	return "IN_PROGRESS"
}
