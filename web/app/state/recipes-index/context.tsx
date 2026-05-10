import type { ReactNode } from "react";

import { bootstrapProvider } from "@eetr/react-reducer-utils";

import { recipesIndexReducer } from "./reducer";
import { recipesIndexInitialState } from "./types";
import type { RecipesIndexAction, RecipesIndexState } from "./types";

const { Provider, useContextAccessors } = bootstrapProvider<
  RecipesIndexState,
  RecipesIndexAction
>(recipesIndexReducer, recipesIndexInitialState);

export function RecipesIndexProvider({ children }: { children: ReactNode }) {
  return <Provider>{children}</Provider>;
}

export { useContextAccessors as useRecipesIndexState };
