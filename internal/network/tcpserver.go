package network

import (
	"bufio"
	"errors"
	"fmt"
	"sync/atomic"

	"log/slog"
	"net"
	"os"
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
	nextConnID  atomic.Int64
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
	slog.Info("server listening", "address", s.address)
	return nil

}

func (s *TCPServer) Server() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
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
	defer func() { _ = client.Close() }() // the connection is ending: nothing to do if Close fails

	connID := s.nextConnID.Add(1)
	remote := conn.RemoteAddr().String()
	slog.Info("client connected", "remote", remote)
	defer slog.Info("client disconnected", "remote", remote)

	for {
		req, err := client.ReadCommand()

		if errors.Is(err, ErrInvalidFormat) {

			slog.Warn("invalid request", "remote", remote, "err", err)
			if err := client.SendResponse(protocol.CommandResponse{
				Type:    "ERROR",
				Code:    "INVALID_MESSAGE",
				Message: "Requête JSON invalide.",
			}); err != nil {
				return
			}
			continue
		}
		if err != nil {
			switch {
			case errors.Is(err, os.ErrDeadlineExceeded):
				slog.Warn("idle timeout", "remote", remote)

				_ = client.SendResponse(protocol.CommandResponse{
					Type:    "ERROR",
					Code:    "TIMEOUT",
					Message: "Délai d'inactivité dépassé.",
				})
			case errors.Is(err, bufio.ErrTooLong):
				slog.Warn("message too long", "remote", remote)
				_ = client.SendResponse(protocol.CommandResponse{
					Type:    "ERROR",
					Code:    "INVALID_MESSAGE",
					Message: "Message trop long.",
				})
			}
			return
		}

		start := time.Now()
		req.ConnID = connID
		resp := s.dispatcher.Dispatch(req)
		resp.RequestID = req.RequestID

		session := resp.SessionID
		if session == "" {
			session = req.SessionID
		}
		slog.Info("request",
			"remote", remote,
			"connID", req.ConnID,
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
			slog.Warn("send failed", "remote", remote, "err", err)
			return
		}

		if resp.Type == "BYE" {
			return
		}
	}
}
