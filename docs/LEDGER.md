# Task Ledger

Current status: the EAGD mechanism and risk-based file authorization are
installed. Legacy plans have been normalized into this ledger; proposed items
remain unapproved until the maintainer selects them for implementation.

| # | Task | Status | Detail |
|---|------|--------|--------|
| 001 | Install EAGD role-spawn mechanism | Complete | `docs/tasks/001-bootstrap-eagd.md` |
| 002 | Recalibrate file-change authorization | Complete | `docs/tasks/002-file-authorization.md` |
| 003 | Migrate legacy plans into the Tier 2 workflow | Complete | `docs/tasks/003-migrate-legacy-plans.md` |
| 004 | Add local SQLite history | Complete | `docs/tasks/004-sqlite-history.md` |
| 005 | Add side-by-side rental comparison | Complete | `docs/tasks/005-side-by-side-comparison.md` |
| 006 | Increase core package unit-test coverage | Complete | `docs/tasks/006-core-test-coverage.md` |
| 007 | Add CI quality and security gates | Complete | `docs/tasks/007-ci-quality-security.md` |
| 008 | Add batch directory processing | Proposed | `docs/tasks/008-batch-processing.md` |
| 009 | Add customizable message templates | Proposed | `docs/tasks/009-message-templates.md` |
| 010 | Add Vietnamese rental-price normalization | Complete | `docs/tasks/010-price-normalizer.md` |
| 011 | Finish structured-logging integration | Complete | `docs/tasks/011-logging-integration.md` |
| 012 | Flatten task-document layout | Complete | `docs/tasks/012-flatten-task-docs.md` |
| 013 | Add personal deal-breaker filtering | Proposed | `docs/tasks/013-personal-deal-breaker-filter.md` |
| 014 | Calculate true monthly and move-in costs | Proposed | `docs/tasks/014-true-rental-cost.md` |
| 015 | Add viewing packs and a decision trail | Proposed | `docs/tasks/015-viewing-decision-trail.md` |
| 016 | Make UI field ordering deterministic and testable | Complete | `docs/tasks/016-ui-field-ordering-testability.md` |
| 017 | Make lint, format and security scans trustworthy | Complete | `docs/tasks/017-lint-format-security-hardening.md` |

## Resume checklist

1. Read `AGENTS.md`.
2. Read this ledger.
3. Open only the task detail needed for the current work.
4. Check `docs/BUGS.md` and `docs/DECISIONS.md` when they affect that task.
5. Confirm that substantial or resumable work has a Definition of Done and
   Test Plan before implementation.
