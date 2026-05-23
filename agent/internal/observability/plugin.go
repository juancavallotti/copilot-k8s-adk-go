package observability

import (
	"log/slog"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/plugin"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// NewEventPlugin returns an ADK plugin whose OnEventCallback logs each
// non-partial session event as agent.event (with token usage, finish reason,
// and full model text), plus tool.request / tool.response for any
// function-call/response parts and state.delta for non-empty state updates.
// Streaming chunks (Partial=true) are skipped to keep logs readable.
func NewEventPlugin() (*plugin.Plugin, error) {
	return plugin.New(plugin.Config{
		Name: "observability",
		OnEventCallback: func(_ agent.InvocationContext, e *session.Event) (*session.Event, error) {
			if e == nil || e.Partial {
				return nil, nil
			}
			attrs := []any{
				"event_id", e.ID,
				"invocation_id", e.InvocationID,
				"author", e.Author,
				"branch", e.Branch,
				"finish_reason", string(e.FinishReason),
			}
			if u := e.UsageMetadata; u != nil {
				attrs = append(attrs,
					"prompt_tokens", u.PromptTokenCount,
					"output_tokens", u.CandidatesTokenCount,
					"total_tokens", u.TotalTokenCount,
				)
			}
			if text := contentText(e.Content); text != "" {
				attrs = append(attrs, "text", text)
			}
			slog.Info("agent.event", attrs...)

			if e.Content != nil {
				for _, p := range e.Content.Parts {
					if p == nil {
						continue
					}
					if call := p.FunctionCall; call != nil {
						slog.Info("tool.request",
							"invocation_id", e.InvocationID,
							"tool", call.Name,
							"args", call.Args,
						)
					}
					if resp := p.FunctionResponse; resp != nil {
						slog.Info("tool.response",
							"invocation_id", e.InvocationID,
							"tool", resp.Name,
							"response", resp.Response,
						)
					}
				}
			}
			if len(e.Actions.StateDelta) > 0 {
				slog.Info("state.delta",
					"invocation_id", e.InvocationID,
					"delta", e.Actions.StateDelta,
				)
			}
			return nil, nil
		},
	})
}

func contentText(c *genai.Content) string {
	if c == nil {
		return ""
	}
	var b strings.Builder
	for _, p := range c.Parts {
		if p != nil && p.Text != "" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}
