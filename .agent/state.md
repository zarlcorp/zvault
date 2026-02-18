# Agent State: 049-zvault-tui-tasks (By Tag Filter Fix)

**Status:** done
**Exit:** success
**Started:** 2026-02-18T00:30:00Z
**Updated:** 2026-02-18T00:45:00Z

## Acceptance Criteria
- [x] Add `taskFilterByTag` constant between `taskFilterDone` and `taskFilterCount`
- [x] `taskFilter()` returns `task.Filter{Tag: ...}` when in tag mode
- [x] `filterLabel()` shows `"Filter: #tagname"` for tag mode
- [x] Tab cycles through 4 modes when tags exist
- [x] Tab skips tag mode when no tags exist
- [x] Tag cycling within tag filter mode
- [x] Tests: tab cycling through 4 modes when tags exist
- [x] Tests: tab skipping tag mode when no tags exist
- [x] Tests: tag cycling within tag filter mode
- [x] Tests: filter label showing tag name
- [x] Tests: filtered list only showing tasks with matching tag
- [x] All tests pass (25 task list tests, full suite green, go vet clean)

## Log
- 2026-02-18T00:30:00Z: Started. Read existing code. Plan: add taskFilterByTag constant, update filter/label methods, update tab handler for tag cycling.
- 2026-02-18T00:35:00Z: Added taskFilterByTag constant, updated taskFilter() and filterLabel() methods.
- 2026-02-18T00:37:00Z: Replaced simple modular tab cycling with advanceFilter() method supporting tag cycling and skip-when-empty.
- 2026-02-18T00:38:00Z: Fixed collectTags() bug — reset m.tags to nil before rebuilding to prevent accumulation on repeated loadTasks calls.
- 2026-02-18T00:40:00Z: Added 5 new tests covering all acceptance criteria.
- 2026-02-18T00:45:00Z: All 25 task list tests pass, full suite green, go vet clean. Done.
