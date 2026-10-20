package command

import (
	"fmt"
	"strings"

	"hangman/internal/protocol"
)

type CommandDispatcher struct {
	handlers map[string]CommandHandler //name → handler
}

// Constructor, Returns a pointer so all methods operate on the same shared instance.
func NewCommandDispatcher() *CommandDispatcher {
	return &CommandDispatcher{
		handlers: make(map[string]CommandHandler),
	}
}

func (d *CommandDispatcher) Register(action string, handler CommandHandler) {
	d.handlers[strings.ToUpper(action)] = handler
}

func (d *CommandDispatcher) Dispatch(req protocol.CommandRequest) protocol.CommandResponse {
	handler, exists := d.handlers[strings.ToUpper(req.Action)]
	if !exists {
		return protocol.CommandResponse{
			Type:    "ERROR",
			Code:    "UNKNOWN_ACTION",
			Message: fmt.Sprintf("Action non reconnue: '%s'", req.Action),
		}
	}
	return handler.Handle(req)
}
