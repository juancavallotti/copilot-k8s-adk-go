import { ArrowLeft, Pencil } from "lucide-react";
import { useEffect } from "react";
import { Link, useLoaderData, useNavigation, useParams } from "react-router";

import { RecipeViewer } from "~/components/recipe-viewer";
import {
  RecipeDetailProvider,
  useRecipeDetailState,
} from "~/state/recipe-detail/context";
import { RecipeDetailActionType } from "~/state/recipe-detail/types";

import type { Route } from "./+types/recipe";

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

export function meta({ data }: Route.MetaArgs) {
  if (data?.recipe != null) {
    return [
      { title: `${data.recipe.name} · Recipe manager` },
      { name: "description", content: data.recipe.description },
    ];
  }
  return [
    { title: "Recipe · Recipe manager" },
    { name: "description", content: "View recipe details" },
  ];
}

function RecipeDetailContent() {
  const { id } = useParams();
  const loaderData = useLoaderData<typeof loader>();
  const { state, dispatch } = useRecipeDetailState();
  const { recipe, error } = state;
  const navigation = useNavigation();

  useEffect(() => {
    if (id == null || id === "") {
      dispatch({
        type: RecipeDetailActionType.MISSING_ID,
        data: "Missing recipe id.",
      });
      return;
    }
    dispatch({ type: RecipeDetailActionType.LOAD_RESET });
    if (loaderData.error) {
      dispatch({
        type: RecipeDetailActionType.LOAD_FAILED,
        data: loaderData.error,
      });
    } else if (
      loaderData.recipe != null &&
      loaderData.recipe.id === id
    ) {
      dispatch({
        type: RecipeDetailActionType.LOAD_SUCCESS,
        data: loaderData.recipe,
      });
    }
  }, [id, loaderData, dispatch]);

  const isPending =
    id != null &&
    navigation.state === "loading" &&
    navigation.location?.pathname === `/recipe/${id}` &&
    (recipe == null || recipe.id !== id);

  return (
    <div className="mx-auto max-w-3xl">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Link
          to="/"
          className="inline-flex items-center gap-2 text-sm font-medium text-zinc-600 underline-offset-2 hover:text-zinc-900 hover:underline dark:text-zinc-400 dark:hover:text-zinc-100"
        >
          <ArrowLeft className="size-4 stroke-[2]" aria-hidden />
          All recipes
        </Link>
        {!error && recipe !== null && id != null && id !== "" ? (
          <Link
            to={`/recipe/${id}/edit`}
            className="inline-flex items-center gap-2 rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm font-medium text-zinc-800 shadow-sm transition-colors hover:border-zinc-300 hover:bg-zinc-50 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-100 dark:hover:border-zinc-600 dark:hover:bg-zinc-800"
          >
            <Pencil className="size-4 stroke-[2]" aria-hidden />
            Edit
          </Link>
        ) : null}
      </div>

      {error ? (
        <div
          className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-800 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-200"
          role="alert"
        >
          <p>{error}</p>
        </div>
      ) : null}

      {!error && (recipe === null || isPending) ? (
        <p className="mt-8 text-sm text-zinc-500 dark:text-zinc-400">Loading…</p>
      ) : null}

      {!error && recipe !== null && !isPending ? (
        <div className="mt-6">
          <RecipeViewer recipe={recipe} />
        </div>
      ) : null}
    </div>
  );
}

export default function RecipeDetail() {
  return (
    <RecipeDetailProvider>
      <RecipeDetailContent />
    </RecipeDetailProvider>
  );
}
