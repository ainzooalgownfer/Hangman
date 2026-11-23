package protocol

type CommandRequest struct {
	Action    string `json:"action"`
	Player    string `json:"player,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Letter    string `json:"letter,omitempty"`
	RequestID int64  `json:"request_id,omitempty"`
	ConnID    int64  `json:"-"` // set by the server: which connection sent this request
	Sig       string `json:"sig,omitempty"`
	RoomID    string `json:"room_id,omitempty"`
	RoomName  string `json:"room_name,omitempty"`
	Text      string `json:"text,omitempty"` // chat message
}

type CommandResponse struct {
	Type         string   `json:"type"` // STATE, BYE, ERROR
	SessionID    string   `json:"session_id,omitempty"`
	RequestID    int64    `json:"request_id,omitempty"`
	MaskedWord   string   `json:"masked_word,omitempty"`
	AttemptsLeft *int     `json:"attempts_left,omitempty"`
	TriedLetters []string `json:"tried_letters,omitzero"`
	Status       string   `json:"status,omitempty"` // IN_PROGRESS, WON, LOST, ABANDONED
	SecretWord   string   `json:"secret_word,omitempty"`
	Message      string   `json:"message,omitempty"`
	Code         string   `json:"code,omitempty"` // errors
	ResumeToken  string   `json:"resume_token,omitempty"`
	Replayed     bool     `json:"replayed,omitempty"`
	TimeLeftMs   int64    `json:"time_left_ms,omitempty"`
	Reason       string   `json:"reason,omitempty"`

	// Rooms (1 vs 1 )
	RoomID         string        `json:"room_id,omitempty"`
	RoomName       string        `json:"room_name,omitempty"`
	Rooms          []RoomSummary `json:"rooms,omitempty"`   // LIST_ROOMS
	Players        []PlayerView  `json:"players,omitempty"` // ROOM
	You            *int          `json:"you,omitempty"`     // pointers: 0 is a valid index and must be sent
	Turn           *int          `json:"turn,omitempty"`
	TurnTimeLeftMs int64         `json:"turn_time_left_ms,omitempty"`
	Version        int           `json:"version,omitempty"` // goes up on every change
	Result         string        `json:"result,omitempty"`  // WIN, LOSE, DRAW
	From           string        `json:"from,omitempty"`    // CHAT
	Text           string        `json:"text,omitempty"`    // CHAT
}

// RoomSummary is one line of the room list (LIST_ROOMS).
type RoomSummary struct {
	RoomID  string   `json:"room_id"`
	Name    string   `json:"name"`
	Players []string `json:"players"`
	Status  string   `json:"status"` // WAITING, PLAYING, FINISHED
}

// PlayerView is one player as the other side of the room sees them.
type PlayerView struct {
	Name         string `json:"name"`
	Score        int    `json:"score"`
	AttemptsLeft int    `json:"attempts_left"`
	Connected    bool   `json:"connected"`
}
