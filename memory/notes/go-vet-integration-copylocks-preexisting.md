---
name: go-vet-integration-copylocks-preexisting
description: go vet reports a copylocks finding in installer_lifecycle_test.go:273 that is not from new work
type: gotcha
date: 2026-11-13
status: current
supersedes: 
superseded_by: 
source: 
---
`go vet -tags integration ./internal/adapters` (and, on 2026-11-16, `go vet -tags gtk3 ./...`) fail on a pre-existing lock copy
at `internal/adapters/installer_lifecycle_test.go:273`. The project gate `go tool task test` does not run vet, so it stays
green. Suggested fix: correct the lock copy (small, mechanical) or document the exemption.
