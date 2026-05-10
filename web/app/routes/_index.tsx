import { ChefHat, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router";

import { type Recipe, deleteRecipe, listRecipes } from "~/lib/recipe-api";

import type { Route } from "./+types/_index";

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

export default function RecipesIndex() {
  const [recipes, setRecipes] = useState<Recipe[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setError(null);
    listRecipes()
      .then((data) => {
        if (!cancelled) setRecipes(data);
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Something went wrong");
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);

  async function handleDelete(recipe: Recipe) {
    const ok = window.confirm(
      `Delete “${recipe.name}”? This cannot be undone.`,
    );
    if (!ok) return;
    setDeleteError(null);
    setDeletingId(recipe.id);
    try {
      await deleteRecipe(recipe.id);
      setRecipes((prev) =>
        prev == null ? prev : prev.filter((r) => r.id !== recipe.id),
      );
    } catch (err) {
      setDeleteError(
        err instanceof Error ? err.message : "Could not delete recipe",
      );
    } finally {
      setDeletingId(null);
    }
  }

  return (
    <div className="mx-auto max-w-3xl">
      <h2 className="text-base font-medium text-zinc-900 dark:text-zinc-50">
        Recipes
      </h2>
      <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
        All recipes from your library, newest first.
      </p>

      {error ? (
        <div
          className="mt-8 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-800 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-200"
          role="alert"
        >
          <p>{error}</p>
          <button
            type="button"
            className="mt-3 text-sm font-medium text-red-900 underline-offset-2 hover:underline dark:text-red-100"
            onClick={() => {
              setRecipes(null);
              setError(null);
              listRecipes()
                .then(setRecipes)
                .catch((err) => {
                  setError(
                    err instanceof Error ? err.message : "Something went wrong",
                  );
                });
            }}
          >
            Try again
          </button>
        </div>
      ) : null}

      {!error && recipes === null ? (
        <p className="mt-8 text-sm text-zinc-500 dark:text-zinc-400">Loading…</p>
      ) : null}

      {!error && recipes !== null && recipes.length === 0 ? (
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

      {!error && recipes !== null && recipes.length > 0 ? (
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
                onClick={() => setDeleteError(null)}
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
                <button
                  type="button"
                  className="flex flex-1 items-center justify-center px-3 text-zinc-400 transition-colors hover:bg-red-50 hover:text-red-700 disabled:pointer-events-none disabled:opacity-40 dark:hover:bg-red-950/40 dark:hover:text-red-300"
                  aria-label={`Delete ${r.name}`}
                  disabled={deletingId !== null}
                  onClick={() => void handleDelete(r)}
                >
                  {deletingId === r.id ? (
                    <span className="text-xs font-medium text-zinc-500 dark:text-zinc-400">
                      …
                    </span>
                  ) : (
                    <Trash2 className="size-4 stroke-[2]" aria-hidden />
                  )}
                </button>
              </div>
            </li>
          ))}
          </ul>
        </div>
      ) : null}
    </div>
  );
}
