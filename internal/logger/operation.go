package logger

import (
	"log/slog"
	"time"
)

// LogOperationResult records the duration and outcome of one operation. Callers
// must pass a safe failure-stage category rather than a raw error message.
func LogOperationResult(log *slog.Logger, started time.Time, failureStage string, err error) {
	attrs := []any{"duration_ms", time.Since(started).Milliseconds()}
	if err != nil {
		log.Error("operation failed", append(attrs, "error", failureStage)...)

		return
	}

	log.Info("operation completed", attrs...)
}
