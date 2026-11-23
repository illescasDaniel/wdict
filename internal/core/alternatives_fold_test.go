package core

import (
	"math/rand"
	"testing"
)

func TestGivenMixedNamesWhenFoldingThenFastPathMatchesUnicodeFold(t *testing.T) {
	// given
	alphabet := []rune("aAzZ09 -'.éÉèçÇœŒßñÑ¡«ǆÿŸ́ \t+$_łŁ")
	rng := rand.New(rand.NewSource(1))
	// when / then
	for n := 0; n < 200000; n++ {
		runes := make([]rune, 1+rng.Intn(8))
		for i := range runes {
			runes[i] = alphabet[rng.Intn(len(alphabet))]
		}
		if n < 0x180 {
			runes = []rune{rune(n)}
		}
		word := string(runes)
		if got, want := alternativeFold(word), alternativeFoldSlow(word); got != want {
			t.Fatalf("fold(%q) = %q, want %q", word, got, want)
		}
	}
}
