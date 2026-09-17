# Task: Migrate legacy plans into the Tier 2 workflow

## Definition of Done

- [x] Every item in `feature-plan.md` is represented in `docs/LEDGER.md` as
  completed history or proposed work, with a dedicated task document.
- [x] The unfinished logging work in `logging-observability-plan.md` has a
  scoped Definition of Done, Test Plan, program design, and explicit exclusions.
- [x] The completed work described by `test-plan.md` is preserved as historical
  task evidence rather than presented as pending work.
- [x] The three legacy plan files point readers to the authoritative ledger and
  task documents and no longer duplicate mutable task status.
- [x] Internal links and ledger-to-task mappings are valid.

## Test Plan

- Automated: run a repository-local link and mapping check over the changed
  Markdown files; run `git diff --check`.
- Manual: compare all seven legacy feature items and the test plan against the
  new ledger rows and task documents; inspect the final diff for stale status,
  accidental scope changes, and unrelated edits.

## Program design

- Types/interfaces: documentation-only change; task states are represented by
  ledger rows and checklists in task documents.
- Key signatures: not applicable.
- Data flow: legacy plan -> ledger index -> one task document per work item;
  legacy filenames remain as compatibility pointers to the new source of truth.
- Package/file layout: `docs/LEDGER.md` owns status and task discovery;
  `docs/tasks/<NNN>-<slug>.md` owns scope, acceptance criteria, tests, and
  task-specific history.

## Sub-tasks

- [x] 3.1 Classify legacy entries as completed history or proposed work.
- [x] 3.2 Create task records and migrate the logging and test plans.
- [x] 3.3 Replace duplicated legacy plans with compatibility pointers.
- [x] 3.4 Verify mappings, links, and Markdown formatting.

## Notes and deviations

The migration preserves historical implementation references but does not
approve or implement any proposed product feature. Speculative logging features
without a demonstrated current need are retained as exclusions or future
options, not as acceptance criteria.

Verification completed on 2026-09-17: all local Markdown targets and all 11
ledger task references resolve, all seven legacy feature items are mapped,
`git diff --check` passes, and `go test ./...` passes.
