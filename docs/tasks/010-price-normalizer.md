# Task: Add Vietnamese rental-price normalization

Migrated from legacy Feature Plan Task 5. Implemented as store + use for
sorting/filtering: the normalized value is persisted in a new `price_vnd`
column via a schema migration, per explicit maintainer confirmation.

## Definition of Done

- [x] Normalize agreed Vietnamese rental-price forms such as `4tr5`, `4.5tr`,
  `4500k`, `4.5 triệu`, and `4m5` to integer VND values.
- [x] Reject ambiguous or malformed values instead of silently guessing.
- [x] Preserve the original listing text alongside any normalized value.
- [x] Document supported syntax and rounding/decimal rules.

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

**Resolved:** the maintainer chose store + sorting/filtering. `internal/priceparser`
implements `ParseVND` independent of the `gemini`/`ui` packages, as planned.
`database.SaveRental` calls it and persists the result in a new nullable
`price_vnd INTEGER` column; `InitDB` migrates existing databases that predate
the column via `PRAGMA table_info`/`ALTER TABLE`. `database.RentalRecord`
exposes `PriceVND *int64`, nil when the price could not be parsed.

No sort/filter query or UI was added: the Definition of Done only requires the
value to exist and be storable so a future consumer can use it, and building
an unused `ORDER BY price_vnd` query or UI control now would be speculative.
Add it when a concrete feature (e.g. a "sort by price" view) needs it.

Documentation: `docs/price-normalization.md` covers the supported syntax,
rounding/rejection rules, and the schema migration.
