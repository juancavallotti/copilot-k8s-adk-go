import { type FormEvent } from "react";
import { useNavigate } from "react-router";

import { RecipeEditor } from "~/components/recipe-editor";
import { createRecipe } from "~/lib/recipe-api";
import { draftToCreateBody, emptyRecipeDraft } from "~/lib/recipe-draft";
import {
  CreateRecipeProvider,
  useCreateRecipeState,
} from "~/state/create-recipe/context";
import { CreateRecipeActionType } from "~/state/create-recipe/types";

import type { Route } from "./+types/create";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Create recipe · Recipe manager" },
    { name: "description", content: "Add a new recipe" },
  ];
}

function CreateRecipeContent() {
  const navigate = useNavigate();
  const { state, dispatch } = useCreateRecipeState();
  const { draft, submitting, error } = state;

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    dispatch({ type: CreateRecipeActionType.SUBMIT_START });
    try {
      const created = await createRecipe(draftToCreateBody(draft));
      navigate(`/recipe/${created.id}`);
    } catch (err) {
      dispatch({
        type: CreateRecipeActionType.SUBMIT_ERROR,
        data:
          err instanceof Error ? err.message : "Something went wrong",
      });
    }
  }

  return (
    <div className="mx-auto max-w-3xl">
      <h2 className="text-base font-medium text-zinc-900 dark:text-zinc-50">
        Create recipe
      </h2>
      <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
        Fill in the details below. You can reuse this editor later when editing
        a recipe.
      </p>

      <form
        onSubmit={(e) => void handleSubmit(e)}
        className="mt-8 rounded-xl border border-zinc-200 bg-white p-6 shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
      >
        <RecipeEditor
          value={draft}
          onChange={(next) =>
            dispatch({ type: CreateRecipeActionType.UPDATE_DRAFT, data: next })
          }
          disabled={submitting}
        />

        {error ? (
          <p
            className="mt-6 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-200"
            role="alert"
          >
            {error}
          </p>
        ) : null}

        <div className="mt-8 flex flex-wrap items-center gap-3 border-t border-zinc-100 pt-6 dark:border-zinc-800">
          <button
            type="submit"
            disabled={submitting}
            className="inline-flex items-center justify-center rounded-lg bg-zinc-900 px-4 py-2.5 text-sm font-medium text-white shadow-sm transition-colors hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200"
          >
            {submitting ? "Saving…" : "Save recipe"}
          </button>
          <button
            type="button"
            disabled={submitting}
            className="text-sm font-medium text-zinc-600 underline-offset-2 hover:text-zinc-900 hover:underline dark:text-zinc-400 dark:hover:text-zinc-100"
            onClick={() =>
              dispatch({
                type: CreateRecipeActionType.RESET_FORM,
                data: emptyRecipeDraft(),
              })
            }
          >
            Reset form
          </button>
        </div>
      </form>
    </div>
  );
}

export default function CreateRecipe() {
  return (
    <CreateRecipeProvider>
      <CreateRecipeContent />
    </CreateRecipeProvider>
  );
}
