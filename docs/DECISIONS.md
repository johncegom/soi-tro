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

## DECISION-004: Structured logging uses safe failure stages and existing context boundaries

**Context:** Task 011 added structured logging to database, Gemini, and exporter
operations. Returned errors can contain model responses or filesystem paths, so
logging raw errors could expose sensitive data. Gemini operations already receive
a context; database and exporter operations do not.

**Decision:** Boundary events use `operation`, `duration_ms`, and, on failure, an
`error` field containing a stable failure stage such as `generate_content` or
`insert_record`. Original detailed errors continue to reach callers unchanged.
`logger.LogOperationResult` centralizes this contract. Events omit credentials,
listing bodies, images, schemas, contacts, model responses, report contents, and
user paths. `request_id` is attached through `logger.FromContext(ctx)` where an
existing context supplies it; APIs do not acquire context parameters solely for
logging.

**Alternatives considered:** Logging raw errors would provide more diagnostic
detail but risk sensitive-data disclosure. Adding context parameters throughout
database and exporter APIs would permit wider correlation but create API and
caller changes without an existing request-lifecycle need. Separate result-logging
implementations in each package would duplicate the same contract.

**Consequences:** Logs identify the failing operation and stage without carrying
the underlying payload. Diagnosis may require the detailed error returned to the
caller. Request correlation remains limited to operations with an existing
context. Future boundary logs should follow the shared contract and use safe
stage values; context propagation can expand when an operation gains a concrete
lifecycle or correlation requirement.

## DECISION-005: Keep early product ideas separate from proposed tasks

**Context:** Product ideas need a durable home before their scope, requirements,
and verification criteria are clear. Putting every early idea directly into the
task ledger would mix exploration with scoped work.

**Decision:** Use `docs/IDEAS.md` as a single lightweight idea inbox. Track ideas
as New, Exploring, Promoted, Parked, or Rejected. Promotion creates a Proposed
ledger item and a task document with a Definition of Done, Test Plan, and Program
design when required. Maintain links in both directions. Promotion records
proposed work; it does not by itself authorize implementation.

**Alternatives considered:** Recording all ideas directly in the ledger would
avoid another file but blur the distinction between an interesting possibility
and a scoped task. Creating a separate document for every idea would add
maintenance before the idea warrants detailed planning.

**Consequences:** Early ideas can be captured without premature task design.
Promoted ideas remain traceable to their original context, while the ledger
remains focused on scoped work. The inbox requires occasional status updates;
avoid duplicating evolving task details there.

**Status:** Accepted
