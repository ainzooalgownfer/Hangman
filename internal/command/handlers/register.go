package handlers

import (
	"hangman/internal/command"
	"time"
)

func RegisterAll(d *command.CommandDispatcher, store SessionStore, maxAttempts int,
	timelimit time.Duration, words []string) {
	d.Register("NEW_GAME", &NewGameHandler{Sessions: store, MaxAttempts: maxAttempts, TimeLimit: timelimit, Words: words})
	d.Register("LETTER", &GuessLetterHandler{Sessions: store})
	d.Register("RESUME", &ResumeHandler{Sessions: store})
	d.Register("QUIT", &QuitHandler{Sessions: store})
	d.Register("STATUS", &StatusHandler{Sessions: store})

}

// RegisterRooms adds the room actions (1 vs 1 duel). Solo actions are unchanged.
func RegisterRooms(d *command.CommandDispatcher, rooms RoomStore) {
	d.Register("LIST_ROOMS", &ListRoomsHandler{Rooms: rooms})
	d.Register("CREATE_ROOM", &CreateRoomHandler{Rooms: rooms})
	d.Register("JOIN_ROOM", &JoinRoomHandler{Rooms: rooms})
	d.Register("ROOM_LETTER", &RoomLetterHandler{Rooms: rooms})
	d.Register("LEAVE_ROOM", &LeaveRoomHandler{Rooms: rooms})
}
