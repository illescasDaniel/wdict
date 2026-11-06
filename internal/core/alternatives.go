package core

import "context"

// AttemptID identifies one language attempt within a search. The caller assigns
// it before starting each language lookup, including retries, and carries it
// through the corresponding lookup outcome and alternatives result.
type AttemptID uint64

type AlternativesStatus string

const (
	AlternativesReady       AlternativesStatus = "ready"
	AlternativesUnavailable AlternativesStatus = "unavailable"
	AlternativesTimedOut    AlternativesStatus = "timed_out"
)

// AlternativesRequest identifies a confirmed language miss and its attempt.
type AlternativesRequest struct {
	SearchID SearchID
	Attempt  AttemptID
	Query    string
	Language Language
}

// AlternativesResult associates a complete alternatives outcome with the
// originating search and attempt. Ready may contain empty lists. Unavailable
// and timed-out results contain no partial candidate ranking.
type AlternativesResult struct {
	SearchID  SearchID
	Attempt   AttemptID
	Query     string
	Language  Language
	Status    AlternativesStatus
	BaseForms []string
	Spelling  []string
}

// Alternatives finds candidates for one confirmed miss. The caller runs it
// independently of exact lookup and discards results for superseded attempts.
// Parent context cancellation returns an error and no result for delivery;
// an internal time limit returns a timed-out result with no candidates.
// Candidates must be verified against the installed dictionary before return.
type Alternatives interface {
	FindAlternatives(ctx context.Context, request AlternativesRequest) (AlternativesResult, error)
}

// BaseFormSource supplies possible, unverified base forms. It can use a
// morphology dataset or regular-form rules. Implementations honor context
// cancellation and configured work limits; incomplete work returns an error.
type BaseFormSource interface {
	Candidates(ctx context.Context, query string, language Language) ([]string, error)
}

// HeadwordIndex accesses installed dictionary headwords, including synonyms and
// phrases. VerifyHeadword returns the original display form if a candidate is
// present. IterateHeadwords supplies display forms to visit; callback errors
// stop iteration. Both methods honor cancellation and configured work limits;
// incomplete scans return an error rather than appearing complete.
type HeadwordIndex interface {
	VerifyHeadword(ctx context.Context, candidate string, language Language) (headword string, found bool, err error)
	IterateHeadwords(ctx context.Context, language Language, visit func(headword string) error) error
}
