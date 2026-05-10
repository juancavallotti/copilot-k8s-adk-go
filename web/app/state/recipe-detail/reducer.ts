import {
  RecipeDetailActionType,
  type RecipeDetailAction,
  type RecipeDetailState,
  recipeDetailInitialState,
} from "./types";

export function recipeDetailReducer(
  state: RecipeDetailState = recipeDetailInitialState,
  action: RecipeDetailAction,
): RecipeDetailState {
  switch (action.type) {
    case RecipeDetailActionType.LOAD_RESET:
      return { recipe: null, error: null };
    case RecipeDetailActionType.MISSING_ID:
      return { recipe: null, error: action.data };
    case RecipeDetailActionType.LOAD_SUCCESS:
      return { recipe: action.data, error: null };
    case RecipeDetailActionType.LOAD_FAILED:
      return { recipe: null, error: action.data };
    default:
      return state;
  }
}
