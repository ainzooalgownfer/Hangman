package handlers

import (
	"hangman/internal/command"
)

func RegisterAll(d *command.CommandDispatcher, store SessionStore, maxAttempts int, words []string) {
	d.Register("NEW_GAME", &NewGameHandler{Sessions: store, MaxAttempts: maxAttempts, Words: words})
	d.Register("LETTER", &GuessLetterHandler{Sessions: store})
	d.Register("RESUME", &ResumeHandler{Sessions: store})
	d.Register("QUIT", &QuitHandler{Sessions: store})

}
