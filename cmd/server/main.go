package main

import (
	"log"
	"log/slog"
	"time"

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
		"ttl value", cfg.SessionTTL,
		"words", len(cfg.Words),
	)

	store := game.NewSessionManager()
	dispatcher := command.NewCommandDispatcher()
	handlers.RegisterAll(dispatcher, store, cfg.MaxAttempts, cfg.Words)

	stop := store.StartJanitor(cfg.SessionTTL, 30*time.Second)
	defer stop()

	server := network.NewTCPServer(cfg.Address, cfg.MaxLineSize, cfg.IdleTimeout, dispatcher)
	if err := server.Start(); err != nil {
		log.Fatalf("Erreur serveur: %v", err)
	}
}
