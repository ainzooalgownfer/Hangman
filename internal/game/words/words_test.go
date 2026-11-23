package words

import "testing"

// TestMask checks the masked word the duel shows: same output as the solo game
// (letters separated by spaces, accents kept, non-letters always visible).
func TestMask(t *testing.T) {
	cases := []struct {
		secret string
		tried  []string
		want   string
	}{
		{"chat", nil, "_ _ _ _"},
		{"chat", []string{"C", "T"}, "C _ _ T"},
		{"élève", []string{"E"}, "É _ È _ E"},   // tried letters are base letters: E reveals É and È
		{"porte-clé", nil, "_ _ _ _ _ - _ _ _"}, // the dash is never hidden
	}
	for _, c := range cases {
		if got := Mask(c.secret, c.tried); got != c.want {
			t.Errorf("Mask(%q, %v) = %q, want %q", c.secret, c.tried, got, c.want)
		}
	}
}

// TestCountBase checks how many letters a guess reveals, accents ignored:
// that's the score of a correct letter in a duel.
func TestCountBase(t *testing.T) {
	cases := []struct {
		secret string
		base   rune
		want   int
	}{
		{"élève", 'E', 3},
		{"chat", 'A', 1},
		{"chat", 'Z', 0},
	}
	for _, c := range cases {
		if got := CountBase(c.secret, c.base); got != c.want {
			t.Errorf("CountBase(%q, %q) = %d, want %d", c.secret, c.base, got, c.want)
		}
	}
}
