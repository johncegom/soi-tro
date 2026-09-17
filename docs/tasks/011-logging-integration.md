# Task: Finish structured-logging integration

Proposed continuation of legacy Feature Plan Task 6. Core logging exists, but
this remaining scope is not approved for implementation.

## Current baseline

- `internal/logger` provides `log/slog` text or JSON output, level selection,
  file plus optional console output, environment configuration, and tests.
- `cmd/main.go` initializes the logger and attaches request IDs at selected
  orchestration points.
- Operational packages do not yet emit structured logs consistently.

## Definition of Done

- [ ] Add structured logs at meaningful database, Gemini, and exporter operation
  boundaries; do not convert normal terminal UI output into diagnostic logs.
- [ ] Use a small documented field vocabulary, including `operation`, `error`,
  `duration_ms`, and `request_id` when one is available.
- [ ] Never log API keys, raw listing bodies, images, schema contents, or personal
  contact details.
- [ ] Propagate correlation context only across call paths that already benefit
  from it, with targeted signature changes and cancellation preserved.
- [ ] Document logger configuration, field conventions, safe/unsafe data, and
  examples in `docs/logging.md`.
- [ ] Cover configuration and new logging behavior without a live Gemini request
  or writes to real user configuration paths.
- [ ] Keep existing CLI behavior and exported report contents unchanged.

## Test Plan

- Automated: add handler-backed assertions for level, fields, request IDs,
  durations, and redaction/omission; run `go test ./...`, `go test -race ./...`,
  `go vet ./...`, and `golangci-lint run`.
- Manual: run the CLI with text and JSON formats, exercise a successful local
  flow and controlled failures, and inspect logs for correlation and sensitive
  data leakage.

## Program design

- Types/interfaces: continue using the standard `*slog.Logger`; do not introduce
  a duplicate logger interface or a custom application-error hierarchy.
- Key signatures: reuse `logger.Get`, `logger.With`, `logger.NewContext`, and
  `logger.GetRequestID`; add a context helper only if implementation shows that
  repeated field attachment cannot remain at orchestration boundaries.
- Data flow: environment -> `logger.Config` -> process-wide logger; operation
  context -> safe structured attributes -> file/optional console handler.
- Package/file layout: `internal/logger` owns setup and correlation helpers;
  domain packages emit operation events; `internal/ui` retains user-facing
  output; `docs/logging.md` owns usage guidance.

## Excluded from this task

- Log rotation, alerting, health endpoints, distributed tracing, resource
  monitoring, colorized output, and third-party aggregation.
- Custom error codes and stack-capturing error types.
- Logging every function call or user interaction.

These items need separate evidence and approval if a reachable product need
emerges. They are not prerequisites for consistent structured diagnostics in a
local CLI.

## Notes and deviations

Core logging landed in `aa17051`; unimplemented rotation configuration and the
unused custom-error package were removed in `6032c22`. This task replaces the
legacy four-phase plan with the smallest remaining scope supported by current
code and the repository's proportionality rule.
