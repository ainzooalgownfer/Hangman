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
	handlers.RegisterAll(dispatcher, store)

	server := network.NewTCPServer(":8080", dispatcher)
	if err := server.Start(); err != nil {
		log.Fatalf("Erreur serveur: %v", err)
	}
}
