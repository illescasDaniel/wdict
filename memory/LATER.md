# Later

## Docs
- 2026-11-17 (User) Review the 21 indexed docs in `docs/INDEX.md`: none are verified against code; check them against the code, then run `docs_index.py verify`.

## Suggestions
- Fix the `copylocks` vet finding at `internal/adapters/installer_lifecycle_test.go:273` so `go vet -tags gtk3 ./...` passes.
- Recover the Spanish `dt` labels and nested usage notes that `spanishMeanings` drops (needs its own spec).
- Rank lookup alternatives with word-frequency data (ties are alphabetical, so `dictionaric` beats `dictionary`).
