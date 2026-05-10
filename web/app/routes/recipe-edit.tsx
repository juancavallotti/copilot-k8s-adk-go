import { type FormEvent, useEffect } from "react";
import {
  Link,
  redirect,
  useActionData,
  useLoaderData,
  useNavigation,
  useParams,
  useSubmit,
} from "react-router";

import { RecipeEditor } from "~/components/recipe-editor";
import type { Recipe } from "~/lib/recipe-api";
import { draftToRecipeForReplace } from "~/lib/recipe-draft";
import {
  EditRecipeProvider,
  useEditRecipeState,
} from "~/state/edit-recipe/context";
import { EditRecipeActionType } from "~/state/edit-recipe/types";

import type { Route } from "./+types/recipe-edit";

export async function loader({ request, params }: Route.LoaderArgs) {
  const { getRecipe } = await import("~/lib/recipes-http.server");
  const { id } = params;
  if (id == null || id === "") {
    return { recipe: null, error: "Missing recipe id." };
  }
  try {
    const recipe = await getRecipe(request, id);
    return { recipe, error: null as string | null };
  } catch (err) {
    return {
      recipe: null,
      error:
        err instanceof Error ? err.message : "Something went wrong",
    };
  }
}

export async function action({ request, params }: Route.ActionArgs) {
  const { replaceRecipe } = await import("~/lib/recipes-http.server");
  const { id } = params;
  if (id == null || id === "") {
    return { ok: false as const, error: "Missing recipe id." };
  }
  if (request.method !== "POST") {
    return null;
  }
  let recipePayload: Recipe;
  try {
    recipePayload = (await request.json()) as Recipe;
  } catch {
    return { ok: false as const, error: "Invalid request body." };
  }
  if (recipePayload.id !== id) {
    return { ok: false as const, error: "Recipe id mismatch." };
  }
  try {
    await replaceRecipe(request, recipePayload);
    return redirect(`/recipe/${id}`);
  } catch (err) {
    return {
      ok: false as const,
      error:
        err instanceof Error ? err.message : "Something went wrong",
    };
  }
}

export function meta({ data }: Route.MetaArgs) {
  if (data?.recipe != null) {
    return [
      { title: `Edit ${data.recipe.name} · Recipe manager` },
      { name: "description", content: "Update a recipe" },
    ];
  }
  return [
    { title: "Edit recipe · Recipe manager" },
    { name: "description", content: "Update a recipe" },
  ];
}

function RecipeEditContent() {
  const { id } = useParams();
  const loaderData = useLoaderData<typeof loader>();
  const actionData = useActionData<typeof action>();
  const submit = useSubmit();
  const navigation = useNavigation();
  const { state, dispatch } = useEditRecipeState();
  const { baseRecipe, draft, loadError, saveError, submitting } = state;

  useEffect(() => {
    if (id == null || id === "") {
      dispatch({
        type: EditRecipeActionType.MISSING_ID,
        data: "Missing recipe id.",
      });
      return;
    }
    dispatch({ type: EditRecipeActionType.LOAD_RESET });
    if (loaderData.error) {
      dispatch({
        type: EditRecipeActionType.LOAD_FAILED,
        data: loaderData.error,
      });
    } else if (
      loaderData.recipe != null &&
      loaderData.recipe.id === id
    ) {
      dispatch({
        type: EditRecipeActionType.LOAD_SUCCESS,
        data: loaderData.recipe,
      });
    }
  }, [id, loaderData, dispatch]);

  useEffect(() => {
    if (
      actionData != null &&
      typeof actionData === "object" &&
      "ok" in actionData &&
      actionData.ok === false
    ) {
      dispatch({
        type: EditRecipeActionType.SUBMIT_ERROR,
        data: actionData.error,
      });
    }
  }, [actionData, dispatch]);

  const navSubmitting =
    id != null &&
    navigation.state === "submitting" &&
    navigation.location?.pathname === `/recipe/${id}/edit`;

  const isPending =
    id != null &&
    navigation.state === "loading" &&
    navigation.location?.pathname === `/recipe/${id}/edit` &&
    (baseRecipe == null || baseRecipe.id !== id);

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (baseRecipe == null || id == null || id === "") return;
    dispatch({ type: EditRecipeActionType.SUBMIT_START });
    submit(draftToRecipeForReplace(baseRecipe, draft), {
      method: "POST",
      encType: "application/json",
    });
  }

  const busy = submitting || navSubmitting;

  if (loadError != null) {
    return (
      <div className="mx-auto max-w-3xl">
        <Link
          to="/"
          className="text-sm font-medium text-zinc-600 underline-offset-2 hover:text-zinc-900 hover:underline dark:text-zinc-400 dark:hover:text-zinc-100"
        >
          ← All recipes
        </Link>
        <div
          className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-800 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-200"
          role="alert"
        >
          <p>{loadError}</p>
        </div>
      </div>
    );
  }

  if (baseRecipe == null || isPending) {
    return (
      <div className="mx-auto max-w-3xl">
        <p className="text-sm text-zinc-500 dark:text-zinc-400">Loading…</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-base font-medium text-zinc-900 dark:text-zinc-50">
          Edit recipe
        </h2>
        <Link
          to={`/recipe/${id}`}
          className="text-sm font-medium text-zinc-600 underline-offset-2 hover:text-zinc-900 hover:underline dark:text-zinc-400 dark:hover:text-zinc-100"
        >
          Cancel
        </Link>
      </div>
      <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
        Changes are saved to your library when you click save.
      </p>

      <form
        onSubmit={(e) => void handleSubmit(e)}
        className="mt-8 rounded-xl border border-zinc-200 bg-white p-6 shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
      >
        <RecipeEditor
          value={draft}
          onChange={(next) =>
            dispatch({ type: EditRecipeActionType.UPDATE_DRAFT, data: next })
          }
          disabled={busy}
        />

        {saveError ? (
          <p
            className="mt-6 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-200"
            role="alert"
          >
            {saveError}
          </p>
        ) : null}

        <div className="mt-8 flex flex-wrap items-center gap-3 border-t border-zinc-100 pt-6 dark:border-zinc-800">
          <button
            type="submit"
            disabled={busy}
            className="inline-flex items-center justify-center rounded-lg bg-zinc-900 px-4 py-2.5 text-sm font-medium text-white shadow-sm transition-colors hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200"
          >
            {busy ? "Saving…" : "Save changes"}
          </button>
        </div>
      </form>
    </div>
  );
}

export default function RecipeEdit() {
  return (
    <EditRecipeProvider>
      <RecipeEditContent />
    </EditRecipeProvider>
  );
}
