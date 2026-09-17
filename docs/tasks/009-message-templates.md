# Task: Add customizable message templates

Proposed task migrated from legacy Feature Plan Task 4. It is not approved for
implementation.

## Definition of Done

- [ ] Let users view and edit the message template through the CLI.
- [ ] Support documented placeholders for rental fields and missing fields.
- [ ] Validate templates before saving and retain the previous valid template
  when validation fails.
- [ ] Store configuration without exposing or overwriting API credentials.
- [ ] Preserve current generated-message behavior when no custom template exists.

## Test Plan

- Automated: unit-test parsing, supported and unsupported placeholders, invalid
  template recovery, configuration round trips, and default fallback.
- Manual: edit a template, generate a message from a saved analysis, restart the
  CLI, and verify that the customization persists.

## Program design

- Types/interfaces: use Go `text/template`; keep template configuration separate
  from API-key material even if both live below the same application directory.
- Key signatures: define only when the task is selected and the current config
  ownership has been reviewed.
- Data flow: CLI editor -> validation -> atomic config write -> render data ->
  generated message.
- Package/file layout: template rendering belongs outside terminal rendering;
  UI code should collect edits and display validation results.

## Notes and deviations

The legacy plan proposed `~/.config/soi-tro/config.json`. The exact storage
shape remains undecided because that file may also contain sensitive API-key
configuration and must not be overwritten accidentally.
