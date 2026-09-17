# Decision Log

## DECISION-001: Use the Tier 2 working agreement

**Context:** Soi Trọ has one primary human maintainer but is developed across
independent agent sessions. It is maintained and released, and it handles API
credentials, personal contact information, local SQLite data, and report
files.

**Decision:** Use a task-approval gate with a ledger, per-task detail
documents, a bug log, and this decision log.

**Alternatives considered:** Tier 1 would add less maintenance, but would not
preserve task scope and approval context between agent sessions. Tier 3 would
require deriving behavior from an external oracle, but this project is not
porting an authoritative implementation. A retrospective log was also
considered and deferred until recurring process lessons make it useful.

**Consequences:** Non-trivial work needs an approved Definition of Done and
Test Plan before implementation. The maintainer must keep the ledger and logs
current. The process should be recalibrated if that cost stops paying for
itself or the project's risk changes.

## DECISION-002: Use separate agents for EAGD review roles

**Context:** Work in this repository can involve hard-to-reverse architecture
choices, local user data, credentials, and release artifacts. The task workflow
also provides explicit criteria that can be graded independently.

**Decision:** Keep the primary session as Execute and use separate agent calls
for Advise, Grade, and Dream. Grade failures default to targeted fixes. Dream
may propose durable entries only for `docs/DECISIONS.md` and remains subject to
the file-change approval gate. Live model routing belongs in `AGENTs.md` so it
can be recalibrated without duplicating configuration here.

**Alternatives considered:** Same-session role labels would provide some
structure but would not give Advise or Grade independent context. Installing
only Advise would miss the repository's existing Definition-of-Done checks.
Running Dream without a Grade gate would create noisy history.

**Consequences:** Non-trivial completed work receives a fresh review, while
costlier agents are reserved for bounded decisions and lasting lessons. The
mechanism should be recalibrated if a role runs routinely or remains unused.

## DECISION-003: Replace universal file approval with risk-based authorization

**Context:** Requiring a second approval for every file write adds a mandatory
round trip even when the maintainer has explicitly requested an ordinary,
reversible implementation. It also makes routine task bookkeeping subject to
the same gate as destructive or externally visible work.

**Decision:** Treat explicit change, implementation, build, and fix requests as
authorization for ordinary in-scope edits. Keep explicit confirmation for
destructive actions, sensitive data or credentials, external effects, material
scope expansion, and possible overwrites. Read-only requests remain read-only.
This supersedes the universal approval requirement established by
DECISION-001; its remaining Tier 2 workflow decisions still apply.

**Alternatives considered:** Keeping universal approval provides a simple
rule, but turns routine work into ceremony. Removing approval safeguards
entirely would expose user data and release operations to avoidable mistakes.

**Consequences:** Routine implementation can proceed in one turn while risky
actions still stop for human confirmation. Agents must classify the request
correctly, preserve unrelated worktree changes, verify their edits, and report
what changed.
