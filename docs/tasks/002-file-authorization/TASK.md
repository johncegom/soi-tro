# Task: Recalibrate file-change authorization

## Definition of Done

- [x] Explicit implementation requests authorize ordinary in-scope edits.
- [x] Review, diagnosis, explanation, and proposal requests remain read-only.
- [x] High-risk actions still require explicit confirmation.
- [x] Substantial or resumable work retains task planning and verification.
- [x] The ledger and decision log record the recalibration.

## Test Plan

- Automated: run `git diff --check`.
- Manual: confirm the instructions distinguish implementation requests from
  read-only requests.
- Manual: confirm destructive actions, sensitive data, external effects, scope
  expansion, and possible overwrites still require confirmation.
- Manual: search for active instructions that still require approval for every
  file edit.

## Program design

- Types/interfaces: none; this task changes repository working instructions.
- Flow: classify the request, inspect the worktree, perform authorized
  reversible edits, verify them, and summarize the result. Stop for explicit
  confirmation when a listed high-risk condition applies.
- File layout: live instructions remain in `AGENTs.md`; the rationale is
  recorded in `docs/DECISIONS.md`; task status remains in `docs/LEDGER.md`.

## Notes and deviations

This change replaces the universal approval ceremony with risk-based
authorization. It does not weaken the safeguards around credentials, user
data, destructive operations, releases, or external side effects.
