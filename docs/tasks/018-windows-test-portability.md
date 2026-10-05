# Task: Make the test suite pass on Windows

Fixes [BUG-014](../BUGS.md#bug-014-profile-file-mode-test-cannot-pass-on-windows)
and, with a maintainer decision, [BUG-013](../BUGS.md#bug-013-logger-keeps-its-log-file-open-with-no-way-to-close-it).
Found while verifying Task 015 slice 2; `go test ./...` failed on Windows in
`internal/logger` and `internal/preferences` and passed on Linux.

## Definition of Done

- [x] `TestStoreRoundTripAndClear` skips only its POSIX mode assertion on
  Windows and still checks the round trip and the clear.
- [x] The logger no longer leaves its log file open after `resetLogger`, so
  `TestDefaultLoggingKeepsTerminalClean` passes on Windows (BUG-013 option 1).
- [x] `go test ./...` passes on Windows and on Linux.

## Test Plan

- BUG-013 is test-first: the existing failing test is the red run on Windows;
  record the Windows failure line before changing `internal/logger`.
- BUG-014 is a test-only change (no production behavior); say so in the report.
- Run `go test ./...` and `go vet ./...` on Windows, and
  `docker run golang:1.26 go test ./...` for Linux.

## Notes

Approved 2026-10-05; the maintainer chose to close the handle rather than skip
the check. Deviation: `TestDefaultLoggingKeepsTerminalClean` was reordered (see
BUG-013), because its cleanup order made the directory removal run before
`resetLogger`; its expectations are unchanged.
