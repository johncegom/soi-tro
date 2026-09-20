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

- [x] `Close` error is returned on the two write paths: `internal/analyzer/global_config.go` (`SaveGlobalAPIKey`) and `internal/exporter/exporter.go` (`appendToFile`).
- [x] Log file opened 0600 and log dir 0700 in `internal/logger/logger.go`, matching export-file permissions.
- [x] `schemaSecretKey` carries `//nolint:gosec // G101: integrity checksum, not a secret` and a doc comment stating what it protects against (accidental hand-edits). Decision recorded in `docs/DECISIONS.md`.
- [x] `.env.example` committed with `GEMINI_API_KEY=` placeholder.

Phase 2, make lint real:

- [x] `.golangci.yml`: disable `wsl_v5`, `godot`, `paralleltest`; exclude `dupl`, `goconst`, gosec G304/G306 for `_test.go`; exclude G304 everywhere (reading user-named files is the product).
- [x] `golangci-lint run --fix` applied and reviewed for testifylint, staticcheck QF1012, modernize, perfsprint, intrange, gocritic.
- [x] Remaining revive doc comments and the `exhaustive` switch in `internal/ui/forms.go:46` fixed by hand.
- [x] `gocyclo` threshold raised to 20; the three UI renderers over it carry `//nolint:gocyclo` with a reason. `main` is not nolinted.
- [x] `only-new-issues` removed from `lint.yml`; `golangci-lint run ./...` exits 0 locally and in CI.

Phase 3, test what matters:

- [x] Body of `main` extracted into `run(ctx context.Context) error` in `cmd/` (signature deviation, see Program design); dispatch branches have unit tests; `cmd` no longer shows `[no test files]`.
- [x] Tests use `t.Setenv` instead of `os.Setenv` (clears `usetesting`).

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
  - Phase 2: open a PR with a deliberate `errcheck` violation and confirm the Lint job fails. Done 2026-09-20: draft PR #32 (closed unmerged) failed Lint on `Error return value of os.Remove is not checked (errcheck)`, run 35468022472; all other checks passed.

## Sub-tasks

- [x] 0.1 `.gitattributes`, renormalize, gofmt 3 files.
- [x] 0.2 `toolchain` line, pin action, fix gofumpt config key.
- [x] 0.3 Remove stale coverage artifacts and duplicate skill file.
- [x] 1.1 Return `Close` errors on the two write paths.
- [x] 1.2 Log file/dir permissions.
- [x] 1.3 HMAC nolint + doc comment + DECISIONS entry.
- [x] 1.4 `.env.example`.
- [x] 2.1 Shrink `.golangci.yml`.
- [x] 2.2 `--fix` pass, review, commit separately.
- [x] 2.3 Hand fixes (revive, exhaustive).
- [x] 2.4 gocyclo threshold and nolints.
- [x] 2.5 Drop `only-new-issues`; verify CI red on a bad PR, then green.
- [x] 3.1 Extract `run` from `main`, add tests.
- [x] 3.2 `t.Setenv` migration.

## Program design (Phase 3, item 3.1)

Advise ran 2026-09-20 (`docs/eagd-log.md`). Package boundary: `cmd` only; no
change to `internal/*`.

- `type action int` with `actionAnalyze` (zero value, matches today's
  fall-through), `actionHistory`, `actionManage`, `actionExport`,
  `actionModel`, `actionExit`.
- `func dispatch(choice string) action`: pure map from menu value to action;
  unknown value returns `actionAnalyze`.
- `func mimeTypeFor(path string) string`: extracted verbatim from the
  extension switch (`.png`, `.webp`, default `image/jpeg`).
- `func runAnalyze(ctx context.Context, schemaPath string) error`: the body of
  the "analyze" branch moved verbatim (labels and `goto` move with it).
  `log.Fatalf` and `os.Exit(1)` become returned errors.
- `func run(ctx context.Context) error`: env load, DB init, API key, schema,
  then the menu loop switching on `dispatch`. Startup failures return the same
  Vietnamese messages `main` used to `log.Fatalf`.
- `main`: init logger, build ctx, `if err := run(ctx); err != nil { log.Fatal(err) }`.
- Flow: `main` -> `run` -> (`ui.RunFormWithArrows` -> `dispatch`) -> ui handler
  or `runAnalyze`.

Deviations from item 3.1 as written:

- Signature is `run(ctx context.Context) error`, not `(ctx, stdin, stdout)`.
  The item predates reading the UI layer: `huh` and every `ui.*` function use
  `os.Stdin`/`os.Stdout` directly, so the extra parameters would be dead.
- Only `dispatch` and `mimeTypeFor` get unit tests. `run` and `runAnalyze` need
  a TTY and Gemini, so they are not unit-tested; `dispatch` is close to a
  lookup table, so `cmd` coverage is nominal.
- On Gemini client init failure the process still exits non-zero, but through
  `main`'s `log.Fatal` (one extra error line) instead of a bare `os.Exit(1)`.

## Notes and deviations

- Phases are independently shippable; ship each as its own PR.
- 2.1 changes what contributors are held to. Confirm the disable list before
  applying; the rest of the task is mechanical.
- Phase 2 deviations from the plan above: `main` carried a temporary
  `//nolint:gocyclo,funlen` until Phase 3 extracted `run` (removed in 3.1; `runAnalyze` now carries `//nolint:gocyclo`, complexity 25, moved verbatim);
  3.2 (`t.Setenv`) was pulled into Phase 2 because `usetesting` blocked green;
  extra exclusions added on evidence: `funlen`/gosec G104 in tests, testifylint
  `require-error`, dupword ignore for Vietnamese reduplication ("song song",
  "luôn luôn") after `--fix` rewrote user-facing strings.
- errcheck stays on for tests and for `Close`: re-enabling it after the first
  pass exposed `SaveConfig` in `internal/analyzer/engine.go` as a third write
  path with a deferred, unchecked Close. Read-path defers use
  `defer func() { _ = f.Close() }()`.
- Deliberately out of scope: splitting the three UI render functions (cosmetic
  until one has a bug); the duplicate gosec job in `security.yml` (runs
  `-no-fail`, never blocks, free SARIF dashboard, keep); parameterizing SQL
  identifiers in `addColumnIfMissing` (callers are constants).
- PR #31 "Code scanning results / gosec" gate failed on 3 old findings whose lines
  the PR touched (G304 x2, G204). Standalone gosec ignores `//nolint:gosec` and
  `.golangci.yml`, so `security.yml` now passes `-exclude=G304` (same decision as
  the golangci exclusion) and `scripts/release.go` uses `// #nosec G204`. The
  job stays `-no-fail`. G101 at `internal/analyzer/security.go:17` remains an
  open alert (it uses `//nolint`, which gosec ignores); untouched, so it does not
  gate.
