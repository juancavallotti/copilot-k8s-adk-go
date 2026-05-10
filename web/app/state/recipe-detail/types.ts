import type { Recipe } from "~/lib/recipe-api";

export type RecipeDetailState = {
  recipe: Recipe | null;
  error: string | null;
};

export const RecipeDetailActionType = {
  LOAD_RESET: "LOAD_RESET",
  MISSING_ID: "MISSING_ID",
  LOAD_SUCCESS: "LOAD_SUCCESS",
  LOAD_FAILED: "LOAD_FAILED",
} as const;

export type RecipeDetailAction =
  | { type: typeof RecipeDetailActionType.LOAD_RESET }
  | { type: typeof RecipeDetailActionType.MISSING_ID; data: string }
  | { type: typeof RecipeDetailActionType.LOAD_SUCCESS; data: Recipe }
  | { type: typeof RecipeDetailActionType.LOAD_FAILED; data: string };

export const recipeDetailInitialState: RecipeDetailState = {
  recipe: null,
  error: null,
};
