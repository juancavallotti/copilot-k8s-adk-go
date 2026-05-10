import type { RecipeDraft } from "~/lib/recipe-draft";

export type CreateRecipeState = {
  draft: RecipeDraft;
  submitting: boolean;
  error: string | null;
};

export const CreateRecipeActionType = {
  UPDATE_DRAFT: "UPDATE_DRAFT",
  SUBMIT_START: "SUBMIT_START",
  SUBMIT_ERROR: "SUBMIT_ERROR",
  RESET_FORM: "RESET_FORM",
} as const;

export type CreateRecipeAction =
  | { type: typeof CreateRecipeActionType.UPDATE_DRAFT; data: RecipeDraft }
  | { type: typeof CreateRecipeActionType.SUBMIT_START }
  | { type: typeof CreateRecipeActionType.SUBMIT_ERROR; data: string }
  | { type: typeof CreateRecipeActionType.RESET_FORM; data: RecipeDraft };
