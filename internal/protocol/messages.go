package protocol

type CommandRequest struct {
	Action    string `json:"action"`
	Player    string `json:"player,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Letter    string `json:"letter,omitempty"`
}

type CommandResponse struct {
	Type         string   `json:"type"` // STATE, BYE, ERROR
	SessionID    string   `json:"session_id,omitempty"`
	MaskedWord   string   `json:"masked_word,omitempty"`
	AttemptsLeft *int     `json:"attempts_left,omitempty"`
	TriedLetters []string `json:"tried_letters,omitzero"`
	Status       string   `json:"status,omitempty"` // IN_PROGRESS, WON, LOST, ABANDONED
	SecretWord   string   `json:"secret_word,omitempty"`
	Message      string   `json:"message,omitempty"`
	Code         string   `json:"code,omitempty"` // errors
}
