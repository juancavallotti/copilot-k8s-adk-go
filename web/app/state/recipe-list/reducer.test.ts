import { describe, expect, it } from "vitest";

import { recipeListReducer } from "./reducer";
import {
  RecipeListActionType,
  recipeListInitialState,
  type RecipeListState,
} from "./types";

describe("recipeListReducer", () => {
  it("CLEAR_FILTERS resets filter fields only", () => {
    const state: RecipeListState = {
      ...recipeListInitialState,
      filterText: "soup",
      mealType: "Dinner",
      sortBy: "title-asc",
    };
    const next = recipeListReducer(state, {
      type: RecipeListActionType.CLEAR_FILTERS,
    });

    expect(next.filterText).toBe("");
    expect(next.mealType).toBe("");
    expect(next.sortBy).toBe("title-asc");
  });

  it("SUBMIT_DELETE stores the submitted id and clears confirmation", () => {
    const next = recipeListReducer(
      {
        ...recipeListInitialState,
        confirmingId: "recipe-1",
      },
      {
        type: RecipeListActionType.SUBMIT_DELETE,
        data: "recipe-1",
      },
    );

    expect(next.confirmingId).toBeNull();
    expect(next.submittedDeleteId).toBe("recipe-1");
  });

  it("DELETE_FINISHED stores handled result and clears submitted id", () => {
    const result = { ok: true } as const;
    const next = recipeListReducer(
      {
        ...recipeListInitialState,
        submittedDeleteId: "recipe-1",
      },
      {
        type: RecipeListActionType.DELETE_FINISHED,
        data: result,
      },
    );

    expect(next.handledDeleteResult).toBe(result);
    expect(next.submittedDeleteId).toBeNull();
  });
});
