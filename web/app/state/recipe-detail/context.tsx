import type { ReactNode } from "react";

import { bootstrapProvider } from "@eetr/react-reducer-utils";

import { recipeDetailReducer } from "./reducer";
import type { RecipeDetailAction, RecipeDetailState } from "./types";
import { recipeDetailInitialState } from "./types";

const { Provider, useContextAccessors } = bootstrapProvider<
  RecipeDetailState,
  RecipeDetailAction
>(recipeDetailReducer, recipeDetailInitialState);

export function RecipeDetailProvider({ children }: { children: ReactNode }) {
  return <Provider>{children}</Provider>;
}

export { useContextAccessors as useRecipeDetailState };
