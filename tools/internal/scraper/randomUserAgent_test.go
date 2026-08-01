package scraper

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func TestRandomUserAgent(t *testing.T) {
	result := randomUserAgent()
	if result == "" {
		t.Error("expected non-empty user agent string")
	}
	if !slices.Contains(userAgents, result) {
		t.Errorf("returned user agent %q is not in the known list", result)
	}
}

// TestRandomUserAgentUniqueness runs 100 iterations; the probability of all
// returning the same value by chance (with ≥2 agents in the list) is
// astronomically small, so a uniform return = a real distribution bug.
func TestRandomUserAgentUniqueness(t *testing.T) {
	if len(userAgents) < 2 {
		t.Skipf("only %d user agents configured; uniqueness is impossible", len(userAgents))
	}
	seen := make(map[string]int, len(userAgents))
	for range 100 {
		seen[randomUserAgent()]++
	}
	if len(seen) < 2 {
		t.Fatalf("100 calls all returned %q; distribution appears broken", userAgents[rand.IntN(len(userAgents))])
	}
}
