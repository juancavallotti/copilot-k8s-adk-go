package observability

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestInit_nilSinkDoesNotPanic(t *testing.T) {
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })

	Init("info", nil)
	slog.Info("agent.event", "invocation_id", "inv-1")
}

func TestInit_filterAllowsAgentEvent(t *testing.T) {
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })

	var sink bytes.Buffer
	Init("info", &sink)
	slog.Info("agent.event", "invocation_id", "inv-1", "text", "hello")

	got := sink.String()
	if !strings.Contains(got, `"msg":"agent.event"`) {
		t.Fatalf("sink missing agent.event: %q", got)
	}
	if !strings.Contains(got, `"invocation_id":"inv-1"`) {
		t.Fatalf("sink missing invocation_id: %q", got)
	}
}

func TestInit_filterAllowsStateDelta(t *testing.T) {
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })

	var sink bytes.Buffer
	Init("info", &sink)
	slog.Info("state.delta", "invocation_id", "inv-1", "delta", map[string]any{"x": 1})

	got := sink.String()
	if !strings.Contains(got, `"msg":"state.delta"`) {
		t.Fatalf("sink missing state.delta: %q", got)
	}
	if !strings.Contains(got, `"invocation_id":"inv-1"`) {
		t.Fatalf("sink missing invocation_id: %q", got)
	}
}

func TestInit_filterAllowsLLMAndToolPrefixes(t *testing.T) {
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })

	var sink bytes.Buffer
	Init("debug", &sink)
	slog.Debug("llm.start", "invocation_id", "inv-1")
	slog.Debug("llm.end", "invocation_id", "inv-1")
	slog.Info("tool.start", "invocation_id", "inv-1", "tool", "x")
	slog.Info("tool.end", "invocation_id", "inv-1", "tool", "x")
	slog.Info("tool.request", "invocation_id", "inv-1", "tool", "x")

	got := sink.String()
	for _, want := range []string{`"msg":"llm.start"`, `"msg":"llm.end"`, `"msg":"tool.start"`, `"msg":"tool.end"`, `"msg":"tool.request"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("sink missing %s: %q", want, got)
		}
	}
}

func TestInit_filterBlocksNonAgentRecords(t *testing.T) {
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })

	var sink bytes.Buffer
	Init("info", &sink)
	slog.Info("agent.starting", "addr", "x")
	slog.Info("random.app.event")

	if sink.Len() != 0 {
		t.Fatalf("sink should be empty, got %q", sink.String())
	}
}

func TestInit_levelFiltering(t *testing.T) {
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })

	var sink bytes.Buffer
	Init("info", &sink)
	slog.Debug("llm.start", "invocation_id", "inv-1")

	if sink.Len() != 0 {
		t.Fatalf("debug record leaked at info level: %q", sink.String())
	}
}

func TestIsAgentTraceRecord(t *testing.T) {
	t.Parallel()
	cases := []struct {
		msg  string
		want bool
	}{
		{"agent.event", true},
		{"llm.start", true},
		{"llm.end", true},
		{"tool.start", true},
		{"tool.end", true},
		{"tool.request", true},
		{"tool.response", true},
		{"state.delta", true},
		{"agent.starting", false},
		{"", false},
		{"agent.event.extra", false}, // exact match only for agent.event
		{"toolbar.click", false},     // prefix is "tool." not "tool"
	}
	for _, c := range cases {
		got := isAgentTraceRecord(slog.NewRecord(time.Time{}, slog.LevelInfo, c.msg, 0))
		if got != c.want {
			t.Errorf("isAgentTraceRecord(%q) = %v, want %v", c.msg, got, c.want)
		}
	}
}
