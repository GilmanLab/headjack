// Package names provides Docker-style random name generation for sessions.
package names

import (
	"fmt"
	"math/rand/v2"
)

// Default word lists for name generation.
var (
	// adjectives contains positive, memorable adjectives.
	adjectives = []string{
		"agile", "bold", "brave", "bright", "calm",
		"clever", "cool", "daring", "eager", "fair",
		"fancy", "fast", "fine", "flying", "friendly",
		"gentle", "glad", "golden", "good", "grand",
		"happy", "honest", "humble", "jolly", "keen",
		"kind", "lively", "lucky", "merry", "mighty",
		"modest", "noble", "patient", "polite", "proud",
		"quick", "quiet", "rapid", "sharp", "shiny",
		"smart", "smooth", "snappy", "speedy", "steady",
		"swift", "tender", "thrifty", "tidy", "vivid",
		"warm", "wise", "witty", "zealous", "zesty",
	}

	// nouns contains animal names that are easy to remember.
	nouns = []string{
		"badger", "bear", "beaver", "bison", "buffalo",
		"cat", "cheetah", "cobra", "coyote", "crane",
		"crow", "deer", "dolphin", "dove", "eagle",
		"falcon", "ferret", "finch", "fox", "gazelle",
		"goat", "goose", "hare", "hawk", "heron",
		"horse", "jaguar", "koala", "leopard", "lion",
		"llama", "lynx", "marten", "mink", "moose",
		"newt", "otter", "owl", "panda", "panther",
		"parrot", "pelican", "pony", "puma", "rabbit",
		"raven", "seal", "shark", "sloth", "snake",
		"sparrow", "stork", "swan", "tiger", "toucan",
		"turtle", "viper", "walrus", "whale", "wolf",
	}
)

// ExistsFn checks if a name already exists.
type ExistsFn func(name string) bool

// Generator creates random session names.
type Generator struct {
	rng *rand.Rand
}

// New creates a Generator with a random source.
// Pass nil to use the default random source.
func New(src rand.Source) *Generator {
	var rng *rand.Rand
	if src != nil {
		rng = rand.New(src)
	}
	return &Generator{rng: rng}
}

// Generate returns a random adjective-noun name (e.g., "happy-panda").
func (g *Generator) Generate() string {
	var adj, noun string
	if g.rng != nil {
		adj = adjectives[g.rng.IntN(len(adjectives))]
		noun = nouns[g.rng.IntN(len(nouns))]
	} else {
		adj = adjectives[rand.IntN(len(adjectives))]
		noun = nouns[rand.IntN(len(nouns))]
	}
	return fmt.Sprintf("%s-%s", adj, noun)
}

// GenerateUnique returns a name that doesn't exist according to existsFn.
// Returns an error if unable to find a unique name after maxAttempts tries.
func (g *Generator) GenerateUnique(existsFn ExistsFn, maxAttempts int) (string, error) {
	if maxAttempts <= 0 {
		maxAttempts = 100
	}

	for i := 0; i < maxAttempts; i++ {
		name := g.Generate()
		if !existsFn(name) {
			return name, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique name after %d attempts", maxAttempts)
}

// TotalCombinations returns the number of possible unique names.
func TotalCombinations() int {
	return len(adjectives) * len(nouns)
}
