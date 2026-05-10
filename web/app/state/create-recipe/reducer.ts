import {
  CreateRecipeActionType,
  type CreateRecipeAction,
  type CreateRecipeState,
} from "./types";

export function createRecipeReducer(
  state: CreateRecipeState,
  action: CreateRecipeAction,
): CreateRecipeState {
  switch (action.type) {
    case CreateRecipeActionType.UPDATE_DRAFT:
      return { ...state, draft: action.data };
    case CreateRecipeActionType.SUBMIT_START:
      return { ...state, submitting: true, error: null };
    case CreateRecipeActionType.SUBMIT_ERROR:
      return { ...state, submitting: false, error: action.data };
    case CreateRecipeActionType.RESET_FORM:
      return { ...state, draft: action.data, error: null };
    default:
      return state;
  }
}
