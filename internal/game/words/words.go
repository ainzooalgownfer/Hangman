//nolint:staticcheck
package words

import (
	"math/rand/v2"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func RandomWord(words []string) string {
	return words[rand.IntN(len(words))]
}

func BaseLetter(r rune) (rune, bool) {

	decomposed := norm.NFD.String(string(r))
	base := []rune(decomposed)[0]
	base = unicode.ToUpper(base)

	if base >= 'A' && base <= 'Z' {
		return base, true
	}
	return base, false
}

func ContainsBase(w string, base rune) bool {

	runes := []rune(w)
	for _, r := range runes {
		if b, _ := BaseLetter(r); b == base {
			return true
		}
	}
	return false
}

// CountBase returns how many letters of w have base as their base letter
// (accents ignored): what a correct guess reveals, so its score in a duel.
func CountBase(w string, base rune) int {
	n := 0
	for _, r := range w {
		if b, _ := BaseLetter(r); b == base {
			n++
		}
	}
	return n
}

// Mask hides the letters of secret not yet tried, letters separated by spaces.
// tried holds base letters ("E" reveals É, È and E); non-letters are always shown.
// Same output as GameSession.UpdateMaskedWord, which solo games keep using,
// whatever the case of secret: revealed letters come out in capitals.
func Mask(secret string, tried []string) string {
	letterSet := make(map[rune]bool)
	for _, l := range tried {
		for _, r := range l {
			letterSet[r] = true
		}
	}

	var parts []string
	for _, r := range secret {
		b, ok := BaseLetter(r)
		if !ok || letterSet[b] {
			parts = append(parts, string(unicode.ToUpper(r))) // é → É, like solo's uppercase words
		} else {
			parts = append(parts, "_")
		}
	}
	return strings.Join(parts, " ")
}
