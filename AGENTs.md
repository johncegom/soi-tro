# AGENTS

Read this file before working in the repository.

Soi Trọ is a Go CLI that analyzes Vietnamese rental listings with Gemini,
stores local history in SQLite, and exports reports. Changes can affect API
keys, personal contact information, local user data, and released binaries.

## Way of working

This project uses Tier 2 (Standard): risk-based file authorization, a task
ledger, bug log, and decision log. The repository has one primary human
maintainer, but work continues across independent agent sessions and affects
real user data.

- Task ledger: `docs/LEDGER.md`
- Idea inbox: `docs/IDEAS.md`
- Task details: `docs/tasks/<NNN>-<slug>.md` by default; use
  `docs/tasks/<NNN>-<slug>/TASK.md` only when a task needs supporting artifacts
- Bug log: `docs/BUGS.md`
- Decision log: `docs/DECISIONS.md`
- Existing backlog and older task history: `feature-plan.md`

Read the ledger first, then open only the task documents relevant to the
current work. Read the idea inbox when brainstorming or promoting an idea into
the ledger.

## File-change authorization

An explicit request to change, implement, build, or fix something authorizes
ordinary in-scope workspace edits. Do not require a second approval for
reversible changes that directly serve that request.

For substantial or resumable work, create a task document with a concrete
Definition of Done and Test Plan before implementation. Small, localized work
does not need a task document.

Ask for explicit confirmation before:

- destructive or difficult-to-recover operations;
- database migrations or changes to generated user data;
- handling credentials or changing security-sensitive access;
- publishing, deploying, releasing, or causing another external side effect;
- materially expanding the requested scope;
- overwriting user work or resolving an unclear file target.

A request to answer, explain, review, diagnose, or propose changes does not
authorize file edits. If the user requests a diff or proposal first, wait for
approval before applying it.

Before editing, inspect the worktree and preserve unrelated changes. After
editing, run proportionate verification and summarize the changed files,
results, and any deviation from the task plan.

## Program design

Add a Program design section before implementation when a task changes
multiple interacting files or functions, or when an agent will generate a
substantial amount of new code in one pass. Record:

- types and interfaces being added or changed;
- important method or function signatures;
- the call graph, event flow, or data flow;
- package boundaries when more than one package is involved.

Small, localized, mechanical edits do not need a Program design section.

## Scope control

Do not fold unrelated cleanup, stale documentation, or adjacent defects into
the active task. Record them as a new ledger item, a bug, or a note to the
user so they can be considered separately.

## Bugs

Log a defect when the product is already behaving incorrectly. Each entry
must include its symptom, root cause or `unknown`, reachability through the
current product, possible fixes when known, and status.

Logging a bug does not authorize fixing it. Wait for a decision and the normal
file-change approval. Report a live security exposure, active data-loss path,
or other ongoing harm to the user immediately.

## Decisions

Record a decision when an informed reviewer could reasonably have chosen a
different approach and would benefit from knowing why this one was selected.
Routine or easily reversible choices do not need entries.

## Test-driven development

Every behavior change and bug fix is test-first. Size is not an exemption.

1. **Red.** Write the failing test first and run it. It must fail on an
   assertion about the missing behavior, not on a compile error, typo, or
   unrelated failure. Keep the command and the failing line.
2. **Green.** Write the minimum production code that passes it.
3. **Refactor.** Clean up with the tests green, then run `go test ./...`.

A bug fix starts by reproducing the bug as a failing test; that test is the
"regression coverage" named in the `docs/BUGS.md` entry.

Do not write production code before a failing test exists. Do not edit or
weaken a test to make code pass unless the test's expectation is wrong, and
say why. Test observable behavior; do not mock the code under test. If work
turns out to have been written code-first, stash the production change and
redo it from red rather than backfilling a test.

Record the red run and the green run in the final report or PR body. Missing
evidence fails Grade.

Exempt, and say which applies: docs-only, comment or formatting changes,
dependency or config changes with no behavior, and code that only reaches a
live service (Gemini). For the last, extract the logic into a pure function
and test that.

## Build and verification

Use the commands that match the change:

```text
go build ./...
go test ./...
go vet ./...
golangci-lint run
```

The Taskfile also provides `task build`, `task test`, `task test:cover`, and
`task test:race`. Note that `task build` invokes `go-winres@latest` and may
require network access.

Tests must not depend on a live Gemini request unless the task explicitly
requires an integration test and the user approves using external services.
Never print or commit API keys, personal contact details, or generated user
data.

## Proportionality

Do not add an abstraction, defensive branch, dependency, or process step
without a real and currently reachable reason. “Just in case” is not enough.

## Recalibration

Revisit this tier when:

- another human maintainer joins;
- the project starts handling more sensitive data, money, or credentials;
- independent agent sessions become rare enough that the approval gate feels
  ceremonial;
- a required log stays stale across several tasks that should have updated it;
- missing process repeatedly causes lost context or unreviewed scope changes.

If an artifact becomes dead weight, propose retiring it. Archive logs that
contain history, update this file, and leave old cross-references intact.

## Execute / Advise / Grade / Dream

Rationale: `docs/execute-advise-grade-dream.md`. Log: `docs/eagd-log.md`.
Each role below is a genuinely separate call to your sub-agent tool with its
own `model` field, not a labeled phase in your own session.

Bindings are keyed by the sub-agent tool you actually hold (check your tool
list; do not guess your vendor). Use the row for `role` + your tool with
`status=ok`. Never run Advise or Dream on your own model or an unknown one.

<!-- eagd-bindings:start -->
eagd-binding: role=advise tool=Agent model=opus status=ok probed=2026-09-18 reported=claude-opus-5
eagd-binding: role=grade tool=Agent model=haiku status=ok probed=2026-09-18 reported=claude-haiku-4-5-20251001
eagd-binding: role=dream tool=Agent model=opus status=ok probed=2026-09-18 reported=claude-opus-5
eagd-binding: role=advise tool=spawn_agent model=gpt-5.6-sol effort=high retry_model=gpt-6-astra retry_effort=medium status=unverified
eagd-binding: role=grade tool=spawn_agent model=gpt-5.6-luna effort=high status=unverified
eagd-binding: role=dream tool=spawn_agent model=gpt-6-astra effort=medium status=unverified
<!-- eagd-bindings:end -->

**`spawn_agent` harness (GPT models).** These rows were carried over from the
earlier hand-written setup and have never been probed, so they are
`status=unverified` and fall under the "no usable row" rules below (Advise
and Dream skip, Grade falls back). Re-run `/bootstrap-eagd-pattern` from that
harness: it probes each row, and only then sets `status=ok`. `effort` is the
`reasoning_effort` to pass; every spawn starts with no inherited
conversation context. For Advise only: if the advisor explicitly says it
cannot resolve the decision or reports low confidence, retry once with
`retry_model` and its `retry_effort`, add a second Advise-calls row for the
retry, and never substitute another model silently.

**Execute.** The primary session. Do not restart it to change its model. On
a `spawn_agent` harness, when the task runner lets you pick the Execute
model, prefer `gpt-5.6-sol` with `reasoning_effort: medium`.

**Advise.** Fires on observable conditions, never on felt doubt: one call
before writing a Program design section, and one call before changing DB
schema or migrations, the Gemini prompt or listing-fence (injection)
handling, or credential/contact-data handling. Only judgment calls go here:
answer anything the repo can settle by reading it, and take preferences to
the user (`AskUserQuestion`) or a stated default. Write your leaning and why
in one or two lines, then call your tool with the `role=advise` model. The
prompt contains: the question; your leaning with the case for and against;
the artifacts the decision turns on, verbatim (not your summary, not the
transcript); a request to name any context it lacked; and a first
instruction to begin its reply with `model: <its own id>`. Wait for the
reply before continuing. If `reported` differs from the row, set the row to
`status=stale`, add a Binding-changes row, and skip Advise until fixed. No
usable row or the spawn errors: proceed on your leaning, log Status
`SKIPPED reason=no-verified-model-binding` (or the actual code), and say in
the final report that Advise did not run. After each call add one row to
"Advise calls" in `docs/eagd-log.md`.

**Grade.** After finishing a task that has a Definition of Done and Test
Plan, and before reporting it complete, call your tool with the `role=grade`
model in a fresh context. Give it only the approved rubric, the finished
diff or output, and the verification results including the TDD red and
green evidence; withhold the reasoning and discussion that produced them. It returns pass or fail and names every
failed criterion. Default on fail: a targeted fix, then a fresh Grade pass;
rerun the implementation from the approved design only when Grade shows the
core approach or its assumptions are invalid. No usable row: run the same
fresh, context-free call on your own model and add a row to "Grade
fallbacks".

**Dream.** Only after Grade passes and only when the work contains a durable
decision that will help a later session. Call your tool with the
`role=dream` model, passing the full run history (your reasoning, Advise
exchanges, the Grade verdict) and a first instruction to begin with
`model: <its own id>`. It drafts an entry for `docs/DECISIONS.md` or says
none is warranted. Present the exact proposed entry and wait for the normal
file-change approval before writing it. No usable row: skip and say so.

**Re-calibrate** by re-running the bootstrap when: a role fires on nearly
every task; a role never fires over a long period (remove it); Advise's
decision-change rate (answer differed from prior leaning) sits near zero; or
`SKIPPED` rows pile up. Count `SKIPPED` rows per Tool first, since a low
firing rate from missing bindings is fixed by re-running from that harness,
not by changing the trigger. Read the change rate per Tool, not in
aggregate.
