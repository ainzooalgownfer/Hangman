package game

import (
	"hangman/internal/protocol"
	"slices"
)

type InProgressState struct{}

func (s *InProgressState) Guess(session StateSession, letter string) protocol.CommandResponse {

	base , _  := BaseLetter([]rune(letter)[0])

	upper := string(base)
	found := ContainsBase(session.SecretWord() , base)
	message := "Bien vu !"

	if slices.Contains(session.TriedLetters(), upper) {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "ALREADY_GUESSED",
			Message: "Lettre déjà proposée.",
		}
	}

	session.AddTriedLetter(upper)

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
