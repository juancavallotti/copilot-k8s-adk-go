import { type FormEvent, useEffect } from "react";
import {
  redirect,
  useActionData,
  useNavigation,
  useSubmit,
} from "react-router";

import { RecipeEditor } from "~/components/recipe-editor";
import type { CreateRecipeBody } from "~/lib/recipe-api";
import { draftToCreateBody, emptyRecipeDraft } from "~/lib/recipe-draft";
import {
  CreateRecipeProvider,
  useCreateRecipeState,
} from "~/state/create-recipe/context";
import { CreateRecipeActionType } from "~/state/create-recipe/types";

import type { Route } from "./+types/create";

export async function action({ request }: Route.ActionArgs) {
  const { createRecipe } = await import("~/lib/recipes-http.server");
  if (request.method !== "POST") {
    return null;
  }
  let body: CreateRecipeBody;
  try {
    body = (await request.json()) as CreateRecipeBody;
  } catch {
    return { ok: false as const, error: "Invalid request body." };
  }
  if (typeof body.name !== "string") {
    return { ok: false as const, error: "Invalid recipe data." };
  }
  try {
    const created = await createRecipe(request, body);
    return redirect(`/recipe/${created.id}`);
  } catch (err) {
    return {
      ok: false as const,
      error:
        err instanceof Error ? err.message : "Something went wrong",
    };
  }
}

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Create recipe · Recipe manager" },
    { name: "description", content: "Add a new recipe" },
  ];
}

function CreateRecipeContent() {
  const actionData = useActionData<typeof action>();
  const submit = useSubmit();
  const navigation = useNavigation();
  const { state, dispatch } = useCreateRecipeState();
  const { draft, submitting, error } = state;

  const navSubmitting =
    navigation.state === "submitting" &&
    navigation.location?.pathname === "/create";

  useEffect(() => {
    if (
      actionData != null &&
      typeof actionData === "object" &&
      "ok" in actionData &&
      actionData.ok === false
    ) {
      dispatch({
        type: CreateRecipeActionType.SUBMIT_ERROR,
        data: actionData.error,
      });
    }
  }, [actionData, dispatch]);

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    dispatch({ type: CreateRecipeActionType.SUBMIT_START });
    submit(draftToCreateBody(draft), {
      method: "POST",
      encType: "application/json",
    });
  }

  const busy = submitting || navSubmitting;

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
          disabled={busy}
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
            disabled={busy}
            className="inline-flex items-center justify-center rounded-lg bg-zinc-900 px-4 py-2.5 text-sm font-medium text-white shadow-sm transition-colors hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200"
          >
            {busy ? "Saving…" : "Save recipe"}
          </button>
          <button
            type="button"
            disabled={busy}
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
