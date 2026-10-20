package handlers

import "hangman/internal/game"

type SessionStore interface {
	Add(s *game.GameSession)
	Get(id string) (*game.GameSession, bool)
	Remove(id string)
}
