package command

import (
	"strings"

	"hangman/internal/errs"
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
		return errs.UnknownAction(req.Action)
	}
	return handler.Handle(req)
}

// Handler returns the handler registered for action, if any.
func (d *CommandDispatcher) Handler(action string) (CommandHandler, bool) {
	h, ok := d.handlers[strings.ToUpper(action)]
	return h, ok
}
