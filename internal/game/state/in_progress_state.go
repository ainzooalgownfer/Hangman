package state

import (
	"slices"

	"hangman/internal/game/words"
	"hangman/internal/protocol"
)

type InProgressState struct{}

func (s *InProgressState) Guess(session StateSession, letter string) protocol.CommandResponse {

	base, _ := words.BaseLetter([]rune(letter)[0])

	upper := string(base)
	found := words.ContainsBase(session.SecretWord(), base)
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
		resp := StateResponse(session, "WON", "bravo !")
		resp.SecretWord = session.SecretWord()
		return resp
	}

	if session.AttemptsLeft() == 0 {
		session.SetState(&LostState{})
		resp := StateResponse(session, "LOST", "game over")
		resp.SecretWord = session.SecretWord()
		return resp
	}

	return StateResponse(session, "IN_PROGRESS", message)

}

func (s *InProgressState) Quit(session StateSession) protocol.CommandResponse {
	session.SetState(&AbandonedState{})
	resp := StateResponse(session, "ABANDONED", "Partie abandonnée.")
	resp.Type = "BYE"
	resp.SecretWord = session.SecretWord()
	return resp
}

func (s *InProgressState) GetStatus() string {
	return "IN_PROGRESS"
}
