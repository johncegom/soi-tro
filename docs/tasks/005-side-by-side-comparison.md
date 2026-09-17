# Task: Add side-by-side rental comparison

Historical task, migrated from legacy Feature Plan Task 3.

## Definition of Done

- [x] Let the user select two or three rentals from saved history.
- [x] Render the selected rentals side by side in the terminal.
- [x] Handle insufficient history and invalid selections without data loss.
- [x] Cover the supporting UI behavior with unit tests.

## Test Plan

- Automated: `go test ./internal/ui ./internal/database`.
- Manual: save at least three rentals, compare two and then three records, and
  verify that values align under the correct rental.

## Program design

- Types/interfaces: comparison consumes stored `database.RentalRecord` values.
- Data flow: history query -> interactive selection -> comparison renderer.
- Package/file layout: `internal/database` owns records and `internal/ui` owns
  selection and terminal rendering.

## Notes and deviations

Implemented with its SQLite prerequisite in commit `b330cd5`. This task
predates the current workflow; its criteria are reconstructed from the
finished code and old plan.
