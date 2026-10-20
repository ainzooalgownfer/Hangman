package network

import (
	"errors"
	"fmt"
	"log"
	"net"

	"hangman/internal/command"
	"hangman/internal/protocol"
)

// each socket connection is handled in its own goroutine, allowing multiple clients to connect simultaneously

type TCPServer struct {
	address    string
	listener   net.Listener
	dispatcher *command.CommandDispatcher
}

func NewTCPServer(address string, dispatcher *command.CommandDispatcher) *TCPServer {
	return &TCPServer{
		address:    address,
		dispatcher: dispatcher,
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
	client := NewClientConnection(conn)
	defer client.Close()

	for {
		req, err := client.ReadCommand()
		// FIX (bug 1): a malformed line used to close the connection silently.
		// Now we answer with an ERROR and keep reading; the client can retry.
		if errors.Is(err, ErrInvalidFormat) {
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
			return // Conn closed by client, broken pipe, idle timeout or line too long
		}

		resp := s.dispatcher.Dispatch(req)

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
