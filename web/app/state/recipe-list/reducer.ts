import {
  RecipeListActionType,
  recipeListInitialState,
  type RecipeListAction,
  type RecipeListState,
} from "./types";

export function recipeListReducer(
  state: RecipeListState = recipeListInitialState,
  action: RecipeListAction,
): RecipeListState {
  switch (action.type) {
    case RecipeListActionType.CLEAR_FILTERS:
      return { ...state, filterText: "", mealType: "" };
    case RecipeListActionType.DELETE_FINISHED:
      return {
        ...state,
        handledDeleteResult: action.data,
        submittedDeleteId: null,
      };
    case RecipeListActionType.SET_CONFIRMING_ID:
      return { ...state, confirmingId: action.data };
    case RecipeListActionType.SET_FILTER_TEXT:
      return { ...state, filterText: action.data };
    case RecipeListActionType.SET_MEAL_TYPE:
      return { ...state, mealType: action.data };
    case RecipeListActionType.SET_SORT_BY:
      return { ...state, sortBy: action.data };
    case RecipeListActionType.SUBMIT_DELETE:
      return {
        ...state,
        confirmingId: null,
        submittedDeleteId: action.data,
      };
    default:
      return state;
  }
}
