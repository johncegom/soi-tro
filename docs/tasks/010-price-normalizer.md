# Task: Add Vietnamese rental-price normalization

Proposed task migrated from legacy Feature Plan Task 5. It is not approved for
implementation.

## Definition of Done

- [ ] Normalize agreed Vietnamese rental-price forms such as `4tr5`, `4.5tr`,
  `4500k`, `4.5 triệu`, and `4m5` to integer VND values.
- [ ] Reject ambiguous or malformed values instead of silently guessing.
- [ ] Preserve the original listing text alongside any normalized value.
- [ ] Document supported syntax and rounding/decimal rules.

## Test Plan

- Automated: table-driven tests for supported forms, whitespace/case variants,
  separators, zero and boundary values, overflow, and ambiguous input; fuzz the
  parser to ensure it never panics.
- Manual: compare representative listing strings with their displayed and
  persisted normalized values.

## Program design

- Types/interfaces: return a VND integer plus an error; avoid floating-point
  arithmetic for currency.
- Key signatures: settle the package and exact function signature when the
  consuming display/query behavior is selected.
- Data flow: extracted price text -> local parser -> normalized value plus
  original text -> display/persistence consumer.
- Package/file layout: keep normalization independent of Gemini and terminal UI
  packages so it can be tested deterministically.

## Notes and deviations

Before implementation, decide whether normalization is display-only, stored in
SQLite, or used for sorting/filtering. A database schema change requires
separate explicit confirmation under `AGENTS.md`.
