package main

import (
	"log"

	"hangman/internal/command"
	"hangman/internal/command/handlers"
	"hangman/internal/game"
	"hangman/internal/network"
)

func main() {
	store := game.NewSessionManager() 
	dispatcher := command.NewCommandDispatcher()


	dispatcher.Register("NEW_GAME", &handlers.NewGameHandler{Sessions: store})
	dispatcher.Register("LETTER", &handlers.GuessLetterHandler{})
	dispatcher.Register("RESUME", &handlers.ResumeHandler{})
	dispatcher.Register("QUIT", &handlers.QuitHandler{})

	server := network.NewTCPServer(":8080", dispatcher)
	if err := server.Start(); err != nil {
		log.Fatalf("Erreur serveur: %v", err)
	}
}
