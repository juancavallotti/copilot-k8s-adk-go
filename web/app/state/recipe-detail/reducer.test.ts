import { describe, expect, it } from "vitest";

import type { Recipe } from "~/lib/recipe-api";

import { recipeDetailReducer } from "./reducer";
import {
  RecipeDetailActionType,
  recipeDetailInitialState,
} from "./types";

function sampleRecipe(): Recipe {
  return {
    id: "abc",
    name: "Pie",
    description: "Good",
    category: "Dessert",
    image: "",
    ingredients: ["flour"],
    instructions: ["bake"],
    created_at: "2025-01-01T00:00:00Z",
    updated_at: "2025-01-01T00:00:00Z",
  };
}

describe("recipeDetailReducer", () => {
  it("LOAD_RESET clears recipe and error", () => {
    const next = recipeDetailReducer(
      { recipe: sampleRecipe(), error: "x" },
      { type: RecipeDetailActionType.LOAD_RESET },
    );
    expect(next.recipe).toBeNull();
    expect(next.error).toBeNull();
  });

  it("MISSING_ID sets error only", () => {
    const next = recipeDetailReducer(recipeDetailInitialState, {
      type: RecipeDetailActionType.MISSING_ID,
      data: "Missing recipe id.",
    });
    expect(next.error).toBe("Missing recipe id.");
    expect(next.recipe).toBeNull();
  });

  it("LOAD_SUCCESS sets recipe and clears error", () => {
    const r = sampleRecipe();
    const next = recipeDetailReducer(
      { recipe: null, error: "old" },
      { type: RecipeDetailActionType.LOAD_SUCCESS, data: r },
    );
    expect(next.recipe).toEqual(r);
    expect(next.error).toBeNull();
  });

  it("LOAD_FAILED sets error and clears recipe", () => {
    const next = recipeDetailReducer(
      { recipe: sampleRecipe(), error: null },
      {
        type: RecipeDetailActionType.LOAD_FAILED,
        data: "Not found",
      },
    );
    expect(next.recipe).toBeNull();
    expect(next.error).toBe("Not found");
  });
});
