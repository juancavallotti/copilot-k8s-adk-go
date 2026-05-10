import { describe, expect, it } from "vitest";

import type { Recipe } from "~/lib/recipe-api";

import { recipesIndexReducer } from "./reducer";
import {
  RecipesIndexActionType,
  recipesIndexInitialState,
  type RecipesIndexState,
} from "./types";

function recipe(overrides: Partial<Recipe> = {}): Recipe {
  return {
    id: "1",
    name: "Soup",
    description: "",
    category: "Lunch",
    image: "",
    ingredients: [],
    instructions: [],
    created_at: "2025-01-01T00:00:00Z",
    updated_at: "2025-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("recipesIndexReducer", () => {
  it("FETCH_STARTED clears list fields and keeps delete-related state", () => {
    const state: RecipesIndexState = {
      recipes: [recipe()],
      listError: "boom",
      deletingId: null,
      deleteError: "delete failed",
    };
    const next = recipesIndexReducer(state, {
      type: RecipesIndexActionType.FETCH_STARTED,
    });
    expect(next.recipes).toBeNull();
    expect(next.listError).toBeNull();
    expect(next.deleteError).toBe("delete failed");
  });

  it("FETCH_SUCCESS stores recipes and clears list error", () => {
    const next = recipesIndexReducer(
      { ...recipesIndexInitialState, listError: "old" },
      {
        type: RecipesIndexActionType.FETCH_SUCCESS,
        data: [recipe({ id: "a" }), recipe({ id: "b", name: "Stew" })],
      },
    );
    expect(next.recipes).toHaveLength(2);
    expect(next.listError).toBeNull();
  });

  it("FETCH_FAILED sets message and clears recipes", () => {
    const next = recipesIndexReducer(
      { ...recipesIndexInitialState, recipes: [recipe()] },
      {
        type: RecipesIndexActionType.FETCH_FAILED,
        data: "network",
      },
    );
    expect(next.recipes).toBeNull();
    expect(next.listError).toBe("network");
  });

  it("DELETE_SUCCEEDED removes the recipe and clears deletingId", () => {
    const a = recipe({ id: "a" });
    const b = recipe({ id: "b", name: "Other" });
    const next = recipesIndexReducer(
      {
        ...recipesIndexInitialState,
        recipes: [a, b],
        deletingId: "a",
      },
      {
        type: RecipesIndexActionType.DELETE_SUCCEEDED,
        data: "a",
      },
    );
    expect(next.recipes).toEqual([b]);
    expect(next.deletingId).toBeNull();
  });

  it("DELETE_SUCCEEDED leaves recipes null when list was not loaded", () => {
    const next = recipesIndexReducer(
      {
        ...recipesIndexInitialState,
        recipes: null,
        deletingId: "x",
      },
      {
        type: RecipesIndexActionType.DELETE_SUCCEEDED,
        data: "x",
      },
    );
    expect(next.recipes).toBeNull();
    expect(next.deletingId).toBeNull();
  });

  it("DELETE_DISMISS clears deleteError only", () => {
    const next = recipesIndexReducer(
      {
        ...recipesIndexInitialState,
        deleteError: "nope",
        deletingId: null,
      },
      { type: RecipesIndexActionType.DELETE_DISMISS },
    );
    expect(next.deleteError).toBeNull();
  });
});
