package network

import (
	"errors"
	"net"
	"testing"
)

func TestReadCommand_Valid(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	defer serverSide.Close()
	defer clientSide.Close()

	go func() {
		clientSide.Write([]byte(`{"action":"NEW_GAME","player":"Bob"}` + "\n"))
	}()

	conn := NewClientConnection(serverSide)
	req, err := conn.ReadCommand()

	if err != nil {
		t.Fatalf("unexpected err; %v", err)
	}

	if req.Action != "NEW_GAME" || req.Player != "Bob" {
		t.Errorf("got action=%q player=%q", req.Action, req.Player)
	}
}

func TestReadCommand_InValid(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	defer serverSide.Close()
	defer clientSide.Close()

	go func() {
		clientSide.Write([]byte(`{"action":"NEW_GAME" Bad JSON "player":"Alice" ` + "\n"))
	}()

	conn := NewClientConnection(serverSide)
	_, err := conn.ReadCommand()

	if err == nil {
		t.Logf("bad Jason")
	}

	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("expected ErrInvalidFormat, got %v", err)
	}

}
