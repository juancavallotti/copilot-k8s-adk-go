import type { Recipe } from "~/lib/recipe-api";

export type RecipesIndexState = {
  recipes: Recipe[] | null;
  listError: string | null;
  deletingId: string | null;
  deleteError: string | null;
};

export const RecipesIndexActionType = {
  FETCH_STARTED: "FETCH_STARTED",
  FETCH_SUCCESS: "FETCH_SUCCESS",
  FETCH_FAILED: "FETCH_FAILED",
  DELETE_STARTED: "DELETE_STARTED",
  DELETE_SUCCEEDED: "DELETE_SUCCEEDED",
  DELETE_FAILED: "DELETE_FAILED",
  DELETE_DISMISS: "DELETE_DISMISS",
} as const;

export type RecipesIndexAction =
  | { type: typeof RecipesIndexActionType.FETCH_STARTED }
  | { type: typeof RecipesIndexActionType.FETCH_SUCCESS; data: Recipe[] }
  | { type: typeof RecipesIndexActionType.FETCH_FAILED; data: string }
  | { type: typeof RecipesIndexActionType.DELETE_STARTED; data: string }
  | { type: typeof RecipesIndexActionType.DELETE_SUCCEEDED; data: string }
  | { type: typeof RecipesIndexActionType.DELETE_FAILED; data: string }
  | { type: typeof RecipesIndexActionType.DELETE_DISMISS };

export const recipesIndexInitialState: RecipesIndexState = {
  recipes: null,
  listError: null,
  deletingId: null,
  deleteError: null,
};
