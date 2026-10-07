# Task: Add removable sample rentals for manual testing

Requested by the maintainer to try the viewing screens (Task 015) without
calling Gemini. Not a schema migration: it only inserts and deletes rows.

## Definition of Done

- [ ] A developer can insert a small set of fake rentals into the local history,
  covering missing fields and a listing claim an answer can contradict.
- [ ] Every sample row is tagged so it can be told apart from real data; seeding
  twice does not create duplicates.
- [ ] A developer can delete all sample rentals and their decision trails in one
  action, and no real rental or real trail is touched.
- [ ] The entry point is hidden unless `SOI_TRO_DEV=1` is set.
- [ ] No sample text includes real contact details.

## Test Plan

- Automated (no Gemini, temp DB): seed inserts the expected rows, all tagged;
  second seed inserts none; delete removes tagged rentals and their trails and
  keeps an untagged rental and its trail; delete with no samples removes none;
  the history menu offers the sample entry only when dev mode is on. Run
  `go test ./...`, `go vet ./...`, `golangci-lint run`.
- Manual: set `SOI_TRO_DEV=1`, seed from the history menu, open "Hồ sơ xem
  phòng" on a sample, then delete the samples and confirm they are gone.

## Design (small, localized; no Program design section)

- `internal/database/sample.go`: `SeedSampleRentals() (int, error)` and
  `DeleteSampleRentals() (int, error)`. A sample row is a rental whose
  `additional_notes` starts with `[SAMPLE]` (free text that is never a claim, so
  no schema change). Delete removes the matching trails and rentals in one
  transaction.
- `internal/ui/history.go`: `historyMenuOptions(dev bool)` adds "Dữ liệu mẫu
  (dev)" when `SOI_TRO_DEV=1`; a thin form seeds or deletes after a confirm.
