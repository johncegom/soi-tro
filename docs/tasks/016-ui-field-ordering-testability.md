# Task: Make UI field ordering deterministic and testable

Fixes [BUG-005](../BUGS.md#bug-005-custom-schema-fields-render-in-nondeterministic-order).
Origin: analysis of `test_coverage_report.md`, which was stale; `internal/ui`
(1.8%) is the only meaningful remaining coverage gap.

## Definition of Done

- [x] One pure helper returns the ordered keys for a schema: given standard keys
  first, then remaining keys sorted, minus an optional skip list. The display
  tables pass the standard fields (`price`, `deposit`, `floor`, `parking_fee`,
  `pets_allowed`, `electricity`, `water`) and skip `missing_fields`,
  `sample_messages`, `additional_notes`, `phone_number`.
- [x] `RenderResults`, `RenderComparisonTable` and `listFields` use it; the
  duplicated ordering and skip-list code is removed.
- [x] A pure `fieldMissing` helper decides `OK` or `THIẾU` for a field from
  required flag, model-missing list and extracted value; `RenderResults` uses it.
- [x] Rendered output is otherwise unchanged (same headings, columns, labels).
- [x] `logger.LogOperationResult` has a direct test for success and failure.

## Test Plan

- Automated, test-first (record the red run per AGENTS.md):
  - ordering test with at least five custom keys, run repeatedly; it must fail
    on assertion before the helper exists;
  - table-driven `fieldMissing` cases: not required, required and present,
    required and empty, `n/a`, `không đề cập`, `chưa đề cập`, and listed as
    missing by the model;
  - `LogOperationResult` with a buffer-backed `slog` handler: info on nil error,
    error level and `failureStage` on failure, no raw error text logged.
  - `go test ./...` and `go vet ./...`, no live Gemini calls.
- Manual: add three custom fields, run analysis and comparison several times,
  and confirm row order is identical across runs. (Not run; the ordering helper
  is unit-tested and all three call sites use it.)

## Program design

- Types/interfaces: none added; two pure functions in `internal/ui`.
- Key signatures: `orderedKeys(props map[string]*genai.Schema, standard []string,
  skip ...string) []string`; `fieldMissing(required bool, missing []string, key,
  value string) bool`.
- Data flow: `LoadSchema` -> `orderedKeys` -> table rows; row status from
  `fieldMissing`. No I/O in either helper.
- Package layout: helpers live in a new `internal/ui/fields.go` with
  `fields_test.go`; `LogOperationResult` test goes in `internal/logger`.

## Sub-tasks

- [x] 1.1 Red test for ordering, then helper, then switch three call sites.
- [x] 1.2 Red table test for `fieldMissing`, then helper, then use in `RenderResults`.
- [x] 1.3 `LogOperationResult` test.

## Notes and deviations

- The helper takes the standard keys and skip list as arguments instead of a
  fixed schema-only signature, because `listFields` deliberately shows every
  property (including `missing_fields`) with its own standard-key order.
- `fieldMissing` treats a whitespace-only value as missing, matching
  `gemini.deriveMissingFields`; the old renderer only caught exact `""`.
- Red evidence: with stubbed helpers, `TestOrderedKeys_*` and `TestFieldMissing`
  (6 subtests) failed on assertions (`expected [price deposit ...] actual
  []string(nil)`). The `LogOperationResult` test passed on first run because it
  characterizes existing behavior (no production change).
- Coverage: `fields.go` and `operation.go` 100%; total 38.6% -> 40.7%.

Deliberately out of scope: `cmd/main`, `scripts/release.go`, and the interactive
`huh` wrappers (`history.go`, `PromptAndSaveModel`, most of `schema_manager.go`);
faking a terminal costs more than it protects. No coverage-percentage gate,
consistent with task 006. `PromptAfterSuccess` option renumbering is a possible
follow-up, not part of this task.
