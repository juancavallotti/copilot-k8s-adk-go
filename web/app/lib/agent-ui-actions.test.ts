import { describe, expect, it } from "vitest";

import {
  extractUIActionsFromEvent,
  parseAssistantResponse,
  uniqueUIActions,
} from "./agent-ui-actions";

describe("agent UI actions", () => {
  it("parses hidden action directives from assistant text", () => {
    const parsed = parseAssistantResponse(
      'Created it.<ui_actions>{"actions":[{"type":"navigate_recipe","recipeId":"abc"}]}</ui_actions>',
    );

    expect(parsed.content).toBe("Created it.");
    expect(parsed.uiActions).toEqual([{ type: "navigate_recipe", recipeId: "abc" }]);
  });

  it("extracts actions from the issue_ui_actions tool response", () => {
    const actions = extractUIActionsFromEvent({
      content: {
        parts: [
          {
            functionResponse: {
              name: "issue_ui_actions",
              response: { actions: [{ type: "refresh_current_screen" }] },
            },
          },
        ],
      },
    });

    expect(actions).toEqual([{ type: "refresh_current_screen" }]);
  });

  it("deduplicates repeated tool and text actions", () => {
    expect(
      uniqueUIActions([
        { type: "refresh_current_screen" },
        { type: "refresh_current_screen" },
        { type: "navigate_recipe", recipeId: "abc" },
        { type: "navigate_recipe", recipeId: "abc" },
        { type: "navigate_trace", eventId: "inv-1" },
        { type: "navigate_trace", eventId: "inv-1" },
        { type: "navigate_trace", eventId: "inv-2" },
      ]),
    ).toEqual([
      { type: "refresh_current_screen" },
      { type: "navigate_recipe", recipeId: "abc" },
      { type: "navigate_trace", eventId: "inv-1" },
      { type: "navigate_trace", eventId: "inv-2" },
    ]);
  });

  it("parses navigate_trace and navigate_traces_list from hidden directives", () => {
    const parsed = parseAssistantResponse(
      'Done.<ui_actions>{"actions":[{"type":"navigate_trace","eventId":"inv-1"},{"type":"navigate_traces_list"}]}</ui_actions>',
    );
    expect(parsed.uiActions).toEqual([
      { type: "navigate_trace", eventId: "inv-1" },
      { type: "navigate_traces_list" },
    ]);
  });

  it("accepts snake_case event_id from agent payloads", () => {
    const parsed = parseAssistantResponse(
      'Done.<ui_actions>{"actions":[{"type":"navigate_trace","event_id":"inv-snake"}]}</ui_actions>',
    );
    expect(parsed.uiActions).toEqual([
      { type: "navigate_trace", eventId: "inv-snake" },
    ]);
  });

  it("drops navigate_trace without a usable eventId", () => {
    const parsed = parseAssistantResponse(
      'Done.<ui_actions>{"actions":[{"type":"navigate_trace"}]}</ui_actions>',
    );
    expect(parsed.uiActions).toEqual([]);
  });
});
