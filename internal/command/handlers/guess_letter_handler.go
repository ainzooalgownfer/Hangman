package handlers

import (
	"strings"

	"hangman/internal/errs"
	"hangman/internal/game/words"
	"hangman/internal/protocol"
)

type GuessLetterHandler struct {
	Sessions SessionStore
}

func (h *GuessLetterHandler) Handle(req protocol.CommandRequest) protocol.CommandResponse {
	if req.SessionID == "" {
		return errs.MissingSession()
	}

	trimmed := strings.TrimSpace(req.Letter)
	runes := []rune(trimmed)
	if len(runes) != 1 {
		return errs.InvalidLetter()
	}

	if _, ok := words.BaseLetter(runes[0]); !ok {
		return errs.InvalidLetter()
	}

	session, ok := h.Sessions.Get(req.SessionID)
	if !ok {
		return errs.UnknownSession()
	}

	if !validSig(session.ResumeToken(), req) {
		return errs.InvalidSignature()
	}

	return session.GuessOnce(req.RequestID, req.ConnID, trimmed)

}
