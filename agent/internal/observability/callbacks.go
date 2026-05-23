package observability

import (
	"log/slog"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	adktool "google.golang.org/adk/tool"
)

// ModelCallbacks returns before/after callbacks that emit llm.start and
// llm.end log lines. They never short-circuit the model call (always return
// nil response, nil error).
func ModelCallbacks() ([]llmagent.BeforeModelCallback, []llmagent.AfterModelCallback) {
	before := []llmagent.BeforeModelCallback{
		func(ctx agent.CallbackContext, _ *model.LLMRequest) (*model.LLMResponse, error) {
			slog.Debug("llm.start",
				"invocation_id", ctx.InvocationID(),
				"agent", ctx.AgentName(),
			)
			return nil, nil
		},
	}
	after := []llmagent.AfterModelCallback{
		func(ctx agent.CallbackContext, resp *model.LLMResponse, respErr error) (*model.LLMResponse, error) {
			attrs := []any{
				"invocation_id", ctx.InvocationID(),
				"agent", ctx.AgentName(),
				"error", respErr,
			}
			if resp != nil && resp.UsageMetadata != nil {
				u := resp.UsageMetadata
				attrs = append(attrs,
					"prompt_tokens", u.PromptTokenCount,
					"output_tokens", u.CandidatesTokenCount,
					"total_tokens", u.TotalTokenCount,
				)
			}
			slog.Debug("llm.end", attrs...)
			return nil, nil
		},
	}
	return before, after
}

// ToolCallbacks returns before/after callbacks that emit tool.start and
// tool.end log lines. They never alter the tool result (always return
// nil map, nil error).
func ToolCallbacks() ([]llmagent.BeforeToolCallback, []llmagent.AfterToolCallback) {
	before := []llmagent.BeforeToolCallback{
		func(ctx adktool.Context, t adktool.Tool, args map[string]any) (map[string]any, error) {
			slog.Info("tool.start",
				"invocation_id", ctx.InvocationID(),
				"session_id", ctx.SessionID(),
				"agent", ctx.AgentName(),
				"tool", t.Name(),
				"function_call_id", ctx.FunctionCallID(),
				"args", args,
			)
			return nil, nil
		},
	}
	after := []llmagent.AfterToolCallback{
		func(ctx adktool.Context, t adktool.Tool, args, result map[string]any, err error) (map[string]any, error) {
			slog.Info("tool.end",
				"invocation_id", ctx.InvocationID(),
				"tool", t.Name(),
				"args", args,
				"result", result,
				"error", err,
			)
			return nil, nil
		},
	}
	return before, after
}
