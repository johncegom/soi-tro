# Task Documents

Use one Markdown file per non-trivial task by default:

```text
docs/tasks/<NNN>-<slug>.md
```

Use the next available ledger number. Create
`docs/tasks/<NNN>-<slug>/TASK.md` instead only when the task has supporting
artifacts that belong beside it. Do not create a directory for a lone task
document.

Keep task-specific discoveries with the task; put product defects in
`docs/BUGS.md` and deliberate tradeoffs in `docs/DECISIONS.md`.

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
