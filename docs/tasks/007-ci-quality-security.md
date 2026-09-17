# Task: Add CI quality and security gates

Historical task, migrated from legacy Feature Plan Task 7.

## Definition of Done

- [x] Run build, vet, module-drift, race-enabled test, and coverage checks in CI.
- [x] Run `golangci-lint` without making legacy lint debt a blanket merge block.
- [x] Run `govulncheck` and `gosec`, publishing supported security results.
- [x] Configure weekly dependency updates.
- [x] Require the agreed checks on `main` while blocking force-push and deletion.

## Test Plan

- Automated: validate `.github/workflows/*.yml` through a pull request and
  confirm all required checks pass.
- Manual: inspect the repository rules for the required checks and branch
  protections.

## Program design

- Types/interfaces: GitHub Actions workflows are the CI interfaces.
- Data flow: push/pull request -> build/test, lint, and security workflows ->
  required status checks -> merge decision.
- Package/file layout: workflows live in `.github/workflows`; dependency update
  configuration lives in `.github/dependabot.yml`.

## Notes and deviations

Completed in PR #4 and merged in commit `c2b8de8`. During implementation, the
checks exposed module drift, an invalid action tag, and reachable vulnerable
transitive dependencies; those findings were fixed before merge. The existing
release workflow was not changed.
