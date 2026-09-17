# Task: Calculate true monthly and move-in costs

Promoted from [IDEA-002](../IDEAS.md#idea-002---true-monthly-cost-and-move-in-cash).

## Definition of Done

- [ ] Calculate an expected monthly cost range from rent, electricity, water,
  parking, service fees, household size, and user-provided usage assumptions.
- [ ] Calculate the known cash required before moving in, including deposit and
  the first rental payment.
- [ ] Label every component as listing-supplied, user-estimated, or unknown and
  show the assumptions used.
- [ ] Display a lower bound or range when some components vary or are unknown;
  never present an incomplete amount as an exact total.
- [ ] Reject unsupported, ambiguous, negative, or overflowing amounts instead
  of guessing.
- [ ] Calculate locally without sending budget or usage assumptions to Gemini.
- [ ] Expose the calculation for both a newly analyzed rental and a saved
  history record without changing existing records.

## Test Plan

- Automated: table-driven tests for fixed, per-person, per-unit, metered, free,
  ranged, and unknown charges; cover overflow, invalid units, missing inputs,
  deposit expressed in months, provenance labels, and range aggregation; fuzz
  charge parsing and run `go test ./...` plus `go vet ./...` without live API
  calls.
- Manual: calculate costs for representative Vietnamese listings, vary household
  size and usage, confirm all assumptions remain visible, and verify that a
  partially known listing is not shown with an exact total.

## Program design

- Types/interfaces: add typed money and charge values using integer VND, an
  explicit billing unit, provenance, and an unknown state. Add `UsageProfile`
  and a `CostEstimate` with monthly and move-in ranges.
- Key signatures: keep parsing and arithmetic pure, shaped like
  `ParseCharge(text string) (Charge, error)` and
  `Estimate(rental RentalCosts, usage UsageProfile) CostEstimate`.
- Call graph or event/data flow: extracted or saved fields -> deterministic
  charge parsing -> user usage assumptions -> range calculation -> cost
  breakdown renderer.
- Package/file layout: put parsing and arithmetic in a package independent of
  Gemini, SQLite, and terminal UI; let `internal/ui` collect assumptions and
  render the breakdown.

## Dependencies and scope

- This task depends on the price and deposit normalization rules in
  [Task 010](010-price-normalizer.md). It may extend deterministic parsing for
  utility billing units, but must not move arithmetic into Gemini prompts.
- Persisting calculated estimates or changing the SQLite schema is out of scope.
  Estimates should be reproducible from the stored source values and current
  assumptions.

## Notes and deviations

The first version should prefer an honest lower bound over a fabricated average.
Currency calculations must not use floating-point values.
