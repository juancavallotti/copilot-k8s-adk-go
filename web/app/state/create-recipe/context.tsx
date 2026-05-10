import type { ReactNode } from "react";

import { bootstrapProvider } from "@eetr/react-reducer-utils";

import { emptyRecipeDraft } from "~/lib/recipe-draft";

import { createRecipeReducer } from "./reducer";
import type { CreateRecipeAction, CreateRecipeState } from "./types";

const createRecipeInitialState: CreateRecipeState = {
  draft: emptyRecipeDraft(),
  submitting: false,
  error: null,
};

const { Provider, useContextAccessors } = bootstrapProvider<
  CreateRecipeState,
  CreateRecipeAction
>(createRecipeReducer, createRecipeInitialState);

export function CreateRecipeProvider({ children }: { children: ReactNode }) {
  return <Provider>{children}</Provider>;
}

export { useContextAccessors as useCreateRecipeState };
