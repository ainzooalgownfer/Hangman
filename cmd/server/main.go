//nolint:errcheck
package main

import (
	"context"
	"crypto/tls"
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

	// nil = no TLS.
	var tlsConfig *tls.Config
	if cfg.TLSCert != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
		if err != nil {
			slog.Error("failed to load TLS certificate", "err", err)
			os.Exit(1)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
	}

	slog.Info("config loaded",
		"address", cfg.Address,
		"tls", tlsConfig != nil,
		"max_attempts", cfg.MaxAttempts,
		"game_limit_time", cfg.GameTimeLimit,
		"idle_timeout", cfg.IdleTimeout,
		"session_ttl", cfg.SessionTTL,
		"words", len(cfg.Words),
	)

	store := game.NewSessionManager()
	dispatcher := command.NewCommandDispatcher()
	handlers.RegisterAll(dispatcher, store, cfg.MaxAttempts, cfg.GameTimeLimit, cfg.Words)

	stop := store.StartJanitor(cfg.SessionTTL, 30*time.Second)
	defer stop()

	server := network.NewTCPServer(cfg.Address, cfg.MaxLineSize, cfg.IdleTimeout, dispatcher, tlsConfig)
	if err := server.Start(); err != nil {
		slog.Error("cannot listen", "err", err)
		os.Exit(1)
	}

	//gracefull shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	errCh := make(chan error, 1)
	go func() { errCh <- server.Server() }()

	select {
	case <-ctx.Done():
		slog.Info("Shutting down")
		server.Stop() // we're exiting anyway
		<-errCh       // waits for that return, so we don't exit while the server is still mid-shutdown
	case err := <-errCh:
		if err != nil {
			slog.Error("server stopped", "err", err)
			os.Exit(1)
		}
	}
	slog.Info("server stopped")
}
