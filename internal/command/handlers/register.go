package handlers

import (
	"hangman/internal/command"
)

func RegisterAll(d *command.CommandDispatcher, store SessionStore) {
	d.Register("NEW_GAME", &NewGameHandler{Sessions: store})
	d.Register("LETTER", &GuessLetterHandler{Sessions: store})
	d.Register("RESUME", &ResumeHandler{Sessions: store})
	d.Register("QUIT", &QuitHandler{Sessions: store})

}
