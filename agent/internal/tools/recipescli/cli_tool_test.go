package recipescli

import (
	"context"
	"testing"

	"google.golang.org/genai"
)

type fakeTraceContext struct {
	context.Context
}

func (fakeTraceContext) InvocationID() string   { return "inv-1" }
func (fakeTraceContext) SessionID() string      { return "session-1" }
func (fakeTraceContext) UserID() string         { return "user-1" }
func (fakeTraceContext) AppName() string        { return "recipes" }
func (fakeTraceContext) AgentName() string      { return "recipe_agent" }
func (fakeTraceContext) Branch() string         { return "root" }
func (fakeTraceContext) FunctionCallID() string { return "call-1" }
func (fakeTraceContext) UserContent() *genai.Content {
	return genai.NewContentFromText("List dinner recipes", genai.RoleUser)
}

func TestCLITraceAttrs(t *testing.T) {
	attrs := attrMap(cliTraceAttrs(fakeTraceContext{Context: context.Background()}))
	for key, want := range map[string]any{
		"invocation_id":    "inv-1",
		"session_id":       "session-1",
		"user_id":          "user-1",
		"app_name":         "recipes",
		"agent":            "recipe_agent",
		"branch":           "root",
		"function_call_id": "call-1",
		"tool":             "call_recipes_cli",
		"user_prompt":      "List dinner recipes",
	} {
		if got := attrs[key]; got != want {
			t.Fatalf("%s = %v, want %v", key, got, want)
		}
	}
}

func TestTraceUserPromptText_extractsUserMessageFromContextEnvelope(t *testing.T) {
	content := genai.NewContentFromText(`{
		"appContext": {"screen": "other", "path": "/traces/inv-1"},
		"userMessage": "where am I right now?"
	}`, genai.RoleUser)

	if got := traceUserPromptText(content); got != "where am I right now?" {
		t.Fatalf("traceUserPromptText = %q, want %q", got, "where am I right now?")
	}
}

func attrMap(attrs []any) map[string]any {
	out := make(map[string]any, len(attrs)/2)
	for i := 0; i+1 < len(attrs); i += 2 {
		key, ok := attrs[i].(string)
		if !ok {
			continue
		}
		out[key] = attrs[i+1]
	}
	return out
}
