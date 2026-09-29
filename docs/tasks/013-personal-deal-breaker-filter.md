# Task: Add personal deal-breaker filtering

Promoted from [IDEA-001](../IDEAS.md#idea-001---personal-deal-breaker-filter).

## Definition of Done

- [x] Let the user create, view, edit, and clear one local search profile from
  the CLI.
- [x] Support hard requirements for maximum advertised rent, maximum deposit
  months, pets, parking, elevator access, maximum floor, and maximum acceptable
  unknown fields.
- [x] Preserve the previous move-in-cash limit in existing profiles until the
  user explicitly replaces it with a deposit-month limit.
- [x] Classify an analyzed or saved rental as `Shortlist`, `Needs checking`, or
  `Reject` and show a reason for every failed or unknown requirement.
- [x] Treat missing, ambiguous, or unparseable values as unknown rather than a
  pass.
- [x] Keep evaluation deterministic and local; it must not require another
  Gemini request.
- [x] Store the search profile separately from API-key configuration with owner-
  only permissions and atomic replacement.
- [x] Preserve current analysis and history behavior when no profile exists.

## Test Plan

- Automated: table-driven tests for each supported rule, boundary values,
  combined rules, unknown values, invalid profiles, config round trips, and
  classification precedence, exact deposit-month boundaries, legacy profile
  preservation and explicit replacement; run `go test ./...` and `go vet ./...` without live
  API calls. Exercise the shared form wrapper with text cursor movement,
  Yes/No arrows, menu navigation and selection, and Esc back navigation.
- Manual: create a profile, analyze listings that pass, fail, and omit its
  requirements, restart the CLI, and confirm the profile and explanations are
  preserved.

## Program design

- Types/interfaces: `preferences.SearchProfile` has optional integer VND rent
  and deposit-month limits, a legacy cash limit, optional floor and unknown-count limits, and boolean
  requirements for pets, parking, and elevator. `RentalFacts` contains only the
  extracted strings needed for those rules. `Result` has a classification and
  a pass/fail/unknown finding for each configured rule.
- Key signatures: `SearchProfile.Validate() error`,
  `preferences.Evaluate(SearchProfile, RentalFacts) Result`, and
  `preferences.Load/Save/Clear` for the separate profile file. One adapter
  converts either a fresh or saved extraction result to `RentalFacts`.
- Call graph: main-menu profile editor -> validated atomic JSON profile store;
  fresh extraction or saved history -> local fact adapter -> pure evaluator ->
  reason renderer. Evaluation sends no new request to Gemini and writes no
  rental record.
- Package boundaries: `internal/preferences` owns validation, persistence,
  normalization, and rules; `internal/ui` owns forms and display; `cmd` wires
  profile actions and the fresh-analysis flow. SQLite schema stays unchanged.
- Rule semantics: exact rent and deposit-month comparison; unclear
  deposits and conditional amenities are unknown. Legacy profiles retain the
  exact rent-plus-deposit cash rule. An integer-month deposit may be multiplied
  by parsed rent with overflow checks for the informational total. Count unknowns once per
  configured rule; any explicit failure rejects, then an exceeded optional
  unknown limit rejects, then remaining unknowns need checking. Show all
  failed and unknown reasons. An unset unknown limit imposes no threshold.
- Keyboard contract: Esc returns to the previous menu without saving an
  unfinished profile. The form wrapper passes other keys to the focused field:
  Left/Right move the text cursor or toggle Yes/No, Up/Down navigate menus,
  and Enter accepts the field or selected option. Keep the back hint visible
  on every profile editor page. Test these interactions through the wrapper,
  especially when changing its key handling.
- Presentation boundary: keep profile JSON keys and internal classification,
  status, and field codes unchanged. `internal/preferences` returns Vietnamese
  finding reasons as explanatory text, not machine-readable identifiers;
  `internal/ui` renders Vietnamese labels, profile values, forms, and guidance.
  The feature's visible validation errors and form controls use Vietnamese.
  Machine classification and finding codes remain unchanged.
- Deposit-month revision: add `MaxDepositMonths *int` (`max_deposit_months`)
  to `SearchProfile` and retain `MaxMoveInCashVND` only for existing files.
  `Validate` rejects both being set. `Store.Load` never rewrites a profile.
  The editor displays the legacy limit; saving with the new field empty keeps
  it, while entering a month limit explicitly replaces it. `Evaluate` checks
  integer-month deposits directly, and monetary deposits by exact integer
  comparison against a positive parsed monthly rent. Missing/ambiguous inputs
  remain unknown. The previous cash rule still evaluates legacy profiles.
  `Result` carries a derived rent-plus-deposit amount for a clearly labelled
  informational display only; it does not impose a new requirement. UI labels
  the active rule in Vietnamese and warns that the derived sum excludes other
  move-in charges. No DB or Gemini schema change.

## Dependencies and scope

- Monetary rules depend on the deterministic VND values proposed in
  [Task 010](010-price-normalizer.md). Non-monetary rules must still preserve an
  unknown state when the extracted text is ambiguous.
- Commute-time calculation, multiple profiles, preference weighting, and an
  aggregate match score are out of scope for the first version.

## Notes and deviations

Implemented in `internal/preferences` with the CLI editor in `internal/ui`.
The profile lives at `~/.config/soi-tro/search-profile.json` with mode 0600.
History is assessed when viewed; no rental records or database schema change.
New profiles configure maximum deposit months, not a maximum cash total.
The CLI displays rent plus deposit as a reference when both are exact, with a
warning that other upfront charges are excluded. Existing saved cash limits
remain active and visible until the user explicitly enters a month limit.
The manual test that needs a live Gemini request was not run; automated tests
and vet passed.

Do not store the profile by rewriting `config.json`, which contains the Gemini
credential. Avoid an opaque score: a hard failure must remain visible even when
every other field matches.
