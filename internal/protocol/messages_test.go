package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestRoomZeroValuesAreSent checks that a player index or a score of 0 still
func TestRoomZeroValuesAreSent(t *testing.T) {
	zero := 0
	data, err := json.Marshal(CommandResponse{Type: "ROOM", You: &zero, Turn: &zero,
		Players: []PlayerView{{Name: "Samy"}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"you":0`, `"turn":0`, `"score":0`, `"attempts_left":0`, `"connected":false`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("missing %s in %s", want, data)
		}
	}
}
