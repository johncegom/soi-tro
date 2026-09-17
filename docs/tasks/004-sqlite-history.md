# Task: Add local SQLite history

Historical task, migrated from legacy Feature Plan Task 1.

## Definition of Done

- [x] Persist each successful rental analysis in a local SQLite database.
- [x] Initialize the database in the user's Soi Trọ configuration directory.
- [x] Expose history operations needed by the CLI.
- [x] Cover database behavior with unit tests.

## Test Plan

- Automated: `go test ./internal/database`.
- Manual: analyze a listing, restart the CLI, and confirm the saved result is
  available from history.

## Program design

- Types/interfaces: `database.RentalRecord` represents a stored analysis.
- Data flow: successful analysis -> database insert -> history query -> CLI UI.
- Package/file layout: persistence is isolated in `internal/database`; CLI
  orchestration remains in `cmd`, and rendering remains in `internal/ui`.

## Notes and deviations

Implemented in commit `b330cd5`. This task predates the current workflow; its
criteria and design are reconstructed from the finished code and old plan.
