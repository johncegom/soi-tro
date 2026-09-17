# Task: Add batch directory processing

Proposed task migrated from legacy Feature Plan Task 2. It is not approved for
implementation.

## Definition of Done

- [ ] Accept a directory containing `.jpg`, `.jpeg`, `.png`, and `.webp` files.
- [ ] Process supported images with bounded concurrency.
- [ ] Make the concurrency limit explicit and safe for the configured Gemini
  API quota.
- [ ] Report per-file success and failure without discarding successful results.
- [ ] Persist successful analyses through the existing history path.
- [ ] Avoid logging or exposing API keys and raw personal contact information.

## Test Plan

- Automated: unit-test file discovery, filtering, concurrency bounds, partial
  failures, cancellation, and result aggregation with a fake Gemini boundary;
  run `go test -race ./...` without live API calls.
- Manual: process a mixed temporary directory and confirm unsupported files are
  skipped, progress is understandable, and successful results enter history.

## Program design

- Types/interfaces: introduce a narrow analyzer interface only if needed to
  replace the live Gemini boundary in tests; define a per-file result carrying
  its source path and error.
- Key signatures: finalize after the Gemini API quota and desired CLI command
  shape are confirmed.
- Data flow: directory scan -> supported-file queue -> bounded workers ->
  per-file results -> history/export and terminal summary.
- Package/file layout: discovery and scheduling should not live in the UI;
  preserve `internal/gemini` as the external API boundary.

## Notes and deviations

Before implementation, confirm the intended Gemini quota/concurrency limit and
whether the CLI should stop on cancellation or finish already-started files.
