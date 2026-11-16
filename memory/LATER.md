# Later

## Docs
- (User) docs-1 is installed but not set up for the existing docs (`design/`, `reports/`, `sdd/`); ask before setting it up.

## Suggestions
- Fix the `copylocks` vet finding at `internal/adapters/installer_lifecycle_test.go:273` so `go vet -tags gtk3 ./...` passes.
- Recover the Spanish `dt` labels and nested usage notes that `spanishMeanings` drops (needs its own spec).
- Rank lookup alternatives with word-frequency data (ties are alphabetical, so `dictionaric` beats `dictionary`).
