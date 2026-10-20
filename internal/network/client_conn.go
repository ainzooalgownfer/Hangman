package network

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"hangman/internal/protocol"
	"io"
	"net"
	"time"
)

// sentinel error for "the line was read but isn't valid JSON.
var ErrInvalidFormat = errors.New("INVALID_FORMAT")

// limits that protect the server from a single client.
const (
	maxLineSize = 4096
	idleTimeout = 5 * time.Minute
)

type ClientConnection struct {
	conn      net.Conn
	scanner   *bufio.Scanner
	sessionID string
}

func NewClientConnection(conn net.Conn) *ClientConnection {
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 1024), maxLineSize)
	return &ClientConnection{
		conn:    conn,
		scanner: scanner,
	}
}

// ReadCommand reads one line up to '\n' and unmarshals it into CommandRequest
func (c *ClientConnection) ReadCommand() (protocol.CommandRequest, error) {

	// The deadline is reset before every read, so it is an *idle* timeout.
	if err := c.conn.SetReadDeadline(time.Now().Add(idleTimeout)); err != nil {
		return protocol.CommandRequest{}, err
	}

	if !c.scanner.Scan() {
		if err := c.scanner.Err(); err != nil {
			return protocol.CommandRequest{}, err
		}
		return protocol.CommandRequest{}, io.EOF
	}

	line := c.scanner.Bytes()

	var req protocol.CommandRequest
	if err := json.Unmarshal(line, &req); err != nil {
		return protocol.CommandRequest{}, fmt.Errorf("%w: %v", ErrInvalidFormat, err)
	}
	return req, nil
}

func (c *ClientConnection) SendResponse(res protocol.CommandResponse) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(append(data, '\n'))
	return err
}

func (c *ClientConnection) Close() error {
	return c.conn.Close()
}

func (c *ClientConnection) SetSessionID(id string) {
	c.sessionID = id
}
