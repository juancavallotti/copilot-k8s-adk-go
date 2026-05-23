package observability

import (
	"log/slog"
	"sync"
	"time"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	adktool "google.golang.org/adk/tool"
	"google.golang.org/genai"
)

// ModelCallbacks returns before/after callbacks that emit llm.start and
// llm.end log lines. They never short-circuit the model call (always return
// nil response, nil error).
func ModelCallbacks() ([]llmagent.BeforeModelCallback, []llmagent.AfterModelCallback) {
	before := []llmagent.BeforeModelCallback{
		func(ctx agent.CallbackContext, _ *model.LLMRequest) (*model.LLMResponse, error) {
			slog.Debug("llm.start", callbackContextAttrs(ctx)...)
			return nil, nil
		},
	}
	after := []llmagent.AfterModelCallback{
		func(ctx agent.CallbackContext, resp *model.LLMResponse, respErr error) (*model.LLMResponse, error) {
			attrs := append(callbackContextAttrs(ctx),
				"error", respErr,
			)
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
	type traceKey struct {
		invocationID   string
		functionCallID string
	}
	var (
		mu     sync.Mutex
		starts = make(map[traceKey]time.Time)
	)

	before := []llmagent.BeforeToolCallback{
		func(ctx adktool.Context, t adktool.Tool, args map[string]any) (map[string]any, error) {
			key := traceKey{invocationID: ctx.InvocationID(), functionCallID: ctx.FunctionCallID()}
			mu.Lock()
			starts[key] = time.Now()
			mu.Unlock()

			attrs := append(toolContextAttrs(ctx),
				"tool", t.Name(),
				"function_call_id", ctx.FunctionCallID(),
				"args", args,
			)
			slog.Info("tool.start", attrs...)
			return nil, nil
		},
	}
	after := []llmagent.AfterToolCallback{
		func(ctx adktool.Context, t adktool.Tool, args, result map[string]any, err error) (map[string]any, error) {
			key := traceKey{invocationID: ctx.InvocationID(), functionCallID: ctx.FunctionCallID()}
			mu.Lock()
			start, ok := starts[key]
			delete(starts, key)
			mu.Unlock()

			attrs := append(toolContextAttrs(ctx),
				"tool", t.Name(),
				"function_call_id", ctx.FunctionCallID(),
				"args", args,
				"result", result,
				"error", err,
			)
			if ok {
				duration := time.Since(start).Round(time.Millisecond)
				attrs = append(attrs,
					"duration", duration,
					"duration_ms", duration.Milliseconds(),
				)
			}
			slog.Info("tool.end", attrs...)
			return nil, nil
		},
	}
	return before, after
}

type callbackTraceContext interface {
	InvocationID() string
	SessionID() string
	UserID() string
	AppName() string
	AgentName() string
	Branch() string
	UserContent() *genai.Content
}

func callbackContextAttrs(ctx callbackTraceContext) []any {
	attrs := []any{
		"invocation_id", ctx.InvocationID(),
		"session_id", ctx.SessionID(),
		"user_id", ctx.UserID(),
		"app_name", ctx.AppName(),
		"agent", ctx.AgentName(),
		"branch", ctx.Branch(),
	}
	return appendUserPromptAttr(attrs, ctx.UserContent())
}

func toolContextAttrs(ctx callbackTraceContext) []any {
	return callbackContextAttrs(ctx)
}
