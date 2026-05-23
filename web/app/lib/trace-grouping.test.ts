import { describe, expect, it } from "vitest";

import type { Trace } from "./traces-api";
import { findUserPrompt, groupTraces, itemToolName } from "./trace-grouping";

function trace(
  id: string,
  msg: string,
  overrides: Record<string, unknown> = {},
): Trace {
  return {
    id,
    event_id: "inv-1",
    occurred_at: `2026-01-01T00:00:0${id.replace(/\D/g, "") || "0"}Z`,
    data: { msg, ...overrides },
  };
}

describe("groupTraces", () => {
  it("groups tool traces by function_call_id even when interleaved", () => {
    const items = groupTraces([
      trace("t1", "tool.start", {
        function_call_id: "call-a",
        tool: "call_recipes_cli",
      }),
      trace("t2", "tool.start", {
        function_call_id: "call-b",
        tool: "generate_recipe_photos",
      }),
      trace("t3", "tool.end", {
        function_call_id: "call-a",
        tool: "call_recipes_cli",
      }),
      trace("t4", "tool.end", {
        function_call_id: "call-b",
        tool: "generate_recipe_photos",
      }),
    ]);

    expect(items).toHaveLength(2);
    expect(items[0]).toMatchObject({
      kind: "toolGroup",
      functionCallId: "call-a",
    });
    expect(items[0].kind === "toolGroup" ? items[0].traces.map((t) => t.id) : []).toEqual([
      "t1",
      "t3",
    ]);
    expect(items[1]).toMatchObject({
      kind: "toolGroup",
      functionCallId: "call-b",
    });
    expect(items[1].kind === "toolGroup" ? items[1].traces.map((t) => t.id) : []).toEqual([
      "t2",
      "t4",
    ]);
  });

  it("falls back to adjacent tool grouping for traces without function_call_id", () => {
    const items = groupTraces([
      trace("t1", "tool.request", { tool: "legacy_tool" }),
      trace("t2", "tool.response", { tool: "legacy_tool" }),
      trace("t3", "agent.event"),
    ]);

    expect(items).toHaveLength(2);
    expect(items[0]).toMatchObject({
      kind: "toolGroup",
      functionCallId: null,
    });
    expect(items[0].kind === "toolGroup" ? items[0].traces.map((t) => t.id) : []).toEqual([
      "t1",
      "t2",
    ]);
    expect(items[1]).toMatchObject({ kind: "single" });
  });

  it("uses the first available tool name in a group", () => {
    const items = groupTraces([
      trace("t1", "tool.request", { function_call_id: "call-a" }),
      trace("t2", "tool.start", {
        function_call_id: "call-a",
        tool: "call_recipes_cli",
      }),
    ]);

    const item = items[0];
    expect(item.kind === "toolGroup" ? itemToolName(item) : "").toBe(
      "call_recipes_cli",
    );
  });
});

describe("findUserPrompt", () => {
  it("returns the first non-empty prompt from trace data", () => {
    expect(
      findUserPrompt([
        trace("t1", "agent.event", { user_prompt: "   " }),
        trace("t2", "tool.start", { user_prompt: "Find pasta recipes" }),
        trace("t3", "tool.end", { user_prompt: "Find dinner recipes" }),
      ]),
    ).toBe("Find pasta recipes");
  });

  it("extracts userMessage from prompt context JSON", () => {
    expect(
      findUserPrompt([
        trace("t1", "agent.event", {
          user_prompt: JSON.stringify({
            appContext: { screen: "other", path: "/traces/inv-1" },
            userMessage: "where am I right now?",
          }),
        }),
      ]),
    ).toBe("where am I right now?");
  });

  it("returns an empty string when no prompt is available", () => {
    expect(findUserPrompt([trace("t1", "agent.event")])).toBe("");
  });
});
