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
- Task details: `docs/tasks/<NNN>-<slug>/TASK.md`
- Bug log: `docs/BUGS.md`
- Decision log: `docs/DECISIONS.md`
- Existing backlog and older task history: `feature-plan.md`

Read the ledger first, then open only the task documents relevant to the
current work.

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

Read `docs/execute-advise-grade-dream.md` for the design rationale. Follow
these instructions during non-trivial tasks.

**Execute.** Treat the primary session as Execute. When the task runner allows
the Execute model to be selected, prefer `gpt-5.6-sol` with
`reasoning_effort: medium`. Do not restart an active session merely to enforce
this preference.

**Advise.** When a decision is genuinely ambiguous and a wrong choice would be
costly to reverse, pause Execute and spawn a separate agent with:

- `model: gpt-5.6-sol`
- `reasoning_effort: high`
- no inherited conversation context

Give the advisor only the decision and the minimum evidence needed to answer
it. Do not send the whole transcript. Wait for its answer before continuing.
If that advisor explicitly cannot resolve the decision or reports low
confidence, retry once with `model: gpt-6-astra` and
`reasoning_effort: medium`. Do not substitute another model silently.

**Grade.** After finishing a non-trivial task with a Definition of Done and
Test Plan, and before reporting it complete, spawn a fresh agent with:

- `model: gpt-5.6-luna`
- `reasoning_effort: high`
- no inherited conversation context

Give Grade only the approved rubric, the finished diff or output, and the
verification results. Withhold the reasoning and discussion that produced
them. Grade must return pass or fail and name every failed criterion. A failure
defaults to a targeted fix followed by another fresh Grade pass. Restart the
implementation from the approved design only when Grade shows that the core
approach or its assumptions are invalid.

**Dream.** Run Dream only after Grade passes and only when the completed work
contains a durable decision that will help a later session. Spawn a separate
agent with:

- `model: gpt-6-astra`
- `reasoning_effort: medium`
- no automatically inherited conversation context

Supply the full run history explicitly, including Execute's reasoning, Advise
exchanges, and the Grade verdict. Dream drafts an entry for
`docs/DECISIONS.md`, or says that no durable entry is warranted. Execute must
still present the exact proposed entry and wait for the repository's normal
file-change approval before writing it.

Use the available agent-spawn tool, named `Agent` or `spawn_agent` depending on
the runtime. When overriding the model, start the role with fresh context and
pass only the material specified above.

Re-run the EAGD bootstrap if a role fires on nearly every task, since its
trigger or model is then too expensive for the value it adds. Recalibrate or
remove a role that does not fire over a long period instead of keeping unused
process.
