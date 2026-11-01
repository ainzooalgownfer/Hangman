package game

import (
	"crypto/rand"
	"slices"
	"strings"
	"sync"
	"time"

	"hangman/internal/game/state"
	"hangman/internal/game/words"
	"hangman/internal/protocol"
)

type GameSession struct {
	mu            sync.Mutex
	id            string
	player        string
	secretWord    string
	maskedWord    string
	attemptsLeft  int
	triedLetters  []string
	currentState  state.GameState
	lastActive    time.Time
	resumeToken   string
	lastRequestID int64
	lastResponse  protocol.CommandResponse
	owner         int64 // the connection currently allowed to play this game

}

// GameSession is what the states act on, through the state.StateSession interface.
var _ state.StateSession = (*GameSession)(nil)

func NewGameSession(id, player, secretWord string, maxAttempts int) *GameSession {
	session := &GameSession{
		id:           id,
		resumeToken:  rand.Text(),
		player:       player,
		secretWord:   strings.ToUpper(secretWord),
		attemptsLeft: maxAttempts,
		triedLetters: make([]string, 0),
		currentState: &state.InProgressState{},
		lastActive:   time.Now(),
	}

	session.UpdateMaskedWord()

	return session
}

func (s *GameSession) Guess(letter string) protocol.CommandResponse {
	return s.GuessOnce(0, 0, letter)
}

func (s *GameSession) GuessOnce(requestID, connID int64, letter string) protocol.CommandResponse {
	s.mu.Lock()
	defer s.mu.Unlock()
	if connID != s.owner {
		return notOwnerError()
	}
	if resp, ok := s.replay(requestID); ok {
		return resp
	}
	s.lastActive = time.Now()
	resp := s.currentState.Guess(s, letter)
	s.remember(requestID, resp)
	return resp
}

func (s *GameSession) Quit() protocol.CommandResponse {
	return s.QuitOnce(0, 0)

}

func (s *GameSession) QuitOnce(requestID, connID int64) protocol.CommandResponse {
	s.mu.Lock()
	defer s.mu.Unlock()

	if connID != s.owner {
		return notOwnerError()
	}
	if resp, ok := s.replay(requestID); ok {
		return resp
	}
	s.lastActive = time.Now()
	resp := s.currentState.Quit(s)
	s.remember(requestID, resp)
	return resp
}

func (s *GameSession) SetState(st state.GameState) {
	s.currentState = st
}

func (s *GameSession) GetState() string {
	return s.currentState.GetStatus()
}

// Snapshot returns the current state for callers outside the session (handlers).
func (s *GameSession) Snapshot() protocol.CommandResponse {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActive = time.Now()

	resp := state.StateResponse(s, s.currentState.GetStatus(), "")
	if resp.Status == "WON" || resp.Status == "LOST" || resp.Status == "ABANDONED" {
		resp.SecretWord = s.secretWord
	}

	return resp
}

func (s *GameSession) ID() string             { return s.id }
func (s *GameSession) ResumeToken() string    { return s.resumeToken }
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
		b, ok := words.BaseLetter(r)
		if !ok {
			parts = append(parts, string(r))
		} else if letterSet[b] {
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

func (s *GameSession) Claim(connID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.owner = connID
}

// helpers
// replay returns the stored answer if this request was already handled. Caller holds s
func (s *GameSession) replay(requestID int64) (protocol.CommandResponse, bool) {
	if requestID == 0 || requestID != s.lastRequestID {
		return protocol.CommandResponse{}, false
	}
	resp := s.lastResponse
	resp.Replayed = true
	return resp, true
}

// remember stores the answer to a request so it can be replayed. Caller holds s.mu.
func (s *GameSession) remember(requestID int64, resp protocol.CommandResponse) {
	if requestID != 0 {
		s.lastRequestID = requestID
		s.lastResponse = resp
	}
}

func notOwnerError() protocol.CommandResponse {
	return protocol.CommandResponse{
		Type:    "ERROR",
		Code:    "PLAYER_NOT_ALLOWED",
		Message: "Cette partie est jouée depuis une autre connexion.",
	}
}
