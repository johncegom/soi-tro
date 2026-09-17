# Task: Install EAGD role-spawn mechanism

## Definition of Done

- [x] `AGENTs.md` contains imperative triggers for Advise, Grade, and Dream.
- [x] Every installed role names its model, reasoning effort, input boundary,
  and response handling.
- [x] Grade has an explicit failure path and Dream has a real persistence
  target.
- [x] The design rationale is stored inside the repository.
- [x] The role-routing decision and recalibration conditions are recorded.
- [x] No existing skill is rewritten.

## Test Plan

- Automated: run `git diff --check`.
- Manual: inspect the scoped diff and confirm every role has a trigger, exact
  model assignment, context boundary, and result-handling rule.
- Manual: confirm the local rationale path referenced by `AGENTs.md` exists.
- Manual: confirm no file under a skill directory is changed by this task.

## Program design

- Types/interfaces: none; this task changes agent instructions only.
- Key signatures: role calls specify `model`, `reasoning_effort`, and fresh
  context.
- Flow: Execute may block on Advise; completed work goes to a fresh Grade;
  failed work returns to Execute; a Grade pass may trigger Dream.
- File layout: live instructions remain in `AGENTs.md`; rationale lives in
  `docs/execute-advise-grade-dream.md`; lasting Dream output targets
  `docs/DECISIONS.md`.

## Notes and deviations

The bootstrap skill's referenced rationale file was absent from the installed
skill directory. The repository copy was reconstructed from the rationale and
mechanical requirements contained in `SKILL.md`.

The maintainer selected separate agent invocations with these assignments:

- Execute: `gpt-5.6-sol`, medium reasoning.
- Advise: `gpt-5.6-sol`, high reasoning; retry with `gpt-6-astra`, medium
  reasoning, when the first advisor cannot resolve the decision or reports low
  confidence.
- Grade: `gpt-5.6-luna`, high reasoning.
- Dream: `gpt-6-astra`, medium reasoning.
