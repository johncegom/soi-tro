# Task: Increase core package unit-test coverage

Historical task migrated from `test-plan.md`.

## Definition of Done

- [x] Cover `.env`, global API-key, schema-path, and schema-migration branches
  in `internal/analyzer`.
- [x] Cover exporter directory and file-opening failures.
- [x] Cover Gemini client construction, schema/export configuration failures,
  and extraction failures that occur before a network request.
- [x] Keep the unit suite independent of live Gemini requests and real API keys.

## Test Plan

- Automated: `go test ./internal/analyzer ./internal/exporter ./internal/gemini`.
- Manual: inspect tests to confirm they use temporary paths and stubbed or
  pre-network failure paths instead of user configuration or external services.

## Program design

- Types/interfaces: production APIs were exercised through package tests; no
  test-only production abstraction was required for the original scope.
- Data flow: isolated test fixture -> target function -> result/error assertion;
  filesystem cases use temporary directories.
- Package/file layout: tests remain beside their packages in `internal`.

## Notes and deviations

The initial tests landed in commit `4aa4166` and were expanded in `66b136a`.
The legacy plan's numeric coverage targets were planning estimates, not stable
quality gates, and are not carried forward as permanent acceptance thresholds.
`test_coverage_report.md` is a historical point-in-time report.
