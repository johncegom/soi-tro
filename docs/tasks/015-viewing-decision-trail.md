# Task: Add viewing packs and a decision trail

Promoted from [IDEA-003](../IDEAS.md#idea-003---viewing-pack-and-decision-trail).

## Definition of Done

- [ ] Generate a viewing checklist for a saved rental from its missing fields,
  the active search profile when present, and a documented default checklist.
- [ ] Let the user record, edit, and clear an answer for each checklist item.
- [ ] Track a rental as `Unreviewed`, `Viewing planned`, `Considering`,
  `Rejected`, or `Ready to negotiate`, with an optional next-action date.
- [ ] Show listing claims, viewing answers, and unresolved questions separately
  so a contradiction is visible rather than silently overwritten.
- [ ] Draft a follow-up message containing only unresolved questions and copy it
  only after an explicit user action.
- [ ] Persist the decision trail across restarts and retain it when the rental
  history is viewed or compared.
- [ ] Protect free-text notes as personal data: do not include their contents in
  logs, and keep local storage owner-readable only.

## Test Plan

- Automated: test checklist generation, deduplication, answer updates, status
  transitions, next-action dates, contradiction handling, follow-up selection,
  persistence, deletion behavior, and safe logging; run `go test ./...`,
  `go test -race ./...`, and `go vet ./...` without live API calls.
- Manual: create a viewing pack from a listing with missing fields, record mixed
  answers, restart the CLI, confirm the trail remains, and generate a follow-up
  that excludes answered questions.

## Program design

- Types/interfaces: add stable checklist item identifiers, a `ViewingAnswer`, a
  `DecisionStatus` enum, and a `DecisionTrail` linked to a rental ID. Generated
  defaults and user answers must remain separate.
- Key signatures: use a pure checklist boundary shaped like
  `GenerateChecklist(input ChecklistInput) []ChecklistItem`. Define a narrow
  `DecisionTrailRepository` with
  `Get(ctx context.Context, rentalID int64) (DecisionTrail, error)`,
  `Save(ctx context.Context, trail DecisionTrail) error`, and
  `Delete(ctx context.Context, rentalID int64) error`.
- Call graph or event/data flow: saved rental plus optional profile -> checklist
  generator -> history UI editor -> decision repository -> history/comparison
  summaries and follow-up renderer.
- Package/file layout: keep checklist rules and follow-up selection outside
  `internal/ui`; extend `internal/database` only for persistence; keep terminal
  forms and rendering in `internal/ui`.

## Dependencies and scope

- The task can use Task 013 profiles when available but must also work with the
  default checklist alone.
- Photo capture, document storage, reminders outside the running CLI, calendar
  integration, and automated fraud labels are out of scope.
- Persisting this feature requires a SQLite schema migration or new tables. Get
  explicit confirmation immediately before that migration is implemented, as
  required by `AGENTs.md`.

## Notes and deviations

Deleting a rental must not leave orphaned decision data. Decide and document
whether deletion cascades or requires a separate confirmation when the database
design is finalized.
