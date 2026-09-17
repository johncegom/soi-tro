# Task: Flatten task-document layout

## Definition of Done

- [x] Single-file tasks use `docs/tasks/<NNN>-<slug>.md`.
- [x] Task directories remain an allowed exception when a task has supporting
  artifacts in addition to its main document.
- [x] `AGENTS.md`, the task template, the ledger, and cross-references describe
  and use the same convention.
- [x] Existing task content and status are preserved during the move.
- [x] No empty per-task directories remain.

## Test Plan

- Automated: check every ledger task reference resolves; scan maintained
  documentation for stale `/TASK.md` references; run `git diff --check`.
- Manual: compare the pre- and post-move task inventory and inspect the diff for
  content loss or unrelated edits.

## Program design

- Types/interfaces: documentation-only layout change.
- Key signatures: not applicable.
- Data flow: ledger task link -> flat task document by default; a task may use a
  directory only when additional task-local artifacts justify it.
- Package/file layout: `docs/tasks/<NNN>-<slug>.md` is the default;
  `docs/tasks/<NNN>-<slug>/TASK.md` is the exception for multi-file tasks.

## Sub-tasks

- [x] 12.1 Update the documented convention and template.
- [x] 12.2 Move existing task documents and update references.
- [x] 12.3 Verify inventory, references, and formatting.

## Notes and deviations

This changes documentation organization only. Product code, task scope, and
task status are not changed.

Verification completed on 2026-09-17: all 12 ledger references resolve to 12
flat task files, all local Markdown links resolve, no task directories remain,
the two previously tracked task files retain identical blob content, and
`git diff --check` passes.
