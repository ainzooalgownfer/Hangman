package game

import "math/rand/v2"

func RandomWord(words []string) string {
	return words[rand.IntN(len(words))]
}
