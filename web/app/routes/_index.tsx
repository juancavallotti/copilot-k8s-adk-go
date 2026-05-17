import { useCallback, useEffect } from "react";
import { Link, useLoaderData, useNavigation, useRevalidator } from "react-router";

import { RecipeList } from "~/components/recipe-list";
import type { Recipe } from "~/lib/recipe-api";
import {
  RecipesIndexProvider,
  useRecipesIndexState,
} from "~/state/recipes-index/context";
import { RecipesIndexActionType } from "~/state/recipes-index/types";

import type { Route } from "./+types/_index";

export async function loader({ request }: Route.LoaderArgs) {
  const { listRecipes } = await import("~/lib/recipes-http.server");
  try {
    const recipes = await listRecipes(request);
    return { recipes, listError: null as string | null };
  } catch (err) {
    return {
      recipes: null as Recipe[] | null,
      listError:
        err instanceof Error ? err.message : "Something went wrong",
    };
  }
}

export async function action({ request }: Route.ActionArgs) {
  const { deleteRecipe } = await import("~/lib/recipes-http.server");
  const formData = await request.formData();
  const intent = formData.get("intent");
  if (intent !== "delete") {
    return { ok: false as const, error: "Unsupported action." };
  }
  const id = formData.get("id");
  if (typeof id !== "string" || id === "") {
    return { ok: false as const, error: "Missing recipe id." };
  }
  try {
    await deleteRecipe(request, id);
    return { ok: true as const };
  } catch (err) {
    return {
      ok: false as const,
      error:
        err instanceof Error ? err.message : "Could not delete recipe",
    };
  }
}

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Recipes · Recipe manager" },
    { name: "description", content: "Browse recipes" },
  ];
}

function RecipesIndexContent() {
  const loaderData = useLoaderData<typeof loader>();
  const { state, dispatch } = useRecipesIndexState();
  const { recipes, listError, deletingId, deleteError } = state;
  const navigation = useNavigation();
  const revalidator = useRevalidator();

  const isLoadingList =
    navigation.state === "loading" &&
    navigation.location?.pathname === "/" &&
    navigation.formMethod == null;

  useEffect(() => {
    if (loaderData.listError != null) {
      dispatch({
        type: RecipesIndexActionType.FETCH_FAILED,
        data: loaderData.listError,
      });
    } else if (loaderData.recipes != null) {
      dispatch({
        type: RecipesIndexActionType.FETCH_SUCCESS,
        data: loaderData.recipes,
      });
    }
  }, [loaderData, dispatch]);

  function retryList() {
    dispatch({ type: RecipesIndexActionType.FETCH_STARTED });
    void revalidator.revalidate();
  }

  const handleDeleteStart = useCallback(
    (id: string) =>
      dispatch({
        type: RecipesIndexActionType.DELETE_STARTED,
        data: id,
      }),
    [dispatch],
  );

  const handleDeleteSuccess = useCallback(
    (id: string) =>
      dispatch({
        type: RecipesIndexActionType.DELETE_SUCCEEDED,
        data: id,
      }),
    [dispatch],
  );

  const handleDeleteFailure = useCallback(
    (error: string) =>
      dispatch({
        type: RecipesIndexActionType.DELETE_FAILED,
        data: error,
      }),
    [dispatch],
  );

  const handleDeleteErrorDismiss = useCallback(
    () => dispatch({ type: RecipesIndexActionType.DELETE_DISMISS }),
    [dispatch],
  );

  return (
    <div className="mx-auto max-w-3xl">
      <h2 className="text-base font-medium text-zinc-900 dark:text-zinc-50">
        Recipes
      </h2>
      <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
        All recipes from your library, newest first.
      </p>

      {listError ? (
        <div
          className="mt-8 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-800 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-200"
          role="alert"
        >
          <p>{listError}</p>
          <button
            type="button"
            className="mt-3 text-sm font-medium text-red-900 underline-offset-2 hover:underline dark:text-red-100"
            onClick={retryList}
          >
            Try again
          </button>
        </div>
      ) : null}

      {!listError && (recipes === null || isLoadingList) ? (
        <p className="mt-8 text-sm text-zinc-500 dark:text-zinc-400">Loading…</p>
      ) : null}

      {!listError && recipes !== null && !isLoadingList && recipes.length === 0 ? (
        <div className="mt-8 rounded-xl border border-dashed border-zinc-300 bg-zinc-50/80 p-8 text-center dark:border-zinc-700 dark:bg-zinc-900/40">
          <p className="text-sm text-zinc-600 dark:text-zinc-400">
            No recipes yet. Create one to get started.
          </p>
          <Link
            to="/create"
            className="mt-4 inline-flex items-center justify-center rounded-lg bg-zinc-900 px-4 py-2.5 text-sm font-medium text-white shadow-sm transition-colors hover:bg-zinc-800 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200"
          >
            Create recipe
          </Link>
        </div>
      ) : null}

      {!listError && recipes !== null && !isLoadingList && recipes.length > 0 ? (
        <RecipeList
          recipes={recipes}
          deletingId={deletingId}
          deleteError={deleteError}
          onDeleteStart={handleDeleteStart}
          onDeleteSuccess={handleDeleteSuccess}
          onDeleteFailure={handleDeleteFailure}
          onDeleteErrorDismiss={handleDeleteErrorDismiss}
        />
      ) : null}
    </div>
  );
}

export default function RecipesIndex() {
  return (
    <RecipesIndexProvider>
      <RecipesIndexContent />
    </RecipesIndexProvider>
  );
}
