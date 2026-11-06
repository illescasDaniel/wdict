package core

import "testing"

// referenceOSA is an unbounded optimal-string-alignment matrix used only as an
// oracle for the banded implementation.
func referenceOSA(a, b []rune) int {
	d := make([][]int, len(a)+1)
	for i := range d {
		d[i] = make([]int, len(b)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			v := min(d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+cost)
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				v = min(v, d[i-2][j-2]+1)
			}
			d[i][j] = v
		}
	}
	return d[len(a)][len(b)]
}

func TestGivenAllShortStringsWhenMeasuringDistanceThenBandedMatchesReference(t *testing.T) {
	// Given every string up to length 4 over an alphabet with an accented rune.
	alphabet := []rune{'a', 'b', 'é'}
	var words [][]rune
	var build func(prefix []rune)
	build = func(prefix []rune) {
		words = append(words, append([]rune(nil), prefix...))
		if len(prefix) == 4 {
			return
		}
		for _, r := range alphabet {
			build(append(prefix, r))
		}
	}
	build(nil)
	// When each pair is measured by the banded and the reference implementations.
	for _, a := range words {
		for _, b := range words {
			want := referenceOSA(a, b)
			got := alternativeDistance(a, b)
			// Then exact distances up to the radius match, and larger ones are reported as too far.
			if want <= alternativeMaxDistance && got != want {
				t.Fatalf("distance(%q,%q)=%d, want %d", string(a), string(b), got, want)
			}
			if want > alternativeMaxDistance && got <= alternativeMaxDistance {
				t.Fatalf("distance(%q,%q)=%d, reference %d is beyond the radius", string(a), string(b), got, want)
			}
		}
	}
}

func TestGivenTranspositionWhenMeasuringDistanceThenCountsOneEdit(t *testing.T) {
	// Given a query with two adjacent letters swapped.
	a, b := []rune("wrods"), []rune("words")
	// When the distance is measured.
	got := alternativeDistance(a, b)
	// Then optimal string alignment counts the swap as a single edit.
	if got != 1 {
		t.Fatalf("distance=%d, want 1", got)
	}
}

func TestGivenLengthGapBeyondRadiusWhenMeasuringDistanceThenReturnsBeyondRadius(t *testing.T) {
	// Given two strings that differ in length by more than the radius.
	a, b := []rune("cat"), []rune("catalogue")
	// When the distance is measured.
	got := alternativeDistance(a, b)
	// Then the result is reported as beyond the radius without scanning the matrix.
	if got <= alternativeMaxDistance {
		t.Fatalf("distance=%d, want beyond %d", got, alternativeMaxDistance)
	}
}
