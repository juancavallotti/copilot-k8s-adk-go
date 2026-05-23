// Package observability wires structured slog output and ADK callback hooks
// so we can trace LLM calls, tool invocations, and session events.
package observability

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Init installs a JSON slog handler as the default logger at the given level.
// Records always go to stdout. If traceSink is non-nil, the subset of records
// that represent agent activity (agent.event, state.delta, llm.*, tool.*) is *also*
// written to traceSink — typically a TraceSink piping into recipes-cli
// log-trace for persistence. Other records (agent.starting,
// anything user code emits) stay on stdout only.
//
// Empty or unrecognized levels fall back to info.
func Init(level string, traceSink io.Writer) {
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
	opts := &slog.HandlerOptions{Level: lvl}
	primary := slog.NewJSONHandler(os.Stdout, opts)
	var h slog.Handler = primary
	if traceSink != nil {
		sink := slog.NewJSONHandler(traceSink, opts)
		h = &filteredTeeHandler{primary: primary, sink: sink, accept: isAgentTraceRecord}
	}
	slog.SetDefault(slog.New(h))
}

// isAgentTraceRecord matches the messages we want to persist as traces:
// session-level "agent.event" and "state.delta" records, plus any llm.* /
// tool.* callback. Bootstrap events (agent.starting) and stray library logs are skipped.
func isAgentTraceRecord(r slog.Record) bool {
	m := r.Message
	return m == "agent.event" || m == "state.delta" || strings.HasPrefix(m, "llm.") || strings.HasPrefix(m, "tool.")
}

// filteredTeeHandler always forwards to primary; it forwards to sink only
// when accept(record) returns true. Both branches share the same WithAttrs /
// WithGroup state so loggers built via slog.With still partition correctly.
type filteredTeeHandler struct {
	primary slog.Handler
	sink    slog.Handler
	accept  func(slog.Record) bool
}

func (h *filteredTeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.primary.Enabled(ctx, level) || h.sink.Enabled(ctx, level)
}

func (h *filteredTeeHandler) Handle(ctx context.Context, r slog.Record) error {
	err := h.primary.Handle(ctx, r)
	if h.accept(r) {
		if e := h.sink.Handle(ctx, r); e != nil {
			err = errors.Join(err, e)
		}
	}
	return err
}

func (h *filteredTeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &filteredTeeHandler{primary: h.primary.WithAttrs(attrs), sink: h.sink.WithAttrs(attrs), accept: h.accept}
}

func (h *filteredTeeHandler) WithGroup(name string) slog.Handler {
	return &filteredTeeHandler{primary: h.primary.WithGroup(name), sink: h.sink.WithGroup(name), accept: h.accept}
}
