package game

import (
	"hangman/internal/protocol"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode"
)

type GameSession struct {
	mu           sync.Mutex
	id           string
	player       string
	secretWord   string
	maskedWord   string
	attemptsLeft int
	triedLetters []string
	currentState GameState
	lastActive   time.Time
}

func NewGameSession(id, player, secretWord string, maxAttempts int) *GameSession {
	session := &GameSession{
		id:           id,
		player:       player,
		secretWord:   strings.ToUpper(secretWord),
		attemptsLeft: maxAttempts,
		triedLetters: make([]string, 0),
		currentState: &InProgressState{},
		lastActive:   time.Now(),
	}

	session.UpdateMaskedWord()

	return session
}

func (s *GameSession) Guess(letter string) protocol.CommandResponse {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActive = time.Now()
	return s.currentState.Guess(s, letter)
}

func (s *GameSession) SetState(state GameState) {
	s.currentState = state
}

func (s *GameSession) GetState() string {
	return s.currentState.GetStatus()
}

// Snapshot returns the current state for callers outside the session (handlers).
func (s *GameSession) Snapshot() protocol.CommandResponse {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActive = time.Now()

	return stateResponse(s, s.currentState.GetStatus(), "")
}

func (s *GameSession) ID() string             { return s.id }
func (s *GameSession) Player() string         { return s.player }
func (s *GameSession) SecretWord() string     { return s.secretWord }
func (s *GameSession) MaskedWord() string     { return s.maskedWord }
func (s *GameSession) AttemptsLeft() int      { return s.attemptsLeft }
func (s *GameSession) TriedLetters() []string { return slices.Clone(s.triedLetters) }

func (s *GameSession) DecrementAttempts() {
	if s.attemptsLeft > 0 {
		s.attemptsLeft--
	}
}

func (s *GameSession) AddTriedLetter(letter string) {
	s.triedLetters = append(s.triedLetters, strings.ToUpper(letter))
}

func (s *GameSession) UpdateMaskedWord() {
	letterSet := make(map[rune]bool)
	for _, l := range s.triedLetters {

		for _, r := range l {
			letterSet[r] = true
		}
	}

	var parts []string
	for _, r := range s.secretWord {
		if letterSet[r] || !unicode.IsLetter(r) {
			parts = append(parts, string(r))
		} else {
			parts = append(parts, "_")
		}
	}
	s.maskedWord = strings.Join(parts, " ")
}

func (s *GameSession) IsWordComplete() bool {
	return !strings.Contains(s.maskedWord, "_")
}

func (s *GameSession) IdleSince(now time.Time) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return now.Sub(s.lastActive)
}
