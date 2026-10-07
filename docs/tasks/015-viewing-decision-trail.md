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

### Delivery slices

1. **Pure domain (`internal/viewing`, no DB, no migration).** `ItemID`,
   `ChecklistItem`, `GenerateChecklist`, `DecisionStatus` and its allowed
   transitions, `ViewingAnswer`, contradiction detection, and
   `FollowUpQuestions`. Item IDs are `field:<key>` for missing fields and
   `default:<slug>` for the documented default checklist. Missing-field items
   come first in input order, then default items; duplicates collapse by ID.
2. **Persistence.** `DecisionTrailRepository` in `internal/database`. Needs
   explicit confirmation before the migration (see below) and an Advise call.
3. **UI.** History editor, status/next-action form, three-way view, and the
   explicit copy action in `internal/ui` and `cmd`.

Slice 1 is complete. Design choices made in it: any valid status may follow any
other (a rejected rental can be reopened, so there is no transition table);
`SetAnswer` rejects blank text and `ClearAnswer` is the only way to remove one;
answers are keyed by item ID and may target listing fields, so a claim can be
checked at the viewing; `AnswerLine.Differs` is a normalized (case and
whitespace) text mismatch that flags a pair for the user to look at, not a proof
of contradiction, so "4tr5" versus "4,500,000 VND" is flagged; the follow-up
lists every unresolved item and takes questions only, so answers and notes
cannot leak into it. `isAbsent` duplicates the Gemini layer's placeholder list
on purpose to keep `viewing` free of a `gemini` import.

Slice 2 is complete (`internal/database/trail.go`). Design choices made in it:
one additive `decision_trails` table (`rental_id` PK, `status`, `next_action` as
UTC RFC3339, `answers` as a JSON object, `updated_at`) with no foreign key;
`Save` is one upsert; `Get` on a missing trail returns `viewing.NewTrail(id)`;
`Save` rejects an unknown status; `InitDB` chmods the database file to 0o600
after the first DDL statement (no-op on Windows); logs carry only the rental ID
and answer count. Orphans are prevented rather than handled: the maintainer
decided (2026-10-04) that the delete confirmation offers only "delete rental and
its viewing notes" or "cancel", and `DeleteRentalAndTrail` removes both in one
transaction. This narrows the 2026-10-03 note: "keep notes after the rental is
gone" is not offered, so no orphan-listing method exists. Slice 3 must route
`ui.DeleteRentalUI` through `DeleteRentalAndTrail` and add that confirmation.

### Slice 3 program design

Advise ran 2026-10-05 (see `docs/eagd-log.md`). No schema change; no new
interface: the UI uses `database.NewDecisionTrailStore(database.DB)` directly.

- `internal/viewing` (pure, tested): `EditableItems(checklist []ChecklistItem,
  claims map[ItemID]string) []ChecklistItem` returns the checklist followed by a
  verify item for each claimed field not already in it (question is
  `Xác nhận: <key> (tin đăng: <claim>)`), so a viewing answer can be checked
  against a listing claim and `AnswerLine.Differs` can become true. Verify items
  are for the editor only; the follow-up still uses the checklist's unresolved
  items, so claims and notes never reach it.
- `internal/ui/viewing.go`: pure helpers `claimsFromResult(*gemini.RentalExtractionResult)
  map[ItemID]string` (uses `RawFields`, fills the standard fields only where
  `RawFields` lacks the key, drops `additional_notes`), `parseNextAction(string)
  (*time.Time, error)`, `formatNextAction(*time.Time) string`,
  `statusLabel(DecisionStatus) string`, `applyAnswer(*DecisionTrail, ItemID, string)
  error` (blank clears, otherwise sets), `trailSummary(DecisionTrail) string` and
  `renderReview(io.Writer, DecisionTrail, Review)`. Thin `huh` forms:
  `ViewingTrailUI()` (pick rental, then answers / status and date / three-way view /
  follow-up) and a "Hồ sơ xem phòng" entry in `ShowHistoryAndCompareMenu`.
- Next-action date is a calendar date, not a moment: `parseNextAction` builds
  `time.Date(y, m, d, 0, 0, 0, 0, time.UTC)` from `YYYY-MM-DD` (blank clears) and
  every display formats with `.UTC()`, so the day cannot shift with the machine's
  time zone (the store round-trips UTC).
- Follow-up: show the message, then a `huh.Confirm` defaulting to No; copy only on
  Yes. An empty message prints "nothing to ask" and offers no copy. A clipboard
  error is printed without the message text.
- History and comparison: `ListRentalsUI` and `RenderComparisonTable` read each
  trail with `Get` only (never `Save` or `Delete`) and show `trailSummary`; a
  failed `Get` shows "không đọc được" for that rental instead of failing the view.
- Delete: `DeleteRentalUI` replaces its Confirm with a Select of "Hủy" (listed
  first and preselected) and "Xóa phòng và ghi chú xem phòng", calling
  `database.DeleteRentalAndTrail`; back also cancels.
- Notes never enter errors or logs: no answer text is passed to `fmt.Errorf` or
  `logger.*`; `renderReview` writes only to its `io.Writer`.
- Follow-ups: `database.DeleteRental` keeps only a test caller after this slice
  (candidate cleanup, separate item); `renderer.go` auto-copies the phone number
  and polite message after analysis (candidate idea, separate item).

Slice 3 is complete (`internal/ui/viewing.go`, `internal/viewing/editable.go`). It follows the design above with no deviation. The history menu gains "Hồ sơ xem phòng"; the list and comparison table show each rental's status read-only; delete uses the cancel-first Select and `DeleteRentalAndTrail`. Manual terminal checks (restart persistence, mixed answers, follow-up excluding answered questions) are still the maintainer's to run: the huh forms are not unit-tested.

Slice 1 takes no search-profile input because Task 013 does not exist yet; the
profile source is added when 013 lands (deviation from the first DoD bullet,
which says "when present").

## Dependencies and scope

- The task can use Task 013 profiles when available but must also work with the
  default checklist alone.
- Photo capture, document storage, reminders outside the running CLI, calendar
  integration, and automated fraud labels are out of scope.
- Persisting this feature requires a SQLite schema migration or new tables. Get
  explicit confirmation immediately before that migration is implemented, as
  required by `AGENTs.md`.

## Notes and deviations

Deleting a rental must not leave orphaned decision data. Decided by the
maintainer (2026-10-03): deleting a rental asks for a separate confirmation about
also deleting its decision trail, rather than cascading silently. The slice 2
schema and delete flow must define how a trail whose rental is gone is handled,
so no orphaned notes are left unreachable.
