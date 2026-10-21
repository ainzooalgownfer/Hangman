package network

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"time"

	"hangman/internal/command"
	"hangman/internal/protocol"
)

// each socket connection is handled in its own goroutine, allowing multiple clients to connect simultaneously

type TCPServer struct {
	address     string
	maxLineSize int
	idleTimeout time.Duration
	listener    net.Listener
	dispatcher  *command.CommandDispatcher
}

func NewTCPServer(address string, maxLineSize int, idleTimeout time.Duration, dispatcher *command.CommandDispatcher) *TCPServer {
	return &TCPServer{
		address:     address,
		maxLineSize: maxLineSize,
		idleTimeout: idleTimeout,
		dispatcher:  dispatcher,
	}
}

func (s *TCPServer) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("impossible d'écouter sur %s: %w", s.address, err)
	}
	log.Printf("TCPServer en écoute sur %s\n", s.address)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		// equivalent to threading.
		go s.handleConnection(conn)
	}
}

func (s *TCPServer) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	client := NewClientConnection(conn, s.maxLineSize, s.idleTimeout)
	defer client.Close()

	remote := conn.RemoteAddr().String()
	slog.Info("client connected", "remote", remote)

	defer slog.Info("client disconnected", "remote", remote)

	for {
		req, err := client.ReadCommand()

		if errors.Is(err, ErrInvalidFormat) {

			slog.Warn("invalid request", "remote", remote, "err", err)
			if err := client.SendResponse(protocol.CommandResponse{
				Type:    "ERROR",
				Code:    "INVALID_FORMAT",
				Message: "Requête JSON invalide.",
			}); err != nil {
				return
			}
			continue
		}
		if err != nil {
			return
		}

		start := time.Now()
		resp := s.dispatcher.Dispatch(req)

		session := resp.SessionID
		if session == "" {
			session = req.SessionID
		}
		slog.Info("request",
			"remote", remote,
			"action", req.Action,
			"session", session,
			"type", resp.Type,
			"status", resp.Status,
			"code", resp.Code,
			"duration", time.Since(start),
		)

		if resp.SessionID != "" {
			client.SetSessionID(resp.SessionID)
		}

		if err := client.SendResponse(resp); err != nil {
			log.Printf("Erreur d'envoi client: %v", err)
			return
		}

		if resp.Type == "BYE" {
			return
		}
	}
}
