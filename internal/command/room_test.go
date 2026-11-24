package command_test

import (
	"sync"
	"testing"

	"hangman/internal/auth"
	"hangman/internal/command"
	"hangman/internal/command/handlers"
	"hangman/internal/protocol"
	"hangman/internal/room"
)

// recordPusher stands in for the network Hub: it keeps what would be pushed.
type recordPusher struct {
	mu   sync.Mutex
	sent map[int64][]protocol.CommandResponse
}

func (p *recordPusher) Push(connID int64, msg protocol.CommandResponse) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sent[connID] = append(p.sent[connID], msg)
}

func (p *recordPusher) last(connID int64) protocol.CommandResponse {
	p.mu.Lock()
	defer p.mu.Unlock()
	msgs := p.sent[connID]
	if len(msgs) == 0 {
		return protocol.CommandResponse{}
	}
	return msgs[len(msgs)-1]
}

// setupRooms is a dispatcher with the solo actions and the room actions.
func setupRooms() (*command.CommandDispatcher, *recordPusher) {
	d := setupDispatcher()
	p := &recordPusher{sent: make(map[int64][]protocol.CommandResponse)}
	handlers.RegisterRooms(d, room.NewManager(p, 6, 0, []string{"chat"}))
	return d, p
}

// roomSigned signs a room request the way the client will: room_id in the
// first field, the player's own resume_token as the key.
func roomSigned(req protocol.CommandRequest, token string) protocol.CommandRequest {
	req.Sig = auth.Sign(token, auth.Message(req.RoomID, req.RequestID, req.Action, req.Letter))
	return req
}

// TestDuelThroughDispatcher plays a duel through Dispatch, the same path a
// real client takes, with the error cases along the way.
func TestDuelThroughDispatcher(t *testing.T) {
	d, p := setupRooms()

	// Samy creates a room.
	if r := d.Dispatch(protocol.CommandRequest{Action: "CREATE_ROOM", ConnID: 1}); r.Code != "INVALID_MESSAGE" {
		t.Errorf("CREATE_ROOM without a player: expected INVALID_MESSAGE, got %+v", r)
	}
	created := d.Dispatch(protocol.CommandRequest{Action: "CREATE_ROOM", Player: "Samy", ConnID: 1})
	if created.Type != "ROOM" || created.Status != "WAITING" || created.ResumeToken == "" {
		t.Fatalf("CREATE_ROOM: expected ROOM WAITING with a token, got %+v", created)
	}
	id, samy := created.RoomID, created.ResumeToken

	// Mohamed lists the rooms and joins Samy's.
	list := d.Dispatch(protocol.CommandRequest{Action: "LIST_ROOMS"})
	if list.Type != "ROOMS" || len(list.Rooms) != 1 || list.Rooms[0].RoomID != id {
		t.Fatalf("LIST_ROOMS: expected Samy's room, got %+v", list)
	}
	if r := d.Dispatch(protocol.CommandRequest{Action: "JOIN_ROOM", Player: "Mohamed", RoomID: "NOPE", ConnID: 2}); r.Code != "UNKNOWN_ROOM" {
		t.Errorf("JOIN_ROOM on an unknown room: expected UNKNOWN_ROOM, got %+v", r)
	}
	joined := d.Dispatch(protocol.CommandRequest{Action: "JOIN_ROOM", Player: "Mohamed", RoomID: id, ConnID: 2})
	if joined.Status != "PLAYING" || joined.ResumeToken == "" {
		t.Fatalf("JOIN_ROOM: expected PLAYING with a token, got %+v", joined)
	}
	mohamed := joined.ResumeToken
	if pushed := p.last(1); pushed.Status != "PLAYING" {
		t.Errorf("Samy must be told the duel started, got %+v", pushed)
	}

	letter := func(l string, requestID int64, connID int64) protocol.CommandRequest {
		return protocol.CommandRequest{Action: "ROOM_LETTER", RoomID: id, Letter: l, RequestID: requestID, ConnID: connID}
	}

	// The signature says who is playing: Mohamed can't play on Samy's turn.
	if r := d.Dispatch(roomSigned(letter("c", 1, 2), mohamed)); r.Code != "INVALID_STATE" {
		t.Errorf("Mohamed on Samy's turn: expected INVALID_STATE, got %+v", r)
	}
	if r := d.Dispatch(letter("c", 1, 1)); r.Code != "PLAYER_NOT_ALLOWED" {
		t.Errorf("unsigned ROOM_LETTER: expected PLAYER_NOT_ALLOWED, got %+v", r)
	}
	tampered := roomSigned(letter("c", 1, 1), samy)
	tampered.Letter = "h"
	if r := d.Dispatch(tampered); r.Code != "PLAYER_NOT_ALLOWED" {
		t.Errorf("letter changed after signing: expected PLAYER_NOT_ALLOWED, got %+v", r)
	}
	if r := d.Dispatch(roomSigned(letter("12", 1, 1), samy)); r.Code != "INVALID_LETTER" {
		t.Errorf("two characters: expected INVALID_LETTER, got %+v", r)
	}
	nope := roomSigned(letter("c", 1, 1), samy)
	nope.RoomID = "NOPE"
	if r := d.Dispatch(nope); r.Code != "UNKNOWN_ROOM" {
		t.Errorf("ROOM_LETTER on an unknown room: expected UNKNOWN_ROOM, got %+v", r)
	}

	// Samy plays C: his reply, and a push to Mohamed.
	played := d.Dispatch(roomSigned(letter("c", 1, 1), samy))
	if played.Type != "ROOM" || played.Players[0].Score != 1 || played.Turn == nil || *played.Turn != 1 {
		t.Fatalf("Samy plays C: expected score 1 and Mohamed's turn, got %+v", played)
	}
	if pushed := p.last(2); pushed.MaskedWord != "C _ _ _" {
		t.Errorf("Mohamed must see C, got %+v", pushed)
	}

	// Samy leaves: Mohamed wins by forfeit.
	leave := protocol.CommandRequest{Action: "LEAVE_ROOM", RoomID: id, RequestID: 2, ConnID: 1}
	if r := d.Dispatch(roomSigned(leave, samy)); r.Type != "BYE" {
		t.Fatalf("LEAVE_ROOM: expected BYE, got %+v", r)
	}
	if pushed := p.last(2); pushed.Status != "FINISHED" || pushed.Result != "WIN" {
		t.Errorf("Mohamed must win by forfeit, got %+v", pushed)
	}

	// Mohamed leaves too: the room is gone from the list.
	leave2 := protocol.CommandRequest{Action: "LEAVE_ROOM", RoomID: id, RequestID: 1, ConnID: 2}
	d.Dispatch(roomSigned(leave2, mohamed))
	if r := d.Dispatch(protocol.CommandRequest{Action: "LIST_ROOMS"}); len(r.Rooms) != 0 {
		t.Errorf("a room both players left must be removed, still listed: %+v", r.Rooms)
	}
}
