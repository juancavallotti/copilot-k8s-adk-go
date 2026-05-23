// Package observability wires structured slog output and ADK callback hooks
// so we can trace LLM calls, tool invocations, and session events.
package observability

import (
	"log/slog"
	"os"
	"strings"
)

// Init installs a JSON slog handler as the default logger at the given level.
// Empty or unrecognized values fall back to info.
func Init(level string) {
	var lvl slog.Level
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	slog.SetDefault(slog.New(handler))
}
