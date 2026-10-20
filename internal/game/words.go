package game

import "math/rand/v2"

var words = []string{
	"apple", "brave", "crane", "delta", "eagle",
	"flame", "grape", "honey", "igloo", "joker",
	"kneel", "lemon", "mango", "noble", "ocean",
	"piano", "queen", "river", "stone", "tiger",
	"ultra", "vivid", "whale", "xenon", "yacht",
	"zebra",
}

func RandomWord() string {
	return words[rand.IntN(len(words))]
}
