import { describe, expect, it } from "vitest";

import {
  isInternalToolName,
  parseAssistantResponse,
  uniqueUIActions,
  type UIAction,
} from "./agent-ui-actions";
import { readAgentStream, type AgentEvent } from "./agent-stream";
import { createToolContextStore } from "./tool-context";

type Bubble = {
  id: string;
  role: "user" | "assistant";
  content: string;
  uiActions?: UIAction[];
  toolCallIds?: string[];
};

function makeSSEResponse(events: AgentEvent[]): Response {
  const encoder = new TextEncoder();
  const stream = new ReadableStream<Uint8Array>({
    start(controller) {
      for (const event of events) {
        controller.enqueue(
          encoder.encode(`data: ${JSON.stringify(event)}\n\n`),
        );
      }
      controller.close();
    },
  });
  return new Response(stream, { status: 200 });
}

let idCounter = 0;
function nextID(prefix: string): string {
  idCounter += 1;
  return `${prefix}-${idCounter}`;
}

/**
 * Mirrors the bubble-list orchestration in sendMessage. Drives readAgentStream
 * with a fresh empty assistant bubble seeded, and returns the final bubble
 * list plus the tool store so tests can assert ordering and attachments.
 */
async function runScenario(events: AgentEvent[]) {
  idCounter = 0;
  const toolStore = createToolContextStore();
  let bubbles: Bubble[] = [];
  let currentAssistantID = nextID("assistant");
  bubbles.push({ id: currentAssistantID, role: "assistant", content: "" });

  let pendingNextBubble = false;
  let lastFinalizedBubbleID: string | null = null;

  function ensureCurrentBubble() {
    if (!pendingNextBubble) return;
    currentAssistantID = nextID("assistant");
    bubbles.push({ id: currentAssistantID, role: "assistant", content: "" });
    pendingNextBubble = false;
  }

  function updateBubble(id: string, patch: Partial<Bubble>) {
    bubbles = bubbles.map((b) => (b.id === id ? { ...b, ...patch } : b));
  }

  await readAgentStream(makeSSEResponse(events), {
    onTextProgress: (chunk, streamUIActions) => {
      ensureCurrentBubble();
      const id = currentAssistantID;
      const parsed = parseAssistantResponse(chunk);
      const uiActions = uniqueUIActions([
        ...streamUIActions,
        ...parsed.uiActions,
      ]);
      updateBubble(id, { content: parsed.content, uiActions });
    },
    onToolEvents: ({ calls, responses }) => {
      for (const call of calls) {
        if (isInternalToolName(call.name)) continue;
        const isNew = toolStore.register(call);
        if (!isNew) continue;
        ensureCurrentBubble();
        const owner = currentAssistantID;
        const target = bubbles.find((b) => b.id === owner);
        const nextIds = [...(target?.toolCallIds ?? []), call.id];
        updateBubble(owner, { toolCallIds: nextIds });
      }
      for (const r of responses) {
        if (isInternalToolName(r.name)) continue;
        toolStore.applyResponse(r.id, r.response);
      }
    },
    onTurnFinalize: (chunk, streamUIActions) => {
      const id = currentAssistantID;
      const parsed = parseAssistantResponse(chunk);
      const uiActions = uniqueUIActions([
        ...streamUIActions,
        ...parsed.uiActions,
      ]);
      updateBubble(id, { content: parsed.content, uiActions });
      lastFinalizedBubbleID = id;
      pendingNextBubble = true;
    },
    onTurnAuthoritativeText: (chunk, streamUIActions, _rawEvents) => {
      const targetId = lastFinalizedBubbleID ?? currentAssistantID;
      const parsed = parseAssistantResponse(chunk);
      const uiActions = uniqueUIActions([
        ...streamUIActions,
        ...parsed.uiActions,
      ]);
      updateBubble(targetId, { content: parsed.content, uiActions });
    },
  });

  return { bubbles, toolStore };
}

describe("readAgentStream message ordering", () => {
  it("renders a single text turn as one bubble with the streamed content", async () => {
    const { bubbles } = await runScenario([
      { partial: true, content: { parts: [{ text: "Hel" }] } },
      { partial: true, content: { parts: [{ text: "Hello" }] } },
      { partial: true, content: { parts: [{ text: "Hello world" }] } },
      { turnComplete: true, content: { parts: [{ text: "Hello world" }] } },
    ]);

    expect(bubbles).toHaveLength(1);
    expect(bubbles[0]).toMatchObject({
      role: "assistant",
      content: "Hello world",
    });
    expect(bubbles[0].toolCallIds ?? []).toEqual([]);
  });

  it("silently replaces the streamed bubble when the authoritative re-emit arrives", async () => {
    const { bubbles } = await runScenario([
      { partial: true, content: { parts: [{ text: "Hello" }] } },
      { turnComplete: true, content: { parts: [{ text: "Hello" }] } },
      // re-emit with a more authoritative variant (e.g. trailing period)
      {
        turnComplete: true,
        content: { parts: [{ text: "Hello world." }] },
      },
    ]);

    expect(bubbles).toHaveLength(1);
    expect(bubbles[0].content).toBe("Hello world.");
  });

  it("attaches tool widgets to the bubble that registered the call (in order)", async () => {
    const { bubbles, toolStore } = await runScenario([
      // Turn 1: text
      { partial: true, content: { parts: [{ text: "Let me search." }] } },
      {
        turnComplete: true,
        content: { parts: [{ text: "Let me search." }] },
      },
      // Tool call arrives in a new bubble
      {
        content: {
          parts: [
            {
              functionCall: {
                id: "tool-1",
                name: "search",
                args: { q: "cake" },
              },
            },
          ],
        },
      },
      // Tool response — silently updates store, must not create a new bubble
      {
        content: {
          parts: [
            {
              functionResponse: {
                id: "tool-1",
                name: "search",
                response: { results: ["cake recipe"] },
              },
            },
          ],
        },
      },
      // Turn 2: assistant follow-up text in the same bubble as the tool
      { partial: true, content: { parts: [{ text: "Found it!" }] } },
      { turnComplete: true, content: { parts: [{ text: "Found it!" }] } },
    ]);

    expect(bubbles.map((b) => b.content)).toEqual([
      "Let me search.",
      "Found it!",
    ]);
    expect(bubbles[0].toolCallIds ?? []).toEqual([]);
    expect(bubbles[1].toolCallIds).toEqual(["tool-1"]);
    expect(toolStore.get("tool-1")?.status).toBe("success");
  });

  it("preserves bubble order across multiple turns each with their own tool", async () => {
    const { bubbles, toolStore } = await runScenario([
      // Turn 1: tool call, no preceding text
      {
        content: {
          parts: [
            { functionCall: { id: "a", name: "list_recipes", args: {} } },
          ],
        },
      },
      {
        content: {
          parts: [
            {
              functionResponse: {
                id: "a",
                name: "list_recipes",
                response: { items: [] },
              },
            },
          ],
        },
      },
      { partial: true, content: { parts: [{ text: "None yet." }] } },
      { turnComplete: true, content: { parts: [{ text: "None yet." }] } },
      // Turn 2: another tool, attached to a new bubble
      {
        content: {
          parts: [
            { functionCall: { id: "b", name: "create_recipe", args: {} } },
          ],
        },
      },
      {
        content: {
          parts: [
            {
              functionResponse: {
                id: "b",
                name: "create_recipe",
                response: { id: "r-1" },
              },
            },
          ],
        },
      },
      { partial: true, content: { parts: [{ text: "Done." }] } },
      { turnComplete: true, content: { parts: [{ text: "Done." }] } },
    ]);

    expect(bubbles).toHaveLength(2);
    expect(bubbles[0]).toMatchObject({
      content: "None yet.",
      toolCallIds: ["a"],
    });
    expect(bubbles[1]).toMatchObject({
      content: "Done.",
      toolCallIds: ["b"],
    });
    expect(toolStore.get("a")?.status).toBe("success");
    expect(toolStore.get("b")?.status).toBe("success");
  });

  it("ignores duplicate tool-call registrations (no second widget)", async () => {
    const { bubbles, toolStore } = await runScenario([
      {
        content: {
          parts: [
            { functionCall: { id: "tool-1", name: "search", args: {} } },
          ],
        },
      },
      // Same call re-emitted (ADK sometimes does this) — must be a no-op
      {
        content: {
          parts: [
            { functionCall: { id: "tool-1", name: "search", args: {} } },
          ],
        },
      },
      {
        content: {
          parts: [
            {
              functionResponse: {
                id: "tool-1",
                name: "search",
                response: { ok: true },
              },
            },
          ],
        },
      },
      { partial: true, content: { parts: [{ text: "Done." }] } },
      { turnComplete: true, content: { parts: [{ text: "Done." }] } },
    ]);

    expect(bubbles).toHaveLength(1);
    expect(bubbles[0].toolCallIds).toEqual(["tool-1"]);
    expect(toolStore.list()).toHaveLength(1);
  });

  it("does not emit a second bubble when the only text event is the re-emit", async () => {
    // Regression for the old merge-hack scenario: turn finalizes with text,
    // then ADK re-emits the same content — must not produce a duplicate bubble.
    const { bubbles } = await runScenario([
      { partial: true, content: { parts: [{ text: "Created it." }] } },
      {
        turnComplete: true,
        content: { parts: [{ text: "Created it." }] },
      },
      {
        turnComplete: true,
        content: { parts: [{ text: "Created it." }] },
      },
    ]);

    expect(bubbles).toHaveLength(1);
    expect(bubbles[0].content).toBe("Created it.");
  });
});
