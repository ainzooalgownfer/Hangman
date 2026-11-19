package handlers

import (
	"hangman/internal/command"
	"time"
)

func RegisterAll(d *command.CommandDispatcher, store SessionStore, maxAttempts int,
	timelimit time.Duration, words []string) {
	d.Register("NEW_GAME", &NewGameHandler{Sessions: store, MaxAttempts: maxAttempts, TimeLimit: timelimit, Words: words})
	d.Register("LETTER", &GuessLetterHandler{Sessions: store})
	d.Register("RESUME", &ResumeHandler{Sessions: store})
	d.Register("QUIT", &QuitHandler{Sessions: store})
	d.Register("STATUS", &StatusHandler{Sessions: store})

}
