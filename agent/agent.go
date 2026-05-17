package main

import (
	"context"
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	adktool "google.golang.org/adk/tool"
	"google.golang.org/genai"
)

const agentName = "recipe_copilot"

func newRecipeCopilot(ctx context.Context, cfg config) (agent.Agent, error) {
	model, err := gemini.NewModel(ctx, cfg.Model, &genai.ClientConfig{
		APIKey: cfg.GeminiAPIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini model: %w", err)
	}

	cliTool, err := newRecipesCLITool()
	if err != nil {
		return nil, fmt.Errorf("create recipes cli tool: %w", err)
	}
	imageGenerator, err := newGeminiRecipeImageGenerator(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create recipe image generator: %w", err)
	}
	photoTool, err := newGenerateRecipePhotosTool(imageGenerator, cfg.ImageGenerationConcurrency)
	if err != nil {
		return nil, fmt.Errorf("create recipe photo tool: %w", err)
	}
	instruction, err := loadInstruction(cfg.InstructionPath)
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
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create llm agent: %w", err)
	}
	return a, nil
}
