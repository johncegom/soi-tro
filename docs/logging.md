# Logging

Soi Trọ uses Go's `log/slog` for diagnostic logs. Terminal prompts, analysis
results, and other user-facing output remain separate from these logs.

## Configuration

The logger reads these environment variables at startup:

| Variable | Values | Default |
|---|---|---|
| `LOG_LEVEL` | `debug`, `info`, `warn`, `error` | `info` |
| `LOG_FORMAT` | `text`, `json` | `text` |
| `LOG_FILE` | Log file path | `~/.config/soi-tro/app.log` |
| `LOG_CONSOLE` | `true`, `false` | `true` |

The file and console receive the same records when console output is enabled.
Log rotation is not implemented; manage retention outside the application if
the log file needs a size or age limit.

## Field conventions

- `operation`: stable dotted operation name, such as `gemini.extract_rental` or
  `database.save_rental`.
- `duration_ms`: elapsed operation time in whole milliseconds.
- `request_id`: correlation ID for context-aware operations. It is currently
  attached to Gemini analysis and related orchestration events.
- `error`: stable failure-stage category such as `load_schema` or
  `insert_record`. It is not a raw returned error.
- `version`: application version on the startup event.

Database and exporter APIs do not accept a context, so their records do not
carry `request_id`. Adding context only for logging would spread a new parameter
through UI call paths without providing cancellation or request ownership.

Example JSON record:

```json
{"level":"INFO","msg":"operation completed","request_id":"example-request-id","operation":"gemini.extract_rental","duration_ms":842}
```

## Data safety

Logs may contain operation names, timings, safe failure stages, request IDs, and
the application version. They must not contain:

- Gemini API keys or other credentials;
- raw listing text or model responses;
- image bytes or image contents;
- schema contents;
- phone numbers, generated messages, or other personal contact details;
- rental result fields or exported report contents;
- filesystem paths supplied by the user or derived from the home directory.

Return detailed errors to callers for display or handling, but log only the safe
failure stage. When adding an operation, test with distinctive secret and contact
markers and assert that neither marker appears in captured logs.

## Usage

Use the process logger for operations without a context:

```go
started := time.Now()
failureStage := "insert_record"
log := logger.With("operation", "database.save_rental")
defer func() {
	logger.LogOperationResult(log, started, failureStage, err)
}()
```

Use `FromContext` when the call path already carries a context:

```go
started := time.Now()
failureStage := "generate_content"
log := logger.FromContext(ctx).With("operation", "gemini.extract_rental")
defer func() {
	logger.LogOperationResult(log, started, failureStage, err)
}()
```

Do not log every function call. Add records at boundaries where an operation can
fail, take meaningful time, or help connect a user-visible failure to a safe
diagnostic event.
