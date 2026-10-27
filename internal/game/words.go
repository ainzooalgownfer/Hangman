package game

import (
	"math/rand/v2"
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
		return base , true
	}
	return base , false
}

func ContainsBase(w string , base rune) bool {
	
	runes := []rune(w)
	for _ , r := range runes {
		if b, _ := BaseLetter(r); b == base {
			return true
		}
	}
	return false
}