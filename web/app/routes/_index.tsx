import { ChefHat, Trash2 } from "lucide-react";
import { useEffect } from "react";
import {
  Link,
  useFetcher,
  useLoaderData,
  useNavigation,
  useRevalidator,
} from "react-router";

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

function formatDate(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function RecipesIndexContent() {
  const loaderData = useLoaderData<typeof loader>();
  const { state, dispatch } = useRecipesIndexState();
  const { recipes, listError, deletingId, deleteError } = state;
  const fetcher = useFetcher<typeof action>();
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

  useEffect(() => {
    if (fetcher.state !== "idle" || fetcher.data == null) return;
    const formData = fetcher.formData as FormData | undefined;
    const submittedId = formData?.get("id");
    if (typeof submittedId !== "string") return;
    if (fetcher.data.ok === true) {
      dispatch({
        type: RecipesIndexActionType.DELETE_SUCCEEDED,
        data: submittedId,
      });
    } else {
      dispatch({
        type: RecipesIndexActionType.DELETE_FAILED,
        data: fetcher.data.error,
      });
    }
  }, [fetcher.state, fetcher.data, fetcher.formData, dispatch]);

  function retryList() {
    dispatch({ type: RecipesIndexActionType.FETCH_STARTED });
    void revalidator.revalidate();
  }

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
        <div className="mt-8 flex flex-col gap-3">
          {deleteError ? (
            <div
              className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-200"
              role="alert"
            >
              <p>{deleteError}</p>
              <button
                type="button"
                className="mt-2 text-sm font-medium text-red-900 underline-offset-2 hover:underline dark:text-red-100"
                onClick={() =>
                  dispatch({ type: RecipesIndexActionType.DELETE_DISMISS })
                }
              >
                Dismiss
              </button>
            </div>
          ) : null}
          <ul className="flex flex-col gap-3">
            {recipes.map((r) => (
              <li
                key={r.id}
                className="flex gap-2 rounded-xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
              >
                <Link
                  to={`/recipe/${r.id}`}
                  className="flex min-w-0 flex-1 gap-4 p-4 outline-none transition-colors hover:bg-zinc-50/80 focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-zinc-400 dark:hover:bg-zinc-800/50 dark:focus-visible:ring-zinc-500"
                >
                  <div className="size-20 shrink-0 overflow-hidden rounded-lg bg-zinc-100 dark:bg-zinc-800">
                    {r.image.trim() !== "" ? (
                      <img
                        src={r.image}
                        alt=""
                        className="size-full object-cover"
                      />
                    ) : (
                      <div className="flex size-full items-center justify-center text-zinc-400 dark:text-zinc-500">
                        <ChefHat className="size-8 stroke-[1.5]" aria-hidden />
                      </div>
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-baseline gap-2">
                      <h3 className="truncate text-sm font-semibold text-zinc-900 dark:text-zinc-50">
                        {r.name}
                      </h3>
                      {r.category.trim() !== "" ? (
                        <span className="shrink-0 rounded-md bg-zinc-100 px-2 py-0.5 text-xs font-medium text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300">
                          {r.category}
                        </span>
                      ) : null}
                    </div>
                    {r.description.trim() !== "" ? (
                      <p className="mt-1 line-clamp-2 text-sm text-zinc-600 dark:text-zinc-400">
                        {r.description}
                      </p>
                    ) : null}
                    <p className="mt-2 text-xs text-zinc-500 dark:text-zinc-500">
                      Added {formatDate(r.created_at)}
                    </p>
                  </div>
                </Link>
                <div className="flex shrink-0 flex-col border-l border-zinc-100 dark:border-zinc-800">
                  <fetcher.Form
                    method="post"
                    className="flex flex-1 flex-col"
                    onSubmit={(e) => {
                      const ok = window.confirm(
                        `Delete “${r.name}”? This cannot be undone.`,
                      );
                      if (!ok) {
                        e.preventDefault();
                        return;
                      }
                      dispatch({
                        type: RecipesIndexActionType.DELETE_STARTED,
                        data: r.id,
                      });
                    }}
                  >
                    <input type="hidden" name="intent" value="delete" />
                    <input type="hidden" name="id" value={r.id} />
                    <button
                      type="submit"
                      className="flex flex-1 items-center justify-center px-3 text-zinc-400 transition-colors hover:bg-red-50 hover:text-red-700 disabled:pointer-events-none disabled:opacity-40 dark:hover:bg-red-950/40 dark:hover:text-red-300"
                      aria-label={`Delete ${r.name}`}
                      disabled={deletingId !== null}
                    >
                      {deletingId === r.id ? (
                        <span className="text-xs font-medium text-zinc-500 dark:text-zinc-400">
                          …
                        </span>
                      ) : (
                        <Trash2 className="size-4 stroke-[2]" aria-hidden />
                      )}
                    </button>
                  </fetcher.Form>
                </div>
              </li>
            ))}
          </ul>
        </div>
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
