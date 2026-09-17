# Task Documents

Create one directory per non-trivial task:

```text
docs/tasks/<NNN>-<slug>/TASK.md
```

Use the next available ledger number. Keep task-specific discoveries here;
put product defects in `docs/BUGS.md` and deliberate tradeoffs in
`docs/DECISIONS.md`.

## Template

```markdown
# Task: <name>

## Definition of Done

- [ ] <concrete, testable outcome>

## Test Plan

- Automated: <tests and exact commands>
- Manual: <checks and exact steps>

## Program design

<!-- Include only when required by AGENTS.md. -->

- Types/interfaces:
- Key signatures:
- Call graph or event/data flow:
- Package/file layout:

## Sub-tasks

<!-- Include only for work that benefits from resumable checkpoints. -->

- [ ] 1.1 <work item>

## Notes and deviations

<Task-specific implementation notes and approved deviations.>
```
