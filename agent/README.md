# Recipe Copilot Protocol

The web app talks to the recipe copilot through the ADK REST `run_sse`
endpoint. Each browser message is sent as a JSON string so the agent receives
both the user's text and the UI context needed to make navigation decisions.

## Browser To Agent

Each `newMessage.parts[0].text` value is a JSON object with this shape:

```json
{
  "appContext": {
    "screen": "recipe_list",
    "path": "/",
    "recipeId": "optional-current-recipe-id",
    "highlightedText": "optional selected browser text"
  },
  "userMessage": "What should I cook tonight?"
}
```

`appContext.screen` is one of:

- `recipe_list`: the user is looking at the recipe list.
- `specific_recipe`: the user is looking at a recipe detail or edit screen.
- `create_recipe`: the user is creating a recipe.
- `other`: the browser is on another app route.

`highlightedText` is omitted when the browser selection is empty. When present,
the web app trims whitespace and caps the value before sending it to the agent.

## Agent To Browser

The agent response is normal markdown for chat output, followed by exactly one
hidden UI action directive:

```text
Here is the recipe I found.

<ui_actions>{"actions":[{"type":"navigate_recipe","recipeId":"abc123"}]}</ui_actions>
```

The browser strips the `<ui_actions>` block from the visible chat message,
parses the JSON, and executes the requested actions after the response stream
finishes.

Internal IDs are not user-facing. The agent can use recipe IDs in tool calls
and in the hidden `<ui_actions>` directive, but visible chat prose should refer
to recipes by human-readable names or conversational context.

Use an empty actions array when no UI action is useful:

```text
That recipe already looks complete.

<ui_actions>{"actions":[]}</ui_actions>
```

## Supported Actions

Open a specific recipe:

```json
{ "type": "navigate_recipe", "recipeId": "abc123" }
```

Open the recipe list:

```json
{ "type": "navigate_recipe_list" }
```

Refresh the current screen:

```json
{ "type": "refresh_current_screen" }
```

Use `refresh_current_screen` after creating, updating, deleting, or importing
recipe data so the browser reloads the active route's loader data.
