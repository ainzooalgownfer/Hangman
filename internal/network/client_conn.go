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

var ErrInvalidFormat = errors.New("INVALID_FORMAT")

type ClientConnection struct {
	conn        net.Conn
	scanner     *bufio.Scanner
	sessionID   string
	idleTimeout time.Duration
}

func NewClientConnection(conn net.Conn, maxLineSize int, idleTimeout time.Duration) *ClientConnection {
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, min(1024, maxLineSize)), maxLineSize)
	return &ClientConnection{
		conn:        conn,
		scanner:     scanner,
		idleTimeout: idleTimeout,
	}
}

// ReadCommand reads one line up to '\n' and unmarshals it into CommandRequest
func (c *ClientConnection) ReadCommand() (protocol.CommandRequest, error) {

	// The deadline is reset before every read, so it is an *idle* timeout.
	if err := c.conn.SetReadDeadline(time.Now().Add(c.idleTimeout)); err != nil {
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

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.idleTimeout)); err != nil {
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
