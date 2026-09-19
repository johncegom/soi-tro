# EAGD log

Append one single-line row per event at the end of the matching table. Use `—` for a field that does not apply and write a literal `|` as `\|`. Never rewrite old rows.

## Advise calls

| Date | Branch | Question | Prior leaning | Answer | Taken | Tool | Requested | Reported | Status |
|---|---|---|---|---|---|---|---|---|---|
| 2026-09-20 | chore/lint-security-phase2 | Task 017 item 3.1: shape of `run` extraction from `main` (keep DoD signature, handlers struct, or enum dispatch?) | `handlers` struct of func fields + `dispatch`, untested `runAnalyze` | Drop the struct; test `dispatch(choice) action` enum + `mimeTypeFor`; do not thread stdin/stdout; extract `runAnalyze` verbatim in this task; record signature change as a deviation | Partly: struct dropped, rest taken | Agent | opus | claude-opus-5 | OK |

## Binding changes

| Date | Role | Tool | Old | New | Reason |
|---|---|---|---|---|---|
| 2026-09-18 | advise | Agent | — | opus (ok) | initial install, probe reported claude-opus-5 |
| 2026-09-18 | grade | Agent | — | haiku (ok) | initial install, probe reported claude-haiku-4-5-20251001 |
| 2026-09-18 | dream | Agent | — | opus (ok) | initial install, same model as advise, covered by the opus probe (claude-opus-5) |

## Grade fallbacks

| Date | Branch | Tool | Reason |
|---|---|---|---|
