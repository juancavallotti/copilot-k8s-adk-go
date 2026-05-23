export type UIAction =
  | { type: "navigate_recipe"; recipeId: string }
  | { type: "navigate_recipe_list" }
  | { type: "navigate_trace"; eventId: string }
  | { type: "navigate_traces_list" }
  | { type: "refresh_current_screen" };

export type ToolCallStatus = "pending" | "success" | "error";

export type ToolCall = {
  id: string;
  name: string;
  args?: Record<string, unknown>;
  response?: unknown;
  status: ToolCallStatus;
  summary?: string;
};

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
    if (action.type === "navigate_trace") {
      const eventId = action.eventId ?? action.event_id;
      return typeof eventId === "string" && eventId.trim() !== ""
        ? [{ type: "navigate_trace", eventId: eventId.trim() }]
        : [];
    }
    if (action.type === "navigate_traces_list") {
      return [{ type: "navigate_traces_list" }];
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

export function extractToolEventsFromEvent(event: AgentEventWithParts): {
  calls: Array<{ id: string; name: string; args?: Record<string, unknown> }>;
  responses: Array<{ id: string; name: string; response: unknown }>;
} {
  const parts = event.content?.parts ?? [];
  const calls: Array<{ id: string; name: string; args?: Record<string, unknown> }> = [];
  const responses: Array<{ id: string; name: string; response: unknown }> = [];
  for (const part of parts) {
    if (part == null || typeof part !== "object") continue;
    const call = getRecord(part, "functionCall", "function_call");
    if (call != null) {
      const id = getString(call, "id");
      const name = getString(call, "name");
      if (id != null && name != null) {
        const args = getRecord(call, "args");
        calls.push({ id, name, args });
      }
    }
    const response = getRecord(part, "functionResponse", "function_response");
    if (response != null) {
      const id = getString(response, "id");
      const name = getString(response, "name");
      if (id != null && name != null) {
        responses.push({ id, name, response: response.response });
      }
    }
  }
  return { calls, responses };
}

export function isInternalToolName(name: string): boolean {
  return name === uiActionsToolName;
}

export function makePendingToolCall(call: {
  id: string;
  name: string;
  args?: Record<string, unknown>;
}): ToolCall {
  return {
    id: call.id,
    name: call.name,
    args: call.args,
    status: "pending",
    summary: summarizeToolCall(call.name, call.args),
  };
}

export function applyToolResponse(toolCall: ToolCall, response: unknown): ToolCall {
  return {
    ...toolCall,
    response,
    status: inferStatus(response),
  };
}

function summarizeToolCall(
  _name: string,
  args: Record<string, unknown> | undefined,
): string | undefined {
  if (args == null) return undefined;
  const command = args.command;
  if (typeof command === "string" && command !== "") return command;
  const query = args.query;
  if (typeof query === "string" && query !== "") return query;
  return undefined;
}

function inferStatus(response: unknown): ToolCallStatus {
  if (response == null || typeof response !== "object") return "success";
  const record = response as Record<string, unknown>;
  if (record.successful === false) return "error";
  if (typeof record.exitCode === "number" && record.exitCode !== 0) return "error";
  if (typeof record.error === "string" && record.error !== "") return "error";
  return "success";
}

export function uniqueUIActions(actions: UIAction[]): UIAction[] {
  const seen = new Set<string>();
  return actions.filter((action) => {
    let key: string;
    if (action.type === "navigate_recipe") {
      key = `${action.type}:${action.recipeId}`;
    } else if (action.type === "navigate_trace") {
      key = `${action.type}:${action.eventId}`;
    } else {
      key = action.type;
    }
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
  return typeof child === "string" && child !== "" ? child : undefined;
}
