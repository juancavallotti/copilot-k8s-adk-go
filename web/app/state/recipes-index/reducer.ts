import {
  RecipesIndexActionType,
  type RecipesIndexAction,
  type RecipesIndexState,
  recipesIndexInitialState,
} from "./types";

export function recipesIndexReducer(
  state: RecipesIndexState = recipesIndexInitialState,
  action: RecipesIndexAction,
): RecipesIndexState {
  switch (action.type) {
    case RecipesIndexActionType.FETCH_STARTED:
      return {
        ...state,
        recipes: null,
        listError: null,
      };
    case RecipesIndexActionType.FETCH_SUCCESS:
      return {
        ...state,
        recipes: action.data,
        listError: null,
      };
    case RecipesIndexActionType.FETCH_FAILED:
      return {
        ...state,
        recipes: null,
        listError: action.data,
      };
    case RecipesIndexActionType.DELETE_STARTED:
      return {
        ...state,
        deletingId: action.data,
        deleteError: null,
      };
    case RecipesIndexActionType.DELETE_SUCCEEDED:
      return {
        ...state,
        recipes:
          state.recipes == null
            ? state.recipes
            : state.recipes.filter((r) => r.id !== action.data),
        deletingId: null,
      };
    case RecipesIndexActionType.DELETE_FAILED:
      return {
        ...state,
        deleteError: action.data,
        deletingId: null,
      };
    case RecipesIndexActionType.DELETE_DISMISS:
      return {
        ...state,
        deleteError: null,
      };
    default:
      return state;
  }
}
