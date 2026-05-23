import type { ReactNode } from "react";

import { bootstrapProvider } from "@eetr/react-reducer-utils";

import { recipeListReducer } from "./reducer";
import { recipeListInitialState } from "./types";
import type { RecipeListAction, RecipeListState } from "./types";

const { Provider, useContextAccessors } = bootstrapProvider<
  RecipeListState,
  RecipeListAction
>(recipeListReducer, recipeListInitialState);

export function RecipeListProvider({ children }: { children: ReactNode }) {
  return <Provider>{children}</Provider>;
}

export { useContextAccessors as useRecipeListState };
