package main

import (
	"log"
	"log/slog"

	"hangman/internal/command"
	"hangman/internal/command/handlers"
	"hangman/internal/config"
	"hangman/internal/game"
	"hangman/internal/network"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	slog.Info("config loaded",
		"address", cfg.Address,
		"max_attempts", cfg.MaxAttempts,
		"idle_timeout", cfg.IdleTimeout,
		"words", len(cfg.Words),
	)

	store := game.NewSessionManager()
	dispatcher := command.NewCommandDispatcher()
	handlers.RegisterAll(dispatcher, store, cfg.MaxAttempts, cfg.Words)

	server := network.NewTCPServer(cfg.Address, cfg.MaxLineSize, cfg.IdleTimeout, dispatcher)
	if err := server.Start(); err != nil {
		log.Fatalf("Erreur serveur: %v", err)
	}
}
