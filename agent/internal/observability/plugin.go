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
		OnEventCallback: func(ctx agent.InvocationContext, e *session.Event) (*session.Event, error) {
			if e == nil || e.Partial {
				return nil, nil
			}
			attrs := eventCommonAttrs(ctx, e)
			attrs = append(attrs,
				"author", e.Author,
				"finish_reason", string(e.FinishReason),
				"turn_complete", e.TurnComplete,
				"interrupted", e.Interrupted,
			)
			if !e.Timestamp.IsZero() {
				attrs = append(attrs, "event_timestamp", e.Timestamp)
			}
			if e.ModelVersion != "" {
				attrs = append(attrs, "model_version", e.ModelVersion)
			}
			if e.ErrorCode != "" {
				attrs = append(attrs, "error_code", e.ErrorCode)
			}
			if e.ErrorMessage != "" {
				attrs = append(attrs, "error_message", e.ErrorMessage)
			}
			if len(e.LongRunningToolIDs) > 0 {
				attrs = append(attrs, "long_running_tool_ids", e.LongRunningToolIDs)
			}
			if len(e.Actions.ArtifactDelta) > 0 {
				attrs = append(attrs, "artifact_delta", e.Actions.ArtifactDelta)
			}
			if e.Actions.TransferToAgent != "" {
				attrs = append(attrs, "transfer_to_agent", e.Actions.TransferToAgent)
			}
			if e.Actions.Escalate {
				attrs = append(attrs, "escalate", true)
			}
			if e.Actions.SkipSummarization {
				attrs = append(attrs, "skip_summarization", true)
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
						attrs := append(eventCommonAttrs(ctx, e),
							"function_call_id", call.ID,
							"tool", call.Name,
							"args", call.Args,
						)
						slog.Info("tool.request", attrs...)
					}
					if resp := p.FunctionResponse; resp != nil {
						attrs := append(eventCommonAttrs(ctx, e),
							"function_call_id", resp.ID,
							"tool", resp.Name,
							"response", resp.Response,
						)
						slog.Info("tool.response", attrs...)
					}
				}
			}
			if len(e.Actions.StateDelta) > 0 {
				attrs := append(eventCommonAttrs(ctx, e),
					"delta", e.Actions.StateDelta,
				)
				slog.Info("state.delta", attrs...)
			}
			return nil, nil
		},
	})
}

func eventCommonAttrs(ctx agent.InvocationContext, e *session.Event) []any {
	attrs := []any{
		"event_id", e.ID,
		"invocation_id", e.InvocationID,
		"branch", e.Branch,
	}
	if ctx == nil {
		return attrs
	}
	if sess := ctx.Session(); sess != nil {
		attrs = append(attrs,
			"session_id", sess.ID(),
			"user_id", sess.UserID(),
			"app_name", sess.AppName(),
		)
	}
	if a := ctx.Agent(); a != nil {
		attrs = append(attrs, "agent", a.Name())
	}
	return attrs
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
