package uiactions

import (
	"fmt"
	"log"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

const ToolName = "issue_ui_actions"

type Action struct {
	Type     string `json:"type" jsonschema:"UI action type. Allowed values: navigate_recipe, navigate_recipe_list, navigate_trace, navigate_traces_list, refresh_current_screen."`
	RecipeID string `json:"recipeId,omitempty" jsonschema:"Required when type is navigate_recipe. Use the internal recipe ID returned by recipes-cli."`
	EventID  string `json:"eventId,omitempty" jsonschema:"Required when type is navigate_trace. The event_id returned by recipes-cli list-events."`
}

type Args struct {
	Actions []Action `json:"actions" jsonschema:"Actions the browser should execute after the agent finishes the current turn."`
}

type Result struct {
	Actions []Action `json:"actions"`
}

func NewTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        ToolName,
		Description: "Issues browser UI actions for the recipe app. Use after successful recipe creates, updates, deletes, imports, or explicit navigation requests.",
	}, issueUIActions)
}

func issueUIActions(ctx tool.Context, input Args) (Result, error) {
	result, err := Normalize(input)
	if err != nil {
		return result, err
	}
	log.Printf("tool %s: actions=%+v", ToolName, result.Actions)
	return result, nil
}

func Normalize(input Args) (Result, error) {
	actions := make([]Action, 0, len(input.Actions))
	for _, action := range input.Actions {
		normalized := Action{
			Type:     strings.TrimSpace(action.Type),
			RecipeID: strings.TrimSpace(action.RecipeID),
			EventID:  strings.TrimSpace(action.EventID),
		}
		switch normalized.Type {
		case "navigate_recipe":
			if normalized.RecipeID == "" {
				return Result{Actions: actions}, fmt.Errorf("navigate_recipe requires recipeId")
			}
			normalized.EventID = ""
			actions = append(actions, normalized)
		case "navigate_trace":
			if normalized.EventID == "" {
				return Result{Actions: actions}, fmt.Errorf("navigate_trace requires eventId")
			}
			normalized.RecipeID = ""
			actions = append(actions, normalized)
		case "navigate_recipe_list", "navigate_traces_list", "refresh_current_screen":
			normalized.RecipeID = ""
			normalized.EventID = ""
			actions = append(actions, normalized)
		default:
			return Result{Actions: actions}, fmt.Errorf("unsupported UI action type %q", normalized.Type)
		}
	}
	return Result{Actions: actions}, nil
}
