import type { Trace } from "./traces-api";

export type TraceItem =
  | { kind: "single"; trace: Trace }
  | { kind: "toolGroup"; traces: Trace[]; functionCallId: string | null };

export function getTraceMsg(data: unknown): string {
  return getStringField(data, "msg");
}

export function getToolName(data: unknown): string {
  return getStringField(data, "tool");
}

export function getFunctionCallId(data: unknown): string {
  return getStringField(data, "function_call_id");
}

export function getUserPrompt(data: unknown): string {
  return normalizeUserPrompt(getStringField(data, "user_prompt"));
}

export function findUserPrompt(traces: Trace[]): string {
  for (const trace of traces) {
    const prompt = getUserPrompt(trace.data).trim();
    if (prompt !== "") return prompt;
  }
  return "";
}

export function groupTraces(traces: Trace[]): TraceItem[] {
  const items: TraceItem[] = [];
  const groupsByCallId = new Map<string, Extract<TraceItem, { kind: "toolGroup" }>>();
  let pendingLegacyToolTraces: Trace[] = [];

  const flushLegacyToolTraces = () => {
    if (pendingLegacyToolTraces.length > 0) {
      items.push({
        kind: "toolGroup",
        traces: pendingLegacyToolTraces,
        functionCallId: null,
      });
      pendingLegacyToolTraces = [];
    }
  };

  for (const trace of traces) {
    if (!isToolTrace(trace)) {
      flushLegacyToolTraces();
      items.push({ kind: "single", trace });
      continue;
    }

    const functionCallId = getFunctionCallId(trace.data);
    if (functionCallId === "") {
      pendingLegacyToolTraces.push(trace);
      continue;
    }

    flushLegacyToolTraces();
    const existing = groupsByCallId.get(functionCallId);
    if (existing != null) {
      existing.traces.push(trace);
      continue;
    }

    const item: Extract<TraceItem, { kind: "toolGroup" }> = {
      kind: "toolGroup",
      traces: [trace],
      functionCallId,
    };
    groupsByCallId.set(functionCallId, item);
    items.push(item);
  }

  flushLegacyToolTraces();
  return items;
}

export function itemKey(item: TraceItem): string {
  return item.kind === "single" ? item.trace.id : item.traces[0].id;
}

export function itemTime(item: TraceItem): string {
  return item.kind === "single"
    ? item.trace.occurred_at
    : item.traces[0].occurred_at;
}

export function itemToolName(
  item: Extract<TraceItem, { kind: "toolGroup" }>,
): string {
  for (const t of item.traces) {
    const name = getToolName(t.data);
    if (name !== "") return name;
  }
  return "tool";
}

function isToolTrace(trace: Trace): boolean {
  return getTraceMsg(trace.data).startsWith("tool.");
}

function getStringField(data: unknown, field: string): string {
  if (data != null && typeof data === "object" && field in data) {
    const value = (data as Record<string, unknown>)[field];
    if (typeof value === "string") return value;
  }
  return "";
}

function normalizeUserPrompt(prompt: string): string {
  const raw = prompt.trim();
  if (raw === "") return "";
  try {
    const parsed: unknown = JSON.parse(raw);
    if (parsed != null && typeof parsed === "object" && "userMessage" in parsed) {
      const message = (parsed as { userMessage?: unknown }).userMessage;
      if (typeof message === "string" && message.trim() !== "") {
        return message.trim();
      }
    }
  } catch {
    // Older and newer traces may store plain text here; keep it as-is.
  }
  return raw;
}
