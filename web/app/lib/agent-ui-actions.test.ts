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
      ]),
    ).toEqual([
      { type: "refresh_current_screen" },
      { type: "navigate_recipe", recipeId: "abc" },
    ]);
  });
});
