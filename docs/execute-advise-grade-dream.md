# Execute / Advise / Grade / Dream

This pattern separates four kinds of work that benefit from different context.
Execute owns the task. Advise handles a hard judgment call during the task.
Grade checks the finished result with fresh eyes. Dream preserves the few
lessons worth carrying into later sessions.

## Why the roles are separate

Execute needs continuity. It sees the request, repository state, earlier
choices, and implementation details. That context helps it work, but it can
also make Execute defend its own assumptions.

Advise receives one bounded question. Its job is to resolve a decision whose
cost of reversal justifies a second opinion. Sending the whole transcript
wastes context and makes the advisor less independent.

Grade receives the approved criteria and finished result without the reasoning
that produced them. This prevents the grader from accepting an output merely
because the implementation story sounds plausible.

Dream runs only after Grade passes. It sees the full run because its job is to
find a lasting decision or lesson across the whole task. Most runs should
produce no Dream entry.

## Information flow

```text
request -> Execute -> finished output -> Grade
              |
              +-> Advise -> decision -> Execute

Grade fail -> targeted fix or design restart -> fresh Grade
Grade pass -> optional Dream -> proposed decision-log entry
```

Advise blocks Execute until it replies. Grade starts fresh for every attempt.
Dream never runs before a passing Grade result.

## Failure handling

A narrow Grade failure should cause a narrow fix. A full restart is justified
when the failure invalidates the chosen design, a core assumption, or enough
of the output that patching it would hide the real problem.

Dream does not bypass repository controls. If it finds a lasting decision, it
drafts an entry for the named repository log. The primary agent presents that
change for approval before writing it.

## Cost and recalibration

Separate calls save useful context only when their prompts stay bounded.
Advise should be rare. Grade belongs on work with explicit pass/fail criteria,
and Dream should be rarer still.

If a role fires on almost every task, narrow its trigger or use a cheaper
model. If it never fires after a meaningful period, recalibrate or remove it.
An unused role adds ceremony without improving the work.
