# Task: Add personal deal-breaker filtering

Promoted from [IDEA-001](../IDEAS.md#idea-001---personal-deal-breaker-filter).

## Definition of Done

- [ ] Let the user create, view, edit, and clear one local search profile from
  the CLI.
- [ ] Support hard requirements for maximum advertised rent, maximum move-in
  cash, pets, parking, elevator access, maximum floor, and maximum acceptable
  unknown fields.
- [ ] Classify an analyzed or saved rental as `Shortlist`, `Needs checking`, or
  `Reject` and show a reason for every failed or unknown requirement.
- [ ] Treat missing, ambiguous, or unparseable values as unknown rather than a
  pass.
- [ ] Keep evaluation deterministic and local; it must not require another
  Gemini request.
- [ ] Store the search profile separately from API-key configuration with owner-
  only permissions and atomic replacement.
- [ ] Preserve current analysis and history behavior when no profile exists.

## Test Plan

- Automated: table-driven tests for each supported rule, boundary values,
  combined rules, unknown values, invalid profiles, config round trips, and
  classification precedence; run `go test ./...` and `go vet ./...` without live
  API calls.
- Manual: create a profile, analyze listings that pass, fail, and omit its
  requirements, restart the CLI, and confirm the profile and explanations are
  preserved.

## Program design

- Types/interfaces: add a `preferences.SearchProfile`, typed optional
  requirements, and an `evaluation.Result` containing a classification plus
  field-level reasons. Keep unknown distinct from false.
- Key signatures: prefer pure entry points shaped like
  `Evaluate(profile SearchProfile, rental RentalFacts) Result`; finalize typed
  normalization boundaries during implementation.
- Call graph or event/data flow: CLI profile editor -> validated atomic profile
  store; extracted or saved rental -> local fact normalization -> rule
  evaluation -> result renderer.
- Package/file layout: a new package should own profile persistence and pure
  evaluation; `internal/ui` should only collect input and render results;
  `cmd` should orchestrate the calls.

## Dependencies and scope

- Monetary rules depend on the deterministic VND values proposed in
  [Task 010](010-price-normalizer.md). Non-monetary rules must still preserve an
  unknown state when the extracted text is ambiguous.
- Commute-time calculation, multiple profiles, preference weighting, and an
  aggregate match score are out of scope for the first version.

## Notes and deviations

Do not store the profile by rewriting `config.json`, which contains the Gemini
credential. Avoid an opaque score: a hard failure must remain visible even when
every other field matches.
