import { ChefHat } from "lucide-react";
import type { ReactNode } from "react";
import { Link } from "react-router";

import { toMarkdownPreview } from "~/lib/markdown-preview";
import type { Recipe } from "~/lib/recipe-api";
import { getRecipeDisplayPhotos } from "~/lib/recipe-photos";

import { MarkdownView } from "./markdown-view";
import { RecipePhotoViewer } from "./recipe-photo-viewer";

export type RecipeRowProps = {
  recipe: Recipe;
  // Right-side content (e.g. delete button, search score). The caller
  // owns its container styling.
  trailing?: ReactNode;
  // Footer text under the description (e.g. "Added Jan 5"). Hidden when null.
  footer?: ReactNode;
};

export function RecipeRow({ recipe, trailing, footer }: RecipeRowProps) {
  const displayPhotos = getRecipeDisplayPhotos(recipe);
  const primaryPhoto = displayPhotos[0] ?? null;
  const descriptionPreview = toMarkdownPreview(recipe.description);

  return (
    <li className="flex gap-2 rounded-xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
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
          {descriptionPreview !== "" ? (
            <MarkdownView
              variant="preview"
              className="mt-1 line-clamp-2 text-sm text-zinc-600 dark:text-zinc-400"
            >
              {descriptionPreview}
            </MarkdownView>
          ) : null}
          {footer != null ? (
            <div className="mt-2 text-xs text-zinc-500 dark:text-zinc-500">
              {footer}
            </div>
          ) : null}
        </div>
      </Link>
      {trailing}
    </li>
  );
}
