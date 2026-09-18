# Bug Log

Add an entry when something in the product is already wrong. Do not use this
file for planned features or general cleanup.

## Entry format

```markdown
## BUG-<NNN>: <short symptom>

**Symptom:** <observed behavior>

**Root cause:** <known cause or `unknown`>

**Reachability:** <current product call path, or test-only reachability>

**Options:** <possible fixes, when known>

**Status:** pending decision / decided / fixed in <reference>
```

## BUG-001: History listing can hide a row-iteration failure

**Symptom:** `database.ListRentals` can return a partial record list and a nil
error when SQLite reports an error after row iteration has started.

**Root cause:** The function exits after the `rows.Next()` loop without checking
`rows.Err()`.

**Reachability:** The history list and comparison flows call `ListRentals`
through `internal/ui/history.go`. The failure requires SQLite to report an
iteration-time error rather than a query or scan error.

**Options:** Check `rows.Err()` after the loop and return the error; add a focused
test using an injectable query boundary or a driver-level failure fixture.

**Status:** fixed in `internal/database/db.go` (`ListRentals` now checks
`rows.Err()` after iteration and returns it as `iterate_records`). No dedicated
regression test was added: reproducing a real iteration-time SQLite error
requires a driver-level failure fixture disproportionate to this one-line
correctness fix; existing `TestDBOperations` and `TestListRentals_DBError`
continue to cover the success and query-error paths.

## BUG-002: Valid Gemini authorization keys can be rejected during setup

**Symptom:** The interactive credential setup rejects a non-empty Gemini
credential unless its value begins with `AIzaSy`, preventing users from saving
valid credentials with another format.

**Root cause:** `EnsureGlobalAPIKey` inferred credential validity from a legacy
key prefix instead of treating the value as an opaque secret for the official
Google GenAI SDK to authenticate.

**Reachability:** A user without `GEMINI_API_KEY` or a saved key reaches the
interactive setup during application startup. The inline form validator blocks
the credential before `gemini.NewClient` can pass it to the SDK.

**Options:** Remove format-specific validation and retain only the non-empty
check; keep trimming, secure local storage, and SDK-side authentication.

**Status:** fixed in the Gemini authorization-key compatibility change
