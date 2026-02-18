# 054: zvault TUI Consistency

Status: done
Updated: 2026-02-18
Exit: success

## Acceptance Criteria
- [x] Update `go.mod` to reference latest `zarlcorp/core`
- [x] Add logo to password screen View using `zstyle.StyledLogo(ZvaultAccent)`
- [x] Add logo to menu screen View using `zstyle.StyledLogo(ZvaultAccent)`
- [x] Refactor `renderHeader` to use `zstyle.RenderHeader`
- [x] Remove local header style variables from `header.go`
- [x] Refactor `renderFooter` to use `zstyle.RenderFooter`
- [x] Remove local footer style variables from `footer.go`
- [x] Change `helpFor` return type to `[]zstyle.HelpPair`
- [x] Separate type selector focus from Name field in secret form (focused == -1 for type)
- [x] Left/right arrows on text input fields move cursor (not cycle type)
- [x] Update `TestSecretFormTabNavigation` for new focus model
- [x] `go test ./...` passes
- [x] `go build ./...` succeeds

## Log
- 2026-02-18: Starting implementation
- 2026-02-18: Updated go.mod: zstyle v0.5.10 -> v0.5.11
- 2026-02-18: Refactored header.go — replaced local renderHeader+styles with zstyle.RenderHeader call
- 2026-02-18: Refactored footer.go — replaced local renderFooter+styles with zstyle.RenderFooter, changed helpEntry to zstyle.HelpPair
- 2026-02-18: Added zarlcorp logo to password.go View (indent+StyledLogo+MutedText pattern)
- 2026-02-18: Added zarlcorp logo to menu.go View (same pattern)
- 2026-02-18: Fixed secret_form.go arrow key bug — type selector is now focused==-1, text inputs at 0+
- 2026-02-18: Updated secret_form_test.go TestSecretFormTabNavigation for new focus model (-1 -> 0 -> 1, shift-tab back)
- 2026-02-18: Fixed bounds check in Update fallthrough for focused==-1 case
- 2026-02-18: go build ./... — success
- 2026-02-18: go test ./... — all packages pass
- 2026-02-18: All acceptance criteria met. Done.
