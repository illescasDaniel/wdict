package core

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGivenInvalidValuesWhenNormalizedThenOnlyValidValueRemains(t *testing.T) {
	// given
	raw := []Pronunciation{
		{Text: ""},
		{Text: "   "},
		{Text: "/" + strings.Repeat("a", 63) + "/"},
		{Text: "/<b>x</b>/"},
		{Text: "/a\x00b/"},
		{Text: "/a\x7fb/"},
		{Text: "/həˈləʊ/"},
	}
	// when
	got := NormalizePronunciations(raw)
	// then
	want := []Pronunciation{{Text: "/həˈləʊ/"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGivenTextAtLengthBoundaryWhenNormalizedThenCountedInRunes(t *testing.T) {
	// given
	atLimit := "/" + strings.Repeat("ʃ", 62) + "/"
	over := "/" + strings.Repeat("ʃ", 63) + "/"
	// when
	got := NormalizePronunciations([]Pronunciation{{Text: over}, {Text: atLimit}})
	// then
	if want := []Pronunciation{{Text: atLimit}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGivenLabelsWhenNormalizedThenInvalidLabelIsBlankedAndTextKept(t *testing.T) {
	// given
	atLimit := strings.Repeat("é", 32)
	raw := []Pronunciation{
		{Text: "/a/", Label: strings.Repeat("é", 33)},
		{Text: "/b/", Label: "<i>x</i>"},
		{Text: "/c/", Label: atLimit},
		{Text: "/d/", Label: "  Northumbria \n"},
		{Text: "/e/", Label: "bad\x01"},
	}
	// when
	got := NormalizePronunciations(raw)
	// then
	want := []Pronunciation{{Text: "/a/"}, {Text: "/b/"}, {Text: "/c/", Label: atLimit}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	again := NormalizePronunciations([]Pronunciation{{Text: "/d/", Label: "  Northumbria \n"}})
	if !reflect.DeepEqual(again, []Pronunciation{{Text: "/d/", Label: "Northumbria"}}) {
		t.Fatalf("got %#v", again)
	}
}

func TestGivenWhitespaceWhenNormalizedThenTrimmedAndCollapsed(t *testing.T) {
	// given
	raw := []Pronunciation{{Text: "  /a \t b/\n", Label: " General   American "}}
	// when
	got := NormalizePronunciations(raw)
	// then
	want := []Pronunciation{{Text: "/a b/", Label: "🇺🇸", LabelTitle: "General American"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGivenKnownAccentLabelsWhenNormalizedThenMappedToFlagsWithOriginalAsTitle(t *testing.T) {
	// given
	raw := []Pronunciation{
		{Text: "/a/", Label: "General American"},
		{Text: "/b/", Label: "Received Pronunciation"},
		{Text: "/c/", Label: "Australia"},
	}
	// when
	got := NormalizePronunciations(raw)
	// then
	want := []Pronunciation{
		{Text: "/a/", Label: "🇺🇸", LabelTitle: "General American"},
		{Text: "/b/", Label: "🇬🇧", LabelTitle: "Received Pronunciation"},
		{Text: "/c/", Label: "🇦🇺", LabelTitle: "Australia"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGivenTextWithoutDelimitersWhenNormalizedThenWrappedInSlashes(t *testing.T) {
	// given
	raw := []Pronunciation{{Text: "ˈola"}, {Text: "[ˈola]"}, {Text: "/ola/"}}
	// when
	got := NormalizePronunciations(raw)
	// then
	want := []Pronunciation{{Text: "/ˈola/"}, {Text: "[ˈola]"}, {Text: "/ola/"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGivenFiveDistinctAndOneRepeatedWhenNormalizedThenFirstThreeDistinctKeptInOrder(t *testing.T) {
	// given
	raw := []Pronunciation{
		{Text: "/a/", Label: "US"}, {Text: "/b/"}, {Text: "/a/", Label: "UK"}, {Text: "/c/"}, {Text: "/d/"}, {Text: "/e/"},
	}
	// when
	got := NormalizePronunciations(raw)
	// then
	want := []Pronunciation{{Text: "/a/", Label: "🇺🇸", LabelTitle: "US"}, {Text: "/b/"}, {Text: "/c/"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGivenInvalidItemsBeforeValidOnesWhenNormalizedThenCapCountsOnlyValidDistinct(t *testing.T) {
	// given
	raw := []Pronunciation{{Text: "<x>"}, {Text: "/a/"}, {Text: "/a/"}, {Text: ""}, {Text: "/b/"}, {Text: "/c/"}, {Text: "/d/"}}
	// when
	got := NormalizePronunciations(raw)
	// then
	if want := []Pronunciation{{Text: "/a/"}, {Text: "/b/"}, {Text: "/c/"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGivenNothingValidWhenNormalizedThenResultIsNil(t *testing.T) {
	// given
	// when
	got := NormalizePronunciations([]Pronunciation{{Text: ""}})
	// then
	if got != nil || NormalizePronunciations(nil) != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestGivenNormalizedValuesWhenNormalizedAgainThenUnchanged(t *testing.T) {
	// given
	once := NormalizePronunciations([]Pronunciation{{Text: "ola", Label: "Received Pronunciation"}, {Text: "[x]"}})
	// when
	twice := NormalizePronunciations(once)
	// then
	if !reflect.DeepEqual(once, twice) {
		t.Fatalf("once %#v, twice %#v", once, twice)
	}
}

func TestGivenPronunciationWithAudioURLWhenNormalizedThenTheDisplayOutputIsUnchangedAndTheURLIsKept(t *testing.T) {
	// given (C1): the URL is metadata; text, label and title behave as before
	// when
	got := NormalizePronunciations([]Pronunciation{
		{Text: "/wɝːd/", Label: "🇺🇸", LabelTitle: "General American", AudioURL: "https://example.test/a.mp3"},
		{Text: "/həˈləʊ/", AudioURL: "https://example.test/b.mp3"},
		{Text: "/[ˈola]/", AudioURL: "https://example.test/c.mp3"},
	})
	// then
	want := []Pronunciation{
		{Text: "/wɝːd/", Label: "🇺🇸", LabelTitle: "General American", AudioURL: "https://example.test/a.mp3"},
		{Text: "/həˈləʊ/", AudioURL: "https://example.test/b.mp3"},
		{Text: "/[ˈola]/", AudioURL: "https://example.test/c.mp3"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
	// and the display output of items without a URL is unchanged
	if got := NormalizePronunciations([]Pronunciation{{Text: "wɜːd"}}); !reflect.DeepEqual(got, []Pronunciation{{Text: "/wɜːd/"}}) {
		t.Fatalf("display changed: %+v", got)
	}
}

func TestGivenInputSliceWhenNormalizedThenInputIsNotModified(t *testing.T) {
	// given
	raw := []Pronunciation{{Text: "  ola ", Label: "General American"}}
	// when
	_ = NormalizePronunciations(raw)
	// then
	if raw[0].Text != "  ola " || raw[0].Label != "General American" {
		t.Fatalf("input changed: %#v", raw)
	}
}

func TestGivenOldShapeCacheRecordWhenLookingUpThenFoundFromCacheWithoutOnlineRequest(t *testing.T) {
	// given
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	record := CacheRecord{Entries: []Entry{{Headword: "hello", Meanings: []Meaning{{Definitions: []string{"A greeting."}}}}}, FetchedAt: now.Add(-24 * time.Hour)}
	online := &fakeSource{found: true, entries: []Entry{{Headword: "online"}}}
	service := LookupService{Memory: &fakeCache{found: true, record: record}, Online: online, Clock: fakeClock{now}}
	// when
	result, err := service.Lookup(context.Background(), LookupRequest{Query: "hello", Language: "en"})
	// then
	if err != nil || result.Status != Found || len(result.Entries) != 1 || result.Entries[0].Headword != "hello" || result.Entries[0].Pronunciations != nil {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if online.calls != 0 {
		t.Fatalf("online calls=%d, want 0", online.calls)
	}
}

func TestGivenSimpleCompoundAndUnknownLabelsWhenNormalizedTwiceThenFlagsWithTitlesAndSameResult(t *testing.T) {
	// given
	labels := []string{"General American", "Received Pronunciation", "UK, Australia, New Zealand", "General American, Canada", "Northumbria", "Early Modern", "UK, Narnia"}
	raw := make([]Pronunciation, 0, len(labels))
	for i, label := range labels {
		raw = append(raw, Pronunciation{Text: "/" + strings.Repeat("a", i+1) + "/", Label: label})
	}
	wantLabels := []string{"🇺🇸", "🇬🇧", "🇬🇧🇦🇺🇳🇿", "🇺🇸🇨🇦", "Northumbria", "Early Modern", "UK, Narnia"}
	wantTitles := []string{"General American", "Received Pronunciation", "UK, Australia, New Zealand", "General American, Canada", "", "", ""}
	// when
	var once, twice []Pronunciation
	for i := 0; i < len(raw); i += 3 {
		end := min(i+3, len(raw))
		batch := NormalizePronunciations(raw[i:end])
		for j, p := range batch {
			if p.Label != wantLabels[i+j] || p.LabelTitle != wantTitles[i+j] {
				t.Fatalf("label %q: got %q title %q, want %q title %q", labels[i+j], p.Label, p.LabelTitle, wantLabels[i+j], wantTitles[i+j])
			}
		}
		once = append(once, batch...)
		twice = append(twice, NormalizePronunciations(batch)...)
	}
	// then
	if !reflect.DeepEqual(once, twice) {
		t.Fatalf("not idempotent: %#v vs %#v", once, twice)
	}
}

func TestGivenAllKnownCountriesCaseInsensitiveWhenNormalizedThenEachBecomesItsFlag(t *testing.T) {
	// given
	cases := map[string]string{"us": "🇺🇸", "UK": "🇬🇧", "australia": "🇦🇺", "NEW ZEALAND": "🇳🇿", " Canada ": "🇨🇦", "Ireland": "🇮🇪", "South Africa": "🇿🇦"}
	for label, flag := range cases {
		// when
		got := NormalizePronunciations([]Pronunciation{{Text: "/a/", Label: label}})
		// then
		if got[0].Label != flag || got[0].LabelTitle != strings.Join(strings.Fields(label), " ") {
			t.Fatalf("%q: got %#v", label, got[0])
		}
	}
}

func TestGivenFlagLabelWithStrayTitleOrUnknownLabelWithTitleWhenNormalizedThenFlagKeepsTitleAndUnknownDropsIt(t *testing.T) {
	// given
	raw := []Pronunciation{
		{Text: "/a/", Label: "🇺🇸", LabelTitle: "General American"},
		{Text: "/b/", Label: "Northumbria", LabelTitle: "stale"},
	}
	// when
	got := NormalizePronunciations(raw)
	// then
	want := []Pronunciation{{Text: "/a/", Label: "🇺🇸", LabelTitle: "General American"}, {Text: "/b/", Label: "Northumbria"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGivenLabelsLongerThanShownLimitWhenNormalizedTwiceThenLimitAppliesToShownLabel(t *testing.T) {
	// given
	long := "UK, Australia, New Zealand, Ireland, South Africa"
	unknown := strings.Repeat("x", 40)
	manyKnown := strings.TrimSuffix(strings.Repeat("UK, ", 17), ", ")
	overRaw := strings.TrimSuffix(strings.Repeat("UK, ", 35), ", ")
	over128 := "UK" + strings.Repeat(", dialectal", 12)                                     // 134 runes, shown form would be 🇬🇧
	exactly128 := "UK" + strings.Repeat(", dialectal", 6) + strings.Repeat(", obsolete", 6) // 2+66+60 runes
	if n := len([]rune(over128)); n != 134 {
		t.Fatalf("over128 has %d runes", n)
	}
	if n := len([]rune(exactly128)); n != 128 {
		t.Fatalf("exactly128 has %d runes", n)
	}
	cases := []struct{ name, label, wantLabel, wantTitle string }{
		{"qualifiers over 128 runes raw but short shown form", over128, "", ""},
		{"qualifiers at exactly 128 runes raw", exactly128, "🇬🇧", exactly128},
		{"known compound over 32 runes raw", long, "🇬🇧🇦🇺🇳🇿🇮🇪🇿🇦", long},
		{"unknown text over 32 runes", unknown, "", ""},
		{"known compound whose flags exceed 32 runes", manyKnown, "", ""},
		{"known compound over 128 runes raw", overRaw, "", ""},
	}
	for _, c := range cases {
		// when
		once := NormalizePronunciations([]Pronunciation{{Text: "/a/", Label: c.label}})
		twice := NormalizePronunciations(once)
		// then
		if once[0].Label != c.wantLabel || once[0].LabelTitle != c.wantTitle {
			t.Fatalf("%s: got %q title %q, want %q title %q", c.name, once[0].Label, once[0].LabelTitle, c.wantLabel, c.wantTitle)
		}
		if !reflect.DeepEqual(once, twice) {
			t.Fatalf("%s: not idempotent: %#v vs %#v", c.name, once, twice)
		}
	}
}

func TestGivenQualifiersAndGenericLabelsWhenNormalizedTwiceThenQualifiersIgnoredAndGenericHasNoLabel(t *testing.T) {
	// given
	cases := []struct{ label, wantLabel, wantTitle string }{
		{"Received Pronunciation, General American, Canada", "", ""},
		{"UK, dialectal", "🇬🇧", "UK, dialectal"},
		{"US, Canada, dialectal", "🇺🇸🇨🇦", "US, Canada, dialectal"},
		{"UK, US", "", ""},
		{"Received Pronunciation, General American", "", ""},
		{"dialectal", "dialectal", ""},
		{"Obsolete, UK", "🇬🇧", "Obsolete, UK"},
		{"UK, dialectal, Narnia", "UK, dialectal, Narnia", ""},
		{"🇬🇧🇺🇸", "", ""},
		{"Australia, New Zealand", "🇦🇺🇳🇿", "Australia, New Zealand"},
	}
	for _, c := range cases {
		// when
		once := NormalizePronunciations([]Pronunciation{{Text: "/a/", Label: c.label}})
		twice := NormalizePronunciations(once)
		// then
		if once[0].Label != c.wantLabel || once[0].LabelTitle != c.wantTitle {
			t.Fatalf("%q: got %q title %q, want %q title %q", c.label, once[0].Label, once[0].LabelTitle, c.wantLabel, c.wantTitle)
		}
		if !reflect.DeepEqual(once, twice) {
			t.Fatalf("%q: not idempotent: %#v vs %#v", c.label, once, twice)
		}
	}
}
