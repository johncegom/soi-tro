# EAGD log

Append one single-line row per event at the end of the matching table. Use `—` for a field that does not apply and write a literal `|` as `\|`. Never rewrite old rows.

## Advise calls

| Date | Branch | Question | Prior leaning | Answer | Taken | Tool | Requested | Reported | Status | Changed |
|---|---|---|---|---|---|---|---|---|---|---|
| 2026-09-20 | chore/lint-security-phase2 | Task 017 item 3.1: shape of `run` extraction from `main` (keep DoD signature, handlers struct, or enum dispatch?) | `handlers` struct of func fields + `dispatch`, untested `runAnalyze` | Drop the struct; test `dispatch(choice) action` enum + `mimeTypeFor`; do not thread stdin/stdout; extract `runAnalyze` verbatim in this task; record signature change as a deviation | Partly: struct dropped, rest taken | Agent | opus | claude-opus-5 | OK |
| 2026-09-29 | task/013 | Conservative local profile evaluation and move-in cash semantics | One preferences package; exact rent plus deposit; unknowns per configured rule | Keep one package; exact-value parsing; unknown tolerance optional; label cash rent + deposit and reject misleading passes | Taken | spawn_agent | gpt-6-astra medium (assumed) | gpt-6 | OK (assumed binding) |
| 2026-09-29 | main | Keyboard navigation contract for Task 013 | Record contract in Task 013 and test input, confirmation, and menu; shared wrapper makes task-only note less visible | Record contract in Task 013, add comment beside wrapper, and test focused controls; no new shared document | Taken | spawn_agent | gpt-6-astra medium (assumed) | /root/advise_navigation_design | OK (assumed binding) |
| 2026-09-29 | main | Vietnamese presentation for Task 013 | Preserve machine codes and JSON; translate evaluator reasons and CLI labels, though reasons then live outside UI | Keep codes stable; evaluator emits Vietnamese explanations and UI translates labels; check validation errors | Taken | spawn_agent | gpt-6-astra medium (assumed) | GPT-6 | OK (assumed binding) |
| 2026-09-29 | main | Replace move-in cash setting with deposit months while preserving saved profiles | Add month field, retain legacy read/eval; explicit month entry replaces old field | Preserve legacy on unrelated saves; no load migration; reject both limits; exact integer comparison | Taken | spawn_agent | gpt-6-astra medium (assumed) | GPT-6 | OK (assumed binding) |

## Binding changes

| Date | Role | Tool | Old | New | Reason |
|---|---|---|---|---|---|
| 2026-09-18 | advise | Agent | — | opus (ok) | initial install, probe reported claude-opus-5 |
| 2026-09-18 | grade | Agent | — | haiku (ok) | initial install, probe reported claude-haiku-4-5-20251001 |
| 2026-09-18 | dream | Agent | — | opus (ok) | initial install, same model as advise, covered by the opus probe (claude-opus-5) |
| 2026-09-23 | advise | Agent | opus (ok, reported claude-opus-5) | opus (ok, reported claude-opus-5-5) | re-probe on /bootstrap-eagd-pattern re-run, underlying opus version upgraded |
| 2026-09-23 | dream | Agent | opus (ok, reported claude-opus-5) | opus (ok, reported claude-opus-5-5) | re-probe on /bootstrap-eagd-pattern re-run, underlying opus version upgraded |
| 2026-09-29 | advise | spawn_agent | gpt-5.6-sol high (unverified; retry gpt-6-astra medium) | gpt-6-astra medium (ok, assumed) | Maintainer requested assumed status=ok; probe reported gpt-6, then unknown, so exact variant is unverified |
| 2026-09-29 | grade | spawn_agent | gpt-5.6-luna high (unverified) | gpt-6-luna high (ok, assumed) | Maintainer requested assumed status=ok; probe reported gpt-5.6-luna, then unknown, so exact variant is unverified |
| 2026-09-29 | dream | spawn_agent | gpt-6-astra medium (unverified) | gpt-6-astra medium (ok, assumed) | Maintainer requested assumed status=ok; probe reported gpt-6, then unknown, so exact variant is unverified |

## Grade fallbacks

| Date | Branch | Tool | Reason |
|---|---|---|---|
