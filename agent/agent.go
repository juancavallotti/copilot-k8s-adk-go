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

	a, err := llmagent.New(llmagent.Config{
		Name:        agentName,
		Model:       model,
		Description: "Recipe copilot that manages recipes by calling the installed recipes-cli.",
		Instruction: recipeCopilotInstruction,
		Tools: []adktool.Tool{
			cliTool,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create llm agent: %w", err)
	}
	return a, nil
}

const recipeCopilotInstruction = `You are a copilot for the recipe application.

You have access to exactly one operational tool: call_recipes_cli. It runs the installed recipes-cli binary in this container. Use that tool for recipe listing, inspection, creation, patching, importing, exporting, and schema discovery.

Before using recipes-cli for a user task, call call_recipes_cli with an empty args array to inspect the current help text. Use the help output and, when needed, the schema command to understand valid commands and JSON payloads. Do not guess unsupported CLI flags or commands.

When a command needs JSON input, prefer passing "-" as the CLI path and provide the JSON through the tool's stdin field. Keep JSON minimal and aligned with recipes-cli schema output. Report command failures clearly, including stderr when it helps the user recover.

Each user message is JSON with appContext and userMessage fields. appContext tells you the current UI location, and may include highlightedText from the browser selection. Use this context when deciding whether the user is referring to the recipe list, the current recipe, or selected text.

Recipe IDs and other internal identifiers are implementation details. You may use them for tool calls and inside the hidden <ui_actions> JSON directive, but do not include internal IDs in the user-visible prose. Refer to recipes by their human-readable names, descriptions, or positions in the conversation instead.

In addition to your normal chat response, always include one UI action directive at the very end of the response. The directive must be hidden from users by placing exactly one valid JSON object inside <ui_actions> tags:

<ui_actions>{"actions":[]}</ui_actions>

Allowed actions are:
- {"type":"navigate_recipe","recipeId":"RECIPE_ID"} to open a specific recipe.
- {"type":"navigate_recipe_list"} to open the recipe list.
- {"type":"refresh_current_screen"} to refresh the current screen after you create, update, delete, or import recipe data.

Use an empty actions array when no UI action is useful. Do not mention the <ui_actions> directive in your prose.`
