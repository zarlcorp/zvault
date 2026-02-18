# 051: Release Prep — Cleanup and Workflow Fix

Status: done
Updated: 2026-02-18
Exit: success

## Acceptance Criteria
- [x] Delete `internal/tui/placeholders.go`
- [x] Remove `TestPlaceholderView` test from `internal/tui/tui_test.go`
- [x] Remove `TestPlaceholderEscNavigatesBack` test from `internal/tui/tui_test.go`
- [x] Verify no remaining references to `placeholderModel` or `newPlaceholder` in any `.go` file
- [x] Add `with: binary-name: zvault` to `.github/workflows/release.yml`
- [x] All existing tests pass after changes (`go test ./...`)
- [x] `go build ./...` succeeds
- [x] `go vet ./...` reports no issues

## Log
- 2026-02-18: Started work. Reading relevant files.
- 2026-02-18: Confirmed only placeholders.go and two tests in tui_test.go reference placeholderModel/newPlaceholder.
- 2026-02-18: Deleted internal/tui/placeholders.go.
- 2026-02-18: Removed TestPlaceholderView and TestPlaceholderEscNavigatesBack from tui_test.go.
- 2026-02-18: Added `with: binary-name: zvault` to .github/workflows/release.yml.
- 2026-02-18: go build ./... — success.
- 2026-02-18: go vet ./... — no issues.
- 2026-02-18: go test ./... — all packages pass.
- 2026-02-18: Verified zero remaining references to placeholderModel/newPlaceholder in .go files.
- 2026-02-18: All acceptance criteria met. Done.
