package copilot

import (
	"context"
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	adktool "google.golang.org/adk/tool"

	"juancavallotti.com/recipes-agent/internal/config"
	"juancavallotti.com/recipes-agent/internal/imagegen"
	"juancavallotti.com/recipes-agent/internal/observability"
	"juancavallotti.com/recipes-agent/internal/tools/recipephotos"
	"juancavallotti.com/recipes-agent/internal/tools/recipescli"
	"juancavallotti.com/recipes-agent/internal/tools/uiactions"
)

const AgentName = "recipe_copilot"

// NewWith builds an agent for one (chat model, image model) selection. The
// systemPrompt has already been assembled at process startup from the template
// file and the skill catalog; this function does not load it.
func NewWith(ctx context.Context, cfg config.Config, systemPrompt string, llm model.LLM, imgGen imagegen.RecipeImageGenerator) (agent.Agent, error) {
	cliTool, err := recipescli.NewTool()
	if err != nil {
		return nil, fmt.Errorf("create recipes cli tool: %w", err)
	}
	photoTool, err := recipephotos.NewTool(imgGen, cfg.ImageGenerationConcurrency, cfg.ImageOutputDir)
	if err != nil {
		return nil, fmt.Errorf("create recipe photo tool: %w", err)
	}
	uiActionsTool, err := uiactions.NewTool()
	if err != nil {
		return nil, fmt.Errorf("create UI actions tool: %w", err)
	}

	beforeModel, afterModel := observability.ModelCallbacks()
	beforeTool, afterTool := observability.ToolCallbacks()

	a, err := llmagent.New(llmagent.Config{
		Name:        AgentName,
		Model:       llm,
		Description: "Recipe copilot that manages recipes by calling the installed recipes-cli.",
		Instruction: systemPrompt,
		BeforeModelCallbacks: append(
			[]llmagent.BeforeModelCallback{newContextTrimCallback()},
			beforeModel...,
		),
		AfterModelCallbacks:  afterModel,
		BeforeToolCallbacks:  beforeTool,
		AfterToolCallbacks:   afterTool,
		Tools: []adktool.Tool{
			photoTool,
			cliTool,
			uiActionsTool,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create llm agent: %w", err)
	}
	return a, nil
}
