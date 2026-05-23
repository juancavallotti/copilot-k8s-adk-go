import { ChefHat, Trash2 } from "lucide-react";
import { useEffect, useId, useMemo, useState } from "react";
import { Link, useFetcher } from "react-router";

import type { Recipe } from "~/lib/recipe-api";
import { getRecipeDisplayPhotos } from "~/lib/recipe-photos";
import { MarkdownView } from "./markdown-view";
import { RecipePhotoViewer } from "./recipe-photo-viewer";

type DeleteRecipeActionResult =
  | { ok: true }
  | { ok: false; error: string };

type RecipeSort = "newest" | "title-asc" | "title-desc";

export type RecipeListProps = {
  recipes: Recipe[];
  deletingId: string | null;
  deleteError: string | null;
  onDeleteStart: (id: string) => void;
  onDeleteSuccess: (id: string) => void;
  onDeleteFailure: (error: string) => void;
  onDeleteErrorDismiss: () => void;
};

function formatDate(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function RecipeList({
  recipes,
  deletingId,
  deleteError,
  onDeleteStart,
  onDeleteSuccess,
  onDeleteFailure,
  onDeleteErrorDismiss,
}: RecipeListProps) {
  const fetcher = useFetcher<DeleteRecipeActionResult>();
  const controlsId = useId();
  const [confirmingId, setConfirmingId] = useState<string | null>(null);
  const [filterText, setFilterText] = useState("");
  const [mealType, setMealType] = useState("");
  const [sortBy, setSortBy] = useState<RecipeSort>("newest");

  const mealTypes = useMemo(() => {
    const byNormalizedName = new Map<string, string>();
    for (const recipe of recipes) {
      const trimmed = recipe.category.trim();
      if (trimmed === "") continue;
      const key = trimmed.toLocaleLowerCase();
      if (!byNormalizedName.has(key)) {
        byNormalizedName.set(key, trimmed);
      }
    }
    return Array.from(byNormalizedName.values()).sort((a, b) =>
      a.localeCompare(b, undefined, { sensitivity: "base" }),
    );
  }, [recipes]);

  const visibleRecipes = useMemo(() => {
    const normalizedFilter = filterText.trim().toLocaleLowerCase();
    const normalizedMealType = mealType.trim().toLocaleLowerCase();

    const filtered = recipes.filter((recipe) => {
      const matchesMealType =
        normalizedMealType === "" ||
        recipe.category.trim().toLocaleLowerCase() === normalizedMealType;

      if (!matchesMealType) return false;
      if (normalizedFilter === "") return true;

      const searchableText = [
        recipe.name,
        recipe.description,
        recipe.category,
        ...recipe.ingredients,
      ]
        .join(" ")
        .toLocaleLowerCase();
      return searchableText.includes(normalizedFilter);
    });

    if (sortBy === "newest") {
      return filtered;
    }

    return [...filtered].sort((a, b) => {
      const result = a.name.localeCompare(b.name, undefined, {
        sensitivity: "base",
        numeric: true,
      });
      return sortBy === "title-asc" ? result : -result;
    });
  }, [filterText, mealType, recipes, sortBy]);

  const hasActiveFilters = filterText.trim() !== "" || mealType !== "";

  useEffect(() => {
    if (fetcher.state !== "idle" || fetcher.data == null) return;
    const formData = fetcher.formData as FormData | undefined;
    const submittedId = formData?.get("id");
    if (typeof submittedId !== "string") return;
    if (fetcher.data.ok === true) {
      onDeleteSuccess(submittedId);
    } else {
      onDeleteFailure(fetcher.data.error);
    }
  }, [
    fetcher.state,
    fetcher.data,
    fetcher.formData,
    onDeleteFailure,
    onDeleteSuccess,
  ]);

  return (
    <div className="mt-8 flex flex-col gap-3">
      <div className="rounded-xl border border-zinc-200 bg-white p-4 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
        <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_10rem_10rem]">
          <div className="space-y-1.5">
            <label
              htmlFor={`${controlsId}-filter`}
              className="text-xs font-medium text-zinc-600 dark:text-zinc-300"
            >
              Filter recipes
            </label>
            <input
              id={`${controlsId}-filter`}
              type="search"
              value={filterText}
              onChange={(e) => setFilterText(e.currentTarget.value)}
              placeholder="Search title, description, or ingredients"
              className="w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors placeholder:text-zinc-400 focus:border-zinc-400 focus:ring-2 focus:ring-zinc-200 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-100 dark:placeholder:text-zinc-500 dark:focus:border-zinc-500 dark:focus:ring-zinc-800"
            />
          </div>
          <div className="space-y-1.5">
            <label
              htmlFor={`${controlsId}-meal-type`}
              className="text-xs font-medium text-zinc-600 dark:text-zinc-300"
            >
              Meal type
            </label>
            <select
              id={`${controlsId}-meal-type`}
              value={mealType}
              onChange={(e) => setMealType(e.currentTarget.value)}
              className="w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-zinc-400 focus:ring-2 focus:ring-zinc-200 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-100 dark:focus:border-zinc-500 dark:focus:ring-zinc-800"
            >
              <option value="">All meal types</option>
              {mealTypes.map((type) => (
                <option key={type} value={type}>
                  {type}
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-1.5">
            <label
              htmlFor={`${controlsId}-sort`}
              className="text-xs font-medium text-zinc-600 dark:text-zinc-300"
            >
              Sort
            </label>
            <select
              id={`${controlsId}-sort`}
              value={sortBy}
              onChange={(e) => setSortBy(e.currentTarget.value as RecipeSort)}
              className="w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-zinc-400 focus:ring-2 focus:ring-zinc-200 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-100 dark:focus:border-zinc-500 dark:focus:ring-zinc-800"
            >
              <option value="newest">Newest first</option>
              <option value="title-asc">Title A-Z</option>
              <option value="title-desc">Title Z-A</option>
            </select>
          </div>
        </div>
        <div className="mt-3 flex flex-wrap items-center justify-between gap-2 text-xs text-zinc-500 dark:text-zinc-400">
          <span>
            Showing {visibleRecipes.length} of {recipes.length} recipes
          </span>
          {hasActiveFilters ? (
            <button
              type="button"
              className="font-medium text-zinc-700 underline-offset-2 hover:underline dark:text-zinc-200"
              onClick={() => {
                setFilterText("");
                setMealType("");
              }}
            >
              Clear filters
            </button>
          ) : null}
        </div>
      </div>
      {deleteError ? (
        <div
          className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-200"
          role="alert"
        >
          <p>{deleteError}</p>
          <button
            type="button"
            className="mt-2 text-sm font-medium text-red-900 underline-offset-2 hover:underline dark:text-red-100"
            onClick={onDeleteErrorDismiss}
          >
            Dismiss
          </button>
        </div>
      ) : null}
      {visibleRecipes.length === 0 ? (
        <div className="rounded-xl border border-dashed border-zinc-300 bg-zinc-50/80 p-8 text-center dark:border-zinc-700 dark:bg-zinc-900/40">
          <p className="text-sm text-zinc-600 dark:text-zinc-400">
            No recipes match those filters.
          </p>
        </div>
      ) : null}
      <ul className="flex flex-col gap-3">
        {visibleRecipes.map((recipe) => {
          const isConfirming = confirmingId === recipe.id;
          const isDeleting = deletingId === recipe.id;
          const displayPhotos = getRecipeDisplayPhotos(recipe);
          const primaryPhoto = displayPhotos[0] ?? null;

          return (
            <li
              key={recipe.id}
              className="flex gap-2 rounded-xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
            >
              <div className="shrink-0 p-4 pr-0">
                <div className="size-20 overflow-hidden rounded-lg bg-zinc-100 dark:bg-zinc-800">
                  {primaryPhoto != null ? (
                    <RecipePhotoViewer
                      photos={displayPhotos}
                      ariaLabel={`Open photos for ${recipe.name}`}
                      className="block size-full overflow-hidden rounded-lg outline-none transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:ring-zinc-400 dark:focus-visible:ring-zinc-500"
                    >
                      <img
                        src={primaryPhoto.src}
                        alt=""
                        className="size-full object-cover"
                      />
                    </RecipePhotoViewer>
                  ) : (
                    <div className="flex size-full items-center justify-center text-zinc-400 dark:text-zinc-500">
                      <ChefHat className="size-8 stroke-[1.5]" aria-hidden />
                    </div>
                  )}
                </div>
              </div>
              <Link
                to={`/recipe/${recipe.id}`}
                className="flex min-w-0 flex-1 p-4 pl-0 outline-none transition-colors hover:bg-zinc-50/80 focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-zinc-400 dark:hover:bg-zinc-800/50 dark:focus-visible:ring-zinc-500"
              >
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-baseline gap-2">
                    <h3 className="truncate text-sm font-semibold text-zinc-900 dark:text-zinc-50">
                      {recipe.name}
                    </h3>
                    {recipe.category.trim() !== "" ? (
                      <span className="shrink-0 rounded-md bg-zinc-100 px-2 py-0.5 text-xs font-medium text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300">
                        {recipe.category}
                      </span>
                    ) : null}
                  </div>
                  {recipe.description.trim() !== "" ? (
                    <MarkdownView
                      variant="preview"
                      className="mt-1 line-clamp-2 text-sm text-zinc-600 dark:text-zinc-400"
                    >
                      {recipe.description}
                    </MarkdownView>
                  ) : null}
                  <p className="mt-2 text-xs text-zinc-500 dark:text-zinc-500">
                    Added {formatDate(recipe.created_at)}
                  </p>
                </div>
              </Link>
              <div className="flex shrink-0 flex-col border-l border-zinc-100 dark:border-zinc-800">
                {isConfirming ? (
                  <fetcher.Form
                    method="post"
                    className="flex flex-1 flex-col justify-center gap-2 px-3 py-3"
                    onSubmit={() => {
                      setConfirmingId(null);
                      onDeleteStart(recipe.id);
                    }}
                  >
                    <input type="hidden" name="intent" value="delete" />
                    <input type="hidden" name="id" value={recipe.id} />
                    <span className="text-xs font-medium text-zinc-700 dark:text-zinc-300">
                      Delete?
                    </span>
                    <div className="flex gap-2">
                      <button
                        type="button"
                        className="rounded-md px-2 py-1 text-xs font-medium text-zinc-600 transition-colors hover:bg-zinc-100 hover:text-zinc-900 disabled:pointer-events-none disabled:opacity-40 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
                        onClick={() => setConfirmingId(null)}
                        disabled={deletingId !== null}
                      >
                        Cancel
                      </button>
                      <button
                        type="submit"
                        className="rounded-md bg-red-600 px-2 py-1 text-xs font-medium text-white transition-colors hover:bg-red-700 disabled:pointer-events-none disabled:opacity-40 dark:bg-red-700 dark:hover:bg-red-600"
                        disabled={deletingId !== null}
                      >
                        Confirm
                      </button>
                    </div>
                  </fetcher.Form>
                ) : (
                  <button
                    type="button"
                    className="flex flex-1 items-center justify-center px-3 text-zinc-400 transition-colors hover:bg-red-50 hover:text-red-700 disabled:pointer-events-none disabled:opacity-40 dark:hover:bg-red-950/40 dark:hover:text-red-300"
                    aria-label={`Delete ${recipe.name}`}
                    disabled={deletingId !== null}
                    onClick={() => setConfirmingId(recipe.id)}
                  >
                    {isDeleting ? (
                      <span className="text-xs font-medium text-zinc-500 dark:text-zinc-400">
                        ...
                      </span>
                    ) : (
                      <Trash2 className="size-4 stroke-[2]" aria-hidden />
                    )}
                  </button>
                )}
              </div>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
