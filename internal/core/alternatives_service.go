package core

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// AlternativesService computes suggestions from the installed dictionary only.
type AlternativesService struct{ Index HeadwordIndex }

type rankedAlternative struct {
	word                                        string
	distance                                    int
	transposition, repeated, initial, lowercase bool
}

func (s *AlternativesService) FindAlternatives(parent context.Context, req AlternativesRequest) (AlternativesResult, error) {
	result := AlternativesResult{SearchID: req.SearchID, Attempt: req.Attempt, Query: req.Query, Language: req.Language, Status: AlternativesUnavailable}
	if err := parent.Err(); err != nil {
		return result, err
	}
	query := strings.TrimSpace(req.Query)
	if query == "" || !utf8.ValidString(req.Query) || strings.IndexFunc(query, unicode.IsControl) >= 0 || utf8.RuneCountInString(query) > 64 || (req.Language != English && req.Language != Spanish && req.Language != French) || s.Index == nil {
		return result, nil
	}
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()
	q := []rune(alternativeFold(query))
	if len(q) == 0 || len(q) > 1021 {
		return result, nil
	}
	isPhrase := strings.Contains(query, " ")
	bases := alternativeBases(query, req.Language)
	foldedBases := make([]string, len(bases))
	foundBases := make([][]string, len(bases))
	const maxBaseVariants = 128
	for i, b := range bases {
		foldedBases[i] = alternativeFold(b)
	}
	best := make([]rankedAlternative, 0, 5)
	// Reused across headwords: zeroing a fresh array per entry is costly.
	var candidateBuffer [1024]rune
	err := s.Index.IterateHeadwords(ctx, req.Language, func(word string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if len(word) > 4096 || !utf8.ValidString(word) {
			return errors.New("invalid or oversized headword")
		}
		// Most lookups have no suffix-derived bases. Avoid Unicode normalization
		// for entries of the wrong phrase kind when only spelling is needed.
		matchingPhrase := strings.Contains(word, " ") == isPhrase
		if len(foldedBases) == 0 {
			if !matchingPhrase {
				return nil
			}
			// Plain ASCII alphanumerics and single spaces do not change length
			// under our folding. Other names must still be normalized.
			if alternativePlainASCII(word) && (len(word) > len(q)+2 || len(word) < len(q)-2) {
				return nil
			}
		}
		wordFold := alternativeFold(word)
		for i, base := range foldedBases {
			if base == wordFold && !containsAlternative(foundBases[i], word) {
				if len(foundBases[i]) >= maxBaseVariants {
					return errors.New("too many base variants")
				}
				foundBases[i] = append(foundBases[i], word)
			}
		}
		if !matchingPhrase {
			return nil
		}
		candidate := candidateBuffer[:0]
		for _, r := range wordFold {
			if len(candidate) >= len(q)+2 {
				return nil
			}
			candidate = append(candidate, r)
		}
		if len(candidate) == 0 || len(candidate) < len(q)-2 {
			return nil
		}
		distance := alternativeDistance(q, candidate)
		if distance == 0 || distance > 2 {
			return nil
		}
		entry := rankedAlternative{word: word, distance: distance, transposition: alternativeTransposition(q, candidate), repeated: alternativeRepeated(q, candidate), initial: q[0] == candidate[0], lowercase: word == strings.ToLower(word)}
		for _, old := range best {
			if old.word == word {
				return nil
			}
		}
		best = append(best, entry)
		sort.Slice(best, func(i, j int) bool { return alternativeLess(best[i], best[j]) })
		if len(best) > 5 {
			best = best[:5]
		}
		return nil
	})
	if err != nil {
		if e := parent.Err(); e != nil {
			return result, e
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			result.Status = AlternativesTimedOut
		}
		return result, nil
	}
	if e := parent.Err(); e != nil {
		return result, e
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		result.Status = AlternativesTimedOut
		return result, nil
	}
	result.Status = AlternativesReady
	for _, variants := range foundBases {
		sort.Strings(variants)
		for _, word := range variants {
			if !containsAlternative(result.BaseForms, word) {
				result.BaseForms = append(result.BaseForms, word)
			}
		}
	}
	for _, r := range best {
		result.Spelling = append(result.Spelling, r.word)
	}
	return result, nil
}

func containsAlternative(words []string, word string) bool {
	for _, v := range words {
		if v == word {
			return true
		}
	}
	return false
}
func alternativeBases(query string, language Language) []string {
	var suffixes []string
	switch language {
	case English:
		suffixes = []string{"s", "es", "ed", "ing"}
	case Spanish:
		suffixes = []string{"s", "es"}
	case French:
		suffixes = []string{"s", "x"}
	default:
		return nil
	}
	query = strings.ToLower(query)
	var out []string
	if language == French && strings.HasSuffix(query, "aux") && len(query) > len("aux") {
		out = append(out, strings.TrimSuffix(query, "aux")+"al")
	}
	for _, suffix := range suffixes {
		if strings.HasSuffix(query, suffix) && len(query) > len(suffix) {
			stem := strings.TrimSuffix(query, suffix)
			if !containsAlternative(out, stem) {
				out = append(out, stem)
			}
			if (suffix == "ed" || suffix == "ing") && !containsAlternative(out, stem+"e") {
				out = append(out, stem+"e")
			}
		}
	}
	return out
}

// alternativeFoldTable caches the slow fold of every single Latin-1 or Latin
// Extended-A rune, so common accented names avoid Unicode normalization.
// foldSimple marks runes whose fold is context-free (no whitespace).
var alternativeFoldTable = func() (t [0x180]string) {
	for r := rune(0); r < 0x180; r++ {
		if unicode.IsSpace(r) || unicode.IsControl(r) || unicode.Is(unicode.Mn, r) || r == 0x85 || r == 0xA0 {
			continue
		}
		t[r] = alternativeFoldSlow(string(r))
		if t[r] == "" && !unicode.IsPunct(r) {
			t[r] = "\x00" // unexpected: force slow path
		}
	}
	return t
}()

// alternativeFoldFast folds names made of single spaces and runes from the
// precomputed table. ok is false when the slow path is required.
func alternativeFoldFast(s string) (string, bool) {
	if len(s) == 0 || s[0] == ' ' || s[len(s)-1] == ' ' {
		return "", false
	}
	var buf [64]byte
	out := buf[:0]
	prevSpace := false
	for i := 0; i < len(s); {
		c := s[i]
		if c < utf8.RuneSelf {
			i++
			if c == ' ' {
				if prevSpace {
					return "", false
				}
				prevSpace = true
				out = append(out, ' ')
				continue
			}
			prevSpace = false
			if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
				out = append(out, c)
				continue
			}
			if c >= 'A' && c <= 'Z' {
				out = append(out, c+'a'-'A')
				continue
			}
			if f := alternativeFoldTable[c]; f != "" && f != "\x00" {
				out = append(out, f...)
				continue
			}
			if unicode.IsPunct(rune(c)) {
				continue
			}
			return "", false
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError || r >= 0x180 {
			return "", false
		}
		f := alternativeFoldTable[r]
		if f == "" || f == "\x00" {
			return "", false
		}
		prevSpace = false
		out = append(out, f...)
		i += size
	}
	return string(out), true
}

func alternativeFold(s string) string {
	if alternativePlainASCII(s) {
		return strings.ToLower(s)
	}
	if f, ok := alternativeFoldFast(s); ok {
		return f
	}
	return alternativeFoldSlow(s)
}

func alternativeFoldSlow(s string) string {
	s = cases.Fold().String(norm.NFD.String(s))
	s = strings.Join(strings.Fields(s), " ")
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) || unicode.IsPunct(r) {
			return -1
		}
		return r
	}, s)
}

// alternativeMaxDistance is the approved spelling-correction radius.
const alternativeMaxDistance = 2

// alternativeDistance computes Unicode optimal string alignment distance
// capped at alternativeMaxDistance+1. It returns the exact distance when that
// distance is at most alternativeMaxDistance; any larger value means "too far".
// It evaluates only the diagonal band that can still reach the limit and stops
// early once a whole row exceeds it, and it allocates rows sized to the input
// so each candidate does not zero fixed-size scratch arrays.
func alternativeDistance(a, b []rune) int {
	const limit = alternativeMaxDistance
	const beyond = limit + 1
	n, m := len(a), len(b)
	if n-m > limit || m-n > limit {
		return beyond
	}
	prevPrev := make([]int, m+1)
	prev := make([]int, m+1)
	cur := make([]int, m+1)
	for j := range prev {
		prev[j] = min(j, beyond)
	}
	for i := 1; i <= n; i++ {
		for j := range cur {
			cur[j] = beyond
		}
		cur[0] = min(i, beyond)
		rowMin := cur[0]
		for j := max(1, i-limit); j <= min(m, i+limit); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			v := prev[j-1] + cost
			if d := prev[j] + 1; d < v {
				v = d
			}
			if d := cur[j-1] + 1; d < v {
				v = d
			}
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				if d := prevPrev[j-2] + 1; d < v {
					v = d
				}
			}
			v = min(v, beyond)
			cur[j] = v
			rowMin = min(rowMin, v)
		}
		if rowMin > limit {
			return beyond
		}
		prevPrev, prev, cur = prev, cur, prevPrev
	}
	return min(prev[m], beyond)
}

func alternativeTransposition(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	first, second := -1, -1
	for i := range a {
		if a[i] != b[i] {
			if first < 0 {
				first = i
			} else if second < 0 {
				second = i
			} else {
				return false
			}
		}
	}
	return first >= 0 && second == first+1 && a[first] == b[second] && a[second] == b[first]
}
func alternativeRepeated(a, b []rune) bool {
	if len(a) < len(b) {
		a, b = b, a
	}
	if len(a) != len(b)+1 {
		return false
	}
	for i := range a {
		if !(i > 0 && a[i] == a[i-1]) && !(i+1 < len(a) && a[i] == a[i+1]) {
			continue
		}
		same := true
		for j := range b {
			k := j
			if j >= i {
				k++
			}
			if a[k] != b[j] {
				same = false
				break
			}
		}
		if same {
			return true
		}
	}
	return false
}
func alternativeLess(a, b rankedAlternative) bool {
	if a.distance != b.distance {
		return a.distance < b.distance
	}
	if a.transposition != b.transposition {
		return a.transposition
	}
	if a.repeated != b.repeated {
		return a.repeated
	}
	if a.initial != b.initial {
		return a.initial
	}
	if a.lowercase != b.lowercase {
		return a.lowercase
	}
	return a.word < b.word
}

func alternativePlainASCII(s string) bool {
	if len(s) == 0 || s[0] == ' ' || s[len(s)-1] == ' ' {
		return false
	}
	prevSpace := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' {
			if prevSpace {
				return false
			}
			prevSpace = true
			continue
		}
		prevSpace = false
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}
