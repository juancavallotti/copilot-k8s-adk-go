package observability

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func TestEventPlugin_logsToolCallIDs(t *testing.T) {
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })

	var sink bytes.Buffer
	Init("info", &sink)
	p, err := NewEventPlugin()
	if err != nil {
		t.Fatalf("NewEventPlugin: %v", err)
	}
	cb := p.OnEventCallback()
	if cb == nil {
		t.Fatal("OnEventCallback is nil")
	}

	ev := session.NewEvent("inv-1")
	ev.ID = "evt-1"
	ev.Content = &genai.Content{
		Parts: []*genai.Part{
			{FunctionCall: &genai.FunctionCall{ID: "call-1", Name: "lookup", Args: map[string]any{"q": "pie"}}},
			{FunctionResponse: &genai.FunctionResponse{ID: "call-1", Name: "lookup", Response: map[string]any{"ok": true}}},
		},
	}
	if _, err := cb(nil, ev); err != nil {
		t.Fatalf("OnEventCallback: %v", err)
	}

	got := sink.String()
	for _, want := range []string{
		`"msg":"tool.request"`,
		`"msg":"tool.response"`,
		`"function_call_id":"call-1"`,
		`"event_id":"evt-1"`,
		`"invocation_id":"inv-1"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("sink missing %s: %q", want, got)
		}
	}
}

func TestAppendUserPromptAttr(t *testing.T) {
	attrs := attrMap(appendUserPromptAttr(nil, genai.NewContentFromText("Create a pasta recipe", genai.RoleUser)))
	if got := attrs["user_prompt"]; got != "Create a pasta recipe" {
		t.Fatalf("user_prompt = %v, want %q", got, "Create a pasta recipe")
	}
}

func TestAppendUserPromptAttr_extractsUserMessageFromContextEnvelope(t *testing.T) {
	content := genai.NewContentFromText(`{
		"appContext": {"screen": "other", "path": "/traces/inv-1"},
		"userMessage": "where am I right now?"
	}`, genai.RoleUser)

	attrs := attrMap(appendUserPromptAttr(nil, content))
	if got := attrs["user_prompt"]; got != "where am I right now?" {
		t.Fatalf("user_prompt = %v, want %q", got, "where am I right now?")
	}
}
