export type UIAction =
  | { type: "navigate_recipe"; recipeId: string }
  | { type: "navigate_recipe_list" }
  | { type: "refresh_current_screen" };

export type ParsedAssistantResponse = {
  content: string;
  uiActions: UIAction[];
};

export type AgentEventWithParts = {
  content?: {
    parts?: unknown[];
  };
};

const uiActionsToolName = "issue_ui_actions";

export function parseAssistantResponse(raw: string): ParsedAssistantResponse {
  const uiActions: UIAction[] = [];
  let hasCompleteActionBlock = false;
  const completeBlock = /<ui_actions>\s*([\s\S]*?)\s*<\/ui_actions>/gi;
  for (const match of raw.matchAll(completeBlock)) {
    hasCompleteActionBlock = true;
    try {
      uiActions.push(...normalizeUIActions(JSON.parse(match[1])));
    } catch {
      // Ignore malformed action directives and keep the chat output usable.
    }
  }

  const content = raw
    .replace(/\s*<ui_actions>[\s\S]*?(?:<\/ui_actions>|$)/gi, "")
    .trim();

  return {
    content: content === "" && hasCompleteActionBlock ? "Done." : content,
    uiActions: uniqueUIActions(uiActions),
  };
}

export function normalizeUIActions(value: unknown): UIAction[] {
  const rawActions =
    value != null &&
    typeof value === "object" &&
    "actions" in value &&
    Array.isArray((value as { actions?: unknown }).actions)
      ? (value as { actions: unknown[] }).actions
      : Array.isArray(value)
        ? value
        : [];

  return rawActions.flatMap((rawAction): UIAction[] => {
    if (rawAction == null || typeof rawAction !== "object") return [];
    const action = rawAction as Record<string, unknown>;
    if (action.type === "navigate_recipe") {
      const recipeId = action.recipeId ?? action.recipe_id;
      return typeof recipeId === "string" && recipeId.trim() !== ""
        ? [{ type: "navigate_recipe", recipeId: recipeId.trim() }]
        : [];
    }
    if (action.type === "navigate_recipe_list") {
      return [{ type: "navigate_recipe_list" }];
    }
    if (action.type === "refresh_current_screen") {
      return [{ type: "refresh_current_screen" }];
    }
    return [];
  });
}

export function extractUIActionsFromEvent(event: AgentEventWithParts): UIAction[] {
  const parts = event.content?.parts ?? [];
  return uniqueUIActions(
    parts.flatMap((part) => {
      if (part == null || typeof part !== "object") return [];
      const functionResponse = getRecord(part, "functionResponse", "function_response");
      if (functionResponse == null) return [];
      const name = getString(functionResponse, "name");
      if (name !== uiActionsToolName) return [];
      return normalizeUIActions(getRecord(functionResponse, "response"));
    }),
  );
}

export function uniqueUIActions(actions: UIAction[]): UIAction[] {
  const seen = new Set<string>();
  return actions.filter((action) => {
    const key =
      action.type === "navigate_recipe"
        ? `${action.type}:${action.recipeId}`
        : action.type;
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}

function getRecord(
  value: unknown,
  ...keys: string[]
): Record<string, unknown> | undefined {
  if (value == null || typeof value !== "object") return undefined;
  const record = value as Record<string, unknown>;
  for (const key of keys) {
    const child = record[key];
    if (child != null && typeof child === "object") {
      return child as Record<string, unknown>;
    }
  }
  return undefined;
}

function getString(value: Record<string, unknown>, key: string): string | undefined {
  const child = value[key];
  return typeof child === "string" ? child : undefined;
}
