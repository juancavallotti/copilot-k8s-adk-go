import { ChevronLeft, ChevronRight, X } from "lucide-react";
import type { ReactNode } from "react";
import { useEffect, useState } from "react";

import type { DisplayPhoto } from "~/lib/recipe-photos";

export type RecipePhotoViewerProps = {
  photos: DisplayPhoto[];
  initialIndex?: number;
  ariaLabel: string;
  className?: string;
  children: ReactNode;
};

export function RecipePhotoViewer({
  photos,
  initialIndex = 0,
  ariaLabel,
  className,
  children,
}: RecipePhotoViewerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(initialIndex);
  const activePhoto = photos[activeIndex] ?? photos[0] ?? null;
  const hasMultiplePhotos = photos.length > 1;

  useEffect(() => {
    if (!isOpen) return;

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setIsOpen(false);
      }
      if (event.key === "ArrowLeft" && hasMultiplePhotos) {
        setActiveIndex((index) => (index - 1 + photos.length) % photos.length);
      }
      if (event.key === "ArrowRight" && hasMultiplePhotos) {
        setActiveIndex((index) => (index + 1) % photos.length);
      }
    }

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [hasMultiplePhotos, isOpen, photos.length]);

  if (photos.length === 0) {
    return children;
  }

  function openViewer() {
    setActiveIndex(Math.min(initialIndex, photos.length - 1));
    setIsOpen(true);
  }

  function showPrevious() {
    setActiveIndex((index) => (index - 1 + photos.length) % photos.length);
  }

  function showNext() {
    setActiveIndex((index) => (index + 1) % photos.length);
  }

  return (
    <>
      <button
        type="button"
        className={className}
        aria-label={ariaLabel}
        onClick={openViewer}
      >
        {children}
      </button>

      {isOpen && activePhoto != null ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-zinc-950/85 p-4"
          role="dialog"
          aria-modal="true"
          aria-label="Recipe photo"
          onClick={() => setIsOpen(false)}
        >
          <div
            className="relative flex max-h-full w-full max-w-5xl flex-col items-center gap-4"
            onClick={(event) => event.stopPropagation()}
          >
            <button
              type="button"
              className="absolute right-3 top-3 rounded-full bg-white/90 p-2 text-zinc-900 shadow-lg transition-colors hover:bg-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white dark:bg-zinc-900/90 dark:text-zinc-50 dark:hover:bg-zinc-900 sm:right-4 sm:top-4"
              aria-label="Close photo viewer"
              onClick={() => setIsOpen(false)}
            >
              <X className="size-5" aria-hidden />
            </button>

            <img
              src={activePhoto.src}
              alt=""
              className="max-h-[82vh] max-w-full rounded-xl object-contain shadow-2xl ring-1 ring-white/10"
            />

            {hasMultiplePhotos ? (
              <>
                <button
                  type="button"
                  className="absolute left-3 top-1/2 -translate-y-1/2 rounded-full bg-white/90 p-2 text-zinc-900 shadow-lg transition-colors hover:bg-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white dark:bg-zinc-900/90 dark:text-zinc-50 dark:hover:bg-zinc-900 sm:left-4"
                  aria-label="Previous photo"
                  onClick={showPrevious}
                >
                  <ChevronLeft className="size-6" aria-hidden />
                </button>
                <button
                  type="button"
                  className="absolute right-3 top-1/2 -translate-y-1/2 rounded-full bg-white/90 p-2 text-zinc-900 shadow-lg transition-colors hover:bg-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white dark:bg-zinc-900/90 dark:text-zinc-50 dark:hover:bg-zinc-900 sm:right-4"
                  aria-label="Next photo"
                  onClick={showNext}
                >
                  <ChevronRight className="size-6" aria-hidden />
                </button>
              </>
            ) : null}

            <p className="rounded-full bg-zinc-950/70 px-3 py-1 text-xs font-medium text-white">
              Photo {activeIndex + 1} of {photos.length}
              {activePhoto.featured ? " · Featured" : ""}
            </p>
          </div>
        </div>
      ) : null}
    </>
  );
}
