package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
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
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	slog.Info("config loaded",
		"address", cfg.Address,
		"max_attempts", cfg.MaxAttempts,
		"idle_timeout", cfg.IdleTimeout,
		"session_ttl", cfg.SessionTTL,
		"words", len(cfg.Words),
	)

	store := game.NewSessionManager()
	dispatcher := command.NewCommandDispatcher()
	handlers.RegisterAll(dispatcher, store, cfg.MaxAttempts, cfg.Words)

	stop := store.StartJanitor(cfg.SessionTTL, 30*time.Second)
	defer stop()

	server := network.NewTCPServer(cfg.Address, cfg.MaxLineSize, cfg.IdleTimeout, dispatcher)
	if err := server.Start(); err != nil {
		slog.Error("cannot listen", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	errCh := make(chan error, 1)
	go func() { errCh <- server.Server() }()

	select {
	case <-ctx.Done():
		slog.Info("Shutting down")
		server.Stop()
		<-errCh 
	case err:= <-errCh:
		if err != nil {
			slog.Error("server stopped", "err",err)
			os.Exit(1)
		}
	}
	slog.Info("server stopped")
}
