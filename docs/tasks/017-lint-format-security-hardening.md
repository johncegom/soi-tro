# Task: Make lint, format and security scans trustworthy

Origin: repo scan on 2026-09-19 (`gofmt -l`, `golangci-lint run` 2.13.2,
`govulncheck`, git-history secret grep). Build, vet and tests all pass. The
findings are about tooling that reports without enforcing, not broken code.

Scan baseline:

| Scan | Result |
|---|---|
| `gofmt -l` | 24 files; 21 are CRLF-only, 3 real (`internal/logger/config.go`, `internal/ui/history.go`, `internal/ui/renderer.go`) |
| `golangci-lint run` | 611 issues, 300 non-test. wsl_v5 164, godot 30, gosec 22, revive 17, errcheck 14, gocyclo 8 |
| `govulncheck` | 7 reachable stdlib vulns, all fixed by Go 1.26.4-1.26.6; local toolchain 1.26.3. After Phase 0: 0 reachable |
| secret scan | clean; `.env` untracked, no key patterns in history |

Root causes, not symptoms:

- `lint.yml` sets `only-new-issues: true`, so the 611 baseline never fails a
  PR. The strict config is decoration.
- `core.autocrlf=true` with no `.gitattributes`; working tree mixes CRLF/LF.
- `go.mod` has no `toolchain` line; CI resolves `1.26.x` to 1.26.8, local
  stays on 1.26.3, so local and CI scans disagree.
- `internal/analyzer/security.go:12` treats the HMAC key as a secret. Anyone
  who can edit `~/.config/soi-tro/schema.json` can read the binary and recompute
  the signature. It is an integrity checksum, not a defense.
- `paralleltest` (80) fights tests that mutate process env and a package-level
  `DB`; they cannot run parallel without a rewrite nobody has scheduled.

## Definition of Done

Phase 0, hygiene:

- [x] `.gitattributes` with `* text=auto eol=lf`; tree renormalized; `gofmt -l ./cmd ./internal ./scripts` prints nothing.
- [x] `go.mod` has `toolchain go1.26.8`; `govulncheck ./...` reports zero reachable vulns locally.
- [x] `lint.yml` pins `golangci-lint-action` to the local major/minor (2.13); `.golangci.yml` no longer emits the deprecated `gofumpt.extra-rules` warning.
- [x] Stale `coverage` and `test_coverage_report.md` removed from git; `coverage` added to `.gitignore`; one of the duplicated `git-push/SKILL.md` copies removed.

Phase 1, security and correctness:

- [ ] `Close` error is returned on the two write paths: `internal/analyzer/global_config.go` (`SaveGlobalAPIKey`) and `internal/exporter/exporter.go` (`appendToFile`).
- [ ] Log file opened 0600 and log dir 0700 in `internal/logger/logger.go`, matching export-file permissions.
- [ ] `schemaSecretKey` carries `//nolint:gosec // G101: integrity checksum, not a secret` and a doc comment stating what it protects against (accidental hand-edits). Decision recorded in `docs/DECISIONS.md`.
- [ ] `.env.example` committed with `GEMINI_API_KEY=` placeholder.

Phase 2, make lint real:

- [ ] `.golangci.yml`: disable `wsl_v5`, `godot`, `paralleltest`; exclude `dupl`, `goconst`, gosec G304/G306 for `_test.go`; exclude G304 everywhere (reading user-named files is the product).
- [ ] `golangci-lint run --fix` applied and reviewed for testifylint, staticcheck QF1012, modernize, perfsprint, intrange, gocritic.
- [ ] Remaining revive doc comments and the `exhaustive` switch in `internal/ui/forms.go:46` fixed by hand.
- [ ] `gocyclo` threshold raised to 20; the three UI renderers over it carry `//nolint:gocyclo` with a reason. `main` is not nolinted.
- [ ] `only-new-issues` removed from `lint.yml`; `golangci-lint run ./...` exits 0 locally and in CI.

Phase 3, test what matters:

- [ ] Body of `main` extracted into `run(ctx context.Context, stdin io.Reader, stdout io.Writer) error` in `cmd/`; dispatch branches have unit tests; `cmd` no longer shows `[no test files]`.
- [ ] Tests use `t.Setenv` instead of `os.Setenv` (clears `usetesting`).

## Test Plan

- Automated, after each phase:
  - `go build ./... && go vet ./... && go test ./...`
  - `gofmt -l ./cmd ./internal ./scripts` (empty)
  - `golangci-lint run ./...` (Phase 2 onward: exit 0)
  - `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` (Phase 0 onward: no reachable vulns)
  - `git diff --exit-code go.mod go.sum` after `go mod tidy`
- Manual:
  - Phase 0: `git status` after renormalize shows only line-ending changes; `file cmd/main.go` reports no CRLF.
  - Phase 1: run the app once, confirm `~/.config/soi-tro/` log file is `-rw-------` on Unix; confirm an existing signed `schema.json` still loads.
  - Phase 2: open a PR with a deliberate `errcheck` violation and confirm the Lint job fails.

## Sub-tasks

- [x] 0.1 `.gitattributes`, renormalize, gofmt 3 files.
- [x] 0.2 `toolchain` line, pin action, fix gofumpt config key.
- [x] 0.3 Remove stale coverage artifacts and duplicate skill file.
- [ ] 1.1 Return `Close` errors on the two write paths.
- [ ] 1.2 Log file/dir permissions.
- [ ] 1.3 HMAC nolint + doc comment + DECISIONS entry.
- [ ] 1.4 `.env.example`.
- [ ] 2.1 Shrink `.golangci.yml`.
- [ ] 2.2 `--fix` pass, review, commit separately.
- [ ] 2.3 Hand fixes (revive, exhaustive).
- [ ] 2.4 gocyclo threshold and nolints.
- [ ] 2.5 Drop `only-new-issues`; verify CI red on a bad PR, then green.
- [ ] 3.1 Extract `run` from `main`, add tests.
- [ ] 3.2 `t.Setenv` migration.

## Notes and deviations

- Phases are independently shippable; ship each as its own PR.
- 2.1 changes what contributors are held to. Confirm the disable list before
  applying; the rest of the task is mechanical.
- Deliberately out of scope: splitting the three UI render functions (cosmetic
  until one has a bug); the duplicate gosec job in `security.yml` (runs
  `-no-fail`, never blocks, free SARIF dashboard, keep); parameterizing SQL
  identifiers in `addColumnIfMissing` (callers are constants).
