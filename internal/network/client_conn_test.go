package network

import (
	"bufio"
	"errors"
	"net"
	"testing"
	"time"
)

func TestReadCommand_Valid(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	defer serverSide.Close()
	defer clientSide.Close()

	go func() {
		clientSide.Write([]byte(`{"action":"NEW_GAME","player":"Bob"}` + "\n"))
	}()

	conn := NewClientConnection(serverSide, 4096, time.Second)
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

	conn := NewClientConnection(serverSide, 4096, time.Second)
	_, err := conn.ReadCommand()

	if err == nil {
		t.Logf("bad Jason")
	}

	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("expected ErrInvalidFormat, got %v", err)
	}

}

func TestReadCommand_TooLong(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	defer serverSide.Close()
	defer clientSide.Close()

	go func() {
		clientSide.Write([]byte(`{"action":"NEW_GAME","player":"a very long name"}` + "\n"))
	}()

	conn := NewClientConnection(serverSide, 16, time.Second)
	_, err := conn.ReadCommand()

	if !errors.Is(err, bufio.ErrTooLong) {
		t.Fatalf("expected bufio.ErrTooLong for a line over the limit, got %v", err)
	}
}
