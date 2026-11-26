// Package errs builds every ERROR response the server sends, so the codes and
// err messages live in one place and stay consistent.
package errs

import (
	"fmt"

	"hangman/internal/protocol"
)

// Error codes, as written in the protocol spec.
const (
	CodeInvalidMessage   = "INVALID_MESSAGE"
	CodeUnknownAction    = "UNKNOWN_ACTION"
	CodeUnknownSession   = "UNKNOWN_SESSION"
	CodeUnknownRoom      = "UNKNOWN_ROOM"
	CodePlayerNotAllowed = "PLAYER_NOT_ALLOWED"
	CodeInvalidLetter    = "INVALID_LETTER"
	CodeAlreadyGuessed   = "ALREADY_GUESSED"
	CodeInvalidState     = "INVALID_STATE"
	CodeTimeout          = "TIMEOUT"
)

// New builds an ERROR response. Prefer the named functions below.
func New(code, message string) protocol.CommandResponse {
	return protocol.CommandResponse{Type: "ERROR", Code: code, Message: message}
}

// Framing and transport (network layer).

func InvalidJSON() protocol.CommandResponse {
	return New(CodeInvalidMessage, "Requête JSON invalide.")
}

func MessageTooLong() protocol.CommandResponse {
	return New(CodeInvalidMessage, "Message trop long.")
}

func Timeout() protocol.CommandResponse {
	return New(CodeTimeout, "Délai d'inactivité dépassé.")
}

// Requests.

func UnknownAction(action string) protocol.CommandResponse {
	return New(CodeUnknownAction, fmt.Sprintf("Action non reconnue : '%s'.", action))
}

func MissingPlayer() protocol.CommandResponse {
	return New(CodeInvalidMessage, "Le champ 'player' est obligatoire.")
}

func MissingSession() protocol.CommandResponse {
	return New(CodeUnknownSession, "Identifiant de session manquant.")
}

func UnknownSession() protocol.CommandResponse {
	return New(CodeUnknownSession, "Session introuvable.")
}

func UnknownRoom() protocol.CommandResponse {
	return New(CodeUnknownRoom, "Salle introuvable.")
}

// Who may play.

func InvalidSignature() protocol.CommandResponse {
	return New(CodePlayerNotAllowed, "Signature invalide.")
}

func NotOwner() protocol.CommandResponse {
	return New(CodePlayerNotAllowed, "Cette partie est jouée depuis une autre connexion.")
}

func ResumeAlreadyUsed() protocol.CommandResponse {
	return New(CodePlayerNotAllowed, "Requête de reprise déjà utilisée.")
}

func RoomFull() protocol.CommandResponse {
	return New(CodePlayerNotAllowed, "Salle complète.")
}

func LeftRoom() protocol.CommandResponse {
	return New(CodePlayerNotAllowed, "Tu as quitté cette salle.")
}

// Letters.

func InvalidLetter() protocol.CommandResponse {
	return New(CodeInvalidLetter, "Une seule lettre alphabétique est attendue.")
}

func AlreadyGuessed() protocol.CommandResponse {
	return New(CodeAlreadyGuessed, "Lettre déjà proposée.")
}

// Game state.

func GameOver() protocol.CommandResponse {
	return New(CodeInvalidState, "La partie est terminée.")
}

func NotPlaying() protocol.CommandResponse {
	return New(CodeInvalidState, "La partie n'est pas en cours.")
}

func NotYourTurn() protocol.CommandResponse {
	return New(CodeInvalidState, "Ce n'est pas ton tour.")
}
