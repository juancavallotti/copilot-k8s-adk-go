export type RecipeListDeleteResult =
  | { ok: true }
  | { ok: false; error: string };

export type RecipeSort = "newest" | "title-asc" | "title-desc";

export type RecipeListState = {
  confirmingId: string | null;
  filterText: string;
  handledDeleteResult: RecipeListDeleteResult | null;
  mealType: string;
  sortBy: RecipeSort;
  submittedDeleteId: string | null;
};

export const RecipeListActionType = {
  CLEAR_FILTERS: "CLEAR_FILTERS",
  DELETE_FINISHED: "DELETE_FINISHED",
  SET_CONFIRMING_ID: "SET_CONFIRMING_ID",
  SET_FILTER_TEXT: "SET_FILTER_TEXT",
  SET_MEAL_TYPE: "SET_MEAL_TYPE",
  SET_SORT_BY: "SET_SORT_BY",
  SUBMIT_DELETE: "SUBMIT_DELETE",
} as const;

export type RecipeListAction =
  | { type: typeof RecipeListActionType.CLEAR_FILTERS }
  | {
      type: typeof RecipeListActionType.DELETE_FINISHED;
      data: RecipeListDeleteResult;
    }
  | { type: typeof RecipeListActionType.SET_CONFIRMING_ID; data: string | null }
  | { type: typeof RecipeListActionType.SET_FILTER_TEXT; data: string }
  | { type: typeof RecipeListActionType.SET_MEAL_TYPE; data: string }
  | { type: typeof RecipeListActionType.SET_SORT_BY; data: RecipeSort }
  | { type: typeof RecipeListActionType.SUBMIT_DELETE; data: string };

export const recipeListInitialState: RecipeListState = {
  confirmingId: null,
  filterText: "",
  handledDeleteResult: null,
  mealType: "",
  sortBy: "newest",
  submittedDeleteId: null,
};
