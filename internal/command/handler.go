package command

import "hangman/internal/protocol"

type CommandHandler interface {
	Handle(req protocol.CommandRequest) protocol.CommandResponse
}
