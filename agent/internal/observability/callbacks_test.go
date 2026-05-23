package observability

import "testing"

type fakeToolTraceContext struct{}

func (fakeToolTraceContext) InvocationID() string { return "inv-1" }
func (fakeToolTraceContext) SessionID() string    { return "session-1" }
func (fakeToolTraceContext) UserID() string       { return "user-1" }
func (fakeToolTraceContext) AppName() string      { return "recipes" }
func (fakeToolTraceContext) AgentName() string    { return "recipe_agent" }
func (fakeToolTraceContext) Branch() string       { return "root" }

func TestToolContextAttrs(t *testing.T) {
	attrs := attrMap(toolContextAttrs(fakeToolTraceContext{}))
	for key, want := range map[string]any{
		"invocation_id": "inv-1",
		"session_id":    "session-1",
		"user_id":       "user-1",
		"app_name":      "recipes",
		"agent":         "recipe_agent",
		"branch":        "root",
	} {
		if got := attrs[key]; got != want {
			t.Fatalf("%s = %v, want %v", key, got, want)
		}
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
