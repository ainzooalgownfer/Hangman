package network

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"hangman/internal/protocol"
)

// TestHubPushReachesClient checks that a pushed message arrives on the
// client's side of the connection, as one JSON line with no request_id.
func TestHubPushReachesClient(t *testing.T) {
	serverSide, clientSide := pipe(t)
	hub := NewHub()
	hub.Add(1, NewClientConnection(serverSide, 4096, time.Second))

	// net.Pipe is synchronous: the push blocks until the other side reads,
	// so it runs in its own goroutine while the test reads.
	go hub.Push(1, protocol.CommandResponse{Type: "ROOM", RoomID: "K3J9"})

	_ = clientSide.SetReadDeadline(time.Now().Add(time.Second))
	line, err := bufio.NewReader(clientSide).ReadString('\n')
	if err != nil {
		t.Fatalf("no push received: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatalf("pushed line is not JSON: %q", line)
	}
	if got["type"] != "ROOM" || got["room_id"] != "K3J9" {
		t.Errorf("expected a ROOM for K3J9, got %s", line)
	}
	if strings.Contains(line, "request_id") {
		t.Errorf("a pushed message must not carry a request_id, got %s", line)
	}
}

// TestHubPushUnknownConnection checks that pushing to a connection the hub
// doesn't know (never added, or already gone) does nothing and doesn't panic.
func TestHubPushUnknownConnection(t *testing.T) {
	NewHub().Push(42, protocol.CommandResponse{Type: "ROOM"})
}

// TestHubPushAfterRemove checks that once a connection is removed, pushes to
// it are dropped instead of being written.
func TestHubPushAfterRemove(t *testing.T) {
	serverSide, clientSide := pipe(t)
	hub := NewHub()
	hub.Add(1, NewClientConnection(serverSide, 4096, time.Second))
	hub.Remove(1)

	hub.Push(1, protocol.CommandResponse{Type: "ROOM"})

	_ = clientSide.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	buf := make([]byte, 1)
	if _, err := clientSide.Read(buf); !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Errorf("expected nothing to be written after Remove, got err=%v", err)
	}
}
