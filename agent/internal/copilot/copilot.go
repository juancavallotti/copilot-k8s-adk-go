package copilot

import (
	"context"
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	adktool "google.golang.org/adk/tool"
	"google.golang.org/genai"

	"juancavallotti.com/recipes-agent/internal/config"
	"juancavallotti.com/recipes-agent/internal/imagegen"
	"juancavallotti.com/recipes-agent/internal/instruction"
	"juancavallotti.com/recipes-agent/internal/tools/recipephotos"
	"juancavallotti.com/recipes-agent/internal/tools/recipescli"
	"juancavallotti.com/recipes-agent/internal/tools/uiactions"
)

const agentName = "recipe_copilot"

func New(ctx context.Context, cfg config.Config) (agent.Agent, error) {
	model, err := gemini.NewModel(ctx, cfg.Model, &genai.ClientConfig{
		APIKey: cfg.GeminiAPIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini model: %w", err)
	}

	cliTool, err := recipescli.NewTool()
	if err != nil {
		return nil, fmt.Errorf("create recipes cli tool: %w", err)
	}
	imageGenerator, err := imagegen.NewGeminiRecipeImageGenerator(ctx, cfg.GeminiAPIKey, cfg.ImageModel)
	if err != nil {
		return nil, fmt.Errorf("create recipe image generator: %w", err)
	}
	photoTool, err := recipephotos.NewTool(imageGenerator, cfg.ImageGenerationConcurrency, cfg.ImageOutputDir)
	if err != nil {
		return nil, fmt.Errorf("create recipe photo tool: %w", err)
	}
	uiActionsTool, err := uiactions.NewTool()
	if err != nil {
		return nil, fmt.Errorf("create UI actions tool: %w", err)
	}
	instruction, err := instruction.Load(cfg.InstructionPath)
	if err != nil {
		return nil, fmt.Errorf("load instruction: %w", err)
	}

	a, err := llmagent.New(llmagent.Config{
		Name:        agentName,
		Model:       model,
		Description: "Recipe copilot that manages recipes by calling the installed recipes-cli.",
		Instruction: instruction,
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
