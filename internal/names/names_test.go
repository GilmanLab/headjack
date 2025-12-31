package names

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func TestGenerator_Generate(t *testing.T) {
	// Use a fixed seed for deterministic testing
	src := rand.NewPCG(42, 0)
	gen := New(src)

	name := gen.Generate()

	// Verify format: adjective-noun
	parts := strings.Split(name, "-")
	if len(parts) != 2 {
		t.Errorf("expected name with format 'adj-noun', got %q", name)
	}

	// Verify adjective is from the list
	if !contains(adjectives, parts[0]) {
		t.Errorf("adjective %q not in word list", parts[0])
	}

	// Verify noun is from the list
	if !contains(nouns, parts[1]) {
		t.Errorf("noun %q not in word list", parts[1])
	}
}

func TestGenerator_Generate_Deterministic(t *testing.T) {
	// Same seed should produce same sequence
	src1 := rand.NewPCG(123, 0)
	src2 := rand.NewPCG(123, 0)
	gen1 := New(src1)
	gen2 := New(src2)

	for i := 0; i < 10; i++ {
		name1 := gen1.Generate()
		name2 := gen2.Generate()
		if name1 != name2 {
			t.Errorf("iteration %d: expected %q, got %q", i, name1, name2)
		}
	}
}

func TestGenerator_Generate_NilSource(t *testing.T) {
	gen := New(nil)

	// Should not panic and return valid format
	name := gen.Generate()
	parts := strings.Split(name, "-")
	if len(parts) != 2 {
		t.Errorf("expected name with format 'adj-noun', got %q", name)
	}
}

func TestGenerator_GenerateUnique(t *testing.T) {
	src := rand.NewPCG(42, 0)
	gen := New(src)

	existing := make(map[string]bool)
	existsFn := func(name string) bool {
		return existing[name]
	}

	// Generate several unique names
	for i := 0; i < 10; i++ {
		name, err := gen.GenerateUnique(existsFn, 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if existing[name] {
			t.Errorf("generated duplicate name: %s", name)
		}
		existing[name] = true
	}
}

func TestGenerator_GenerateUnique_AllExist(t *testing.T) {
	src := rand.NewPCG(42, 0)
	gen := New(src)

	// Everything exists
	existsFn := func(name string) bool {
		return true
	}

	_, err := gen.GenerateUnique(existsFn, 10)
	if err == nil {
		t.Error("expected error when all names exist")
	}
}

func TestGenerator_GenerateUnique_DefaultMaxAttempts(t *testing.T) {
	src := rand.NewPCG(42, 0)
	gen := New(src)

	existsFn := func(name string) bool {
		return true
	}

	// Pass 0 to use default max attempts
	_, err := gen.GenerateUnique(existsFn, 0)
	if err == nil {
		t.Error("expected error when all names exist")
	}
}

func TestTotalCombinations(t *testing.T) {
	total := TotalCombinations()
	expected := len(adjectives) * len(nouns)
	if total != expected {
		t.Errorf("expected %d combinations, got %d", expected, total)
	}
	// With 55 adjectives and 60 nouns, we should have 3300 combinations
	if total < 1000 {
		t.Errorf("expected at least 1000 combinations for reasonable uniqueness, got %d", total)
	}
}

func TestWordLists_NonEmpty(t *testing.T) {
	if len(adjectives) == 0 {
		t.Error("adjectives list is empty")
	}
	if len(nouns) == 0 {
		t.Error("nouns list is empty")
	}
}

func TestWordLists_NoDuplicates(t *testing.T) {
	adjSet := make(map[string]bool)
	for _, adj := range adjectives {
		if adjSet[adj] {
			t.Errorf("duplicate adjective: %s", adj)
		}
		adjSet[adj] = true
	}

	nounSet := make(map[string]bool)
	for _, noun := range nouns {
		if nounSet[noun] {
			t.Errorf("duplicate noun: %s", noun)
		}
		nounSet[noun] = true
	}
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
