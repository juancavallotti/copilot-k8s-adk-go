import { useLoaderData } from "react-router";

import { RecipeRow } from "~/components/recipe-row";
import type { RecipeMatch } from "~/lib/recipe-api";

import type { Route } from "./+types/search";

type LoaderData = {
  query: string;
  matches: RecipeMatch[];
  error: string | null;
  disabled: boolean;
};

export async function loader({ request }: Route.LoaderArgs): Promise<LoaderData> {
  const url = new URL(request.url);
  const query = (url.searchParams.get("q") ?? "").trim();
  if (query === "") {
    return { query: "", matches: [], error: null, disabled: false };
  }
  const { searchRecipes, SearchDisabledError } = await import(
    "~/lib/search-http.server"
  );
  try {
    const matches = await searchRecipes(request, query, { limit: 20 });
    return { query, matches, error: null, disabled: false };
  } catch (err) {
    if (err instanceof SearchDisabledError) {
      return { query, matches: [], error: null, disabled: true };
    }
    return {
      query,
      matches: [],
      error: err instanceof Error ? err.message : "Search failed",
      disabled: false,
    };
  }
}

export function meta({ data }: Route.MetaArgs) {
  const q = data?.query;
  return [
    {
      title: q != null && q !== "" ? `Search · ${q}` : "Search · Recipe manager",
    },
  ];
}

export default function SearchRoute() {
  const { query, matches, error, disabled } = useLoaderData<typeof loader>();

  if (query === "") {
    return (
      <div className="mx-auto max-w-3xl">
        <h2 className="text-base font-medium text-zinc-900 dark:text-zinc-50">
          Search
        </h2>
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
          Type a query in the top-right search box and hit Enter.
        </p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl">
      <h2 className="text-base font-medium text-zinc-900 dark:text-zinc-50">
        Search results
      </h2>
      <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
        {matches.length === 0
          ? `No matches for “${query}”.`
          : `${matches.length} result${matches.length === 1 ? "" : "s"} for “${query}”.`}
      </p>

      {disabled ? (
        <div
          className="mt-4 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-900/50 dark:bg-amber-950/40 dark:text-amber-200"
          role="alert"
        >
          Search is disabled — the backend has no embedding API key
          configured. Set <code>GEMINI_API_KEY</code> or{" "}
          <code>OPENAI_API_KEY</code> on the API and restart.
        </div>
      ) : null}

      {error != null ? (
        <div
          className="mt-4 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-800 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-200"
          role="alert"
        >
          {error}
        </div>
      ) : null}

      {matches.length > 0 ? (
        <ul className="mt-6 flex flex-col gap-3">
          {matches.map((m) => (
            <SearchResultRow key={m.id} match={m} />
          ))}
        </ul>
      ) : null}
    </div>
  );
}

function SearchResultRow({ match }: { match: RecipeMatch }) {
  const scorePct = Math.round(match.score * 100);
  return (
    <RecipeRow
      recipe={match}
      trailing={
        <div className="flex shrink-0 items-center pr-4">
          <span
            className="rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300"
            title={`Similarity score: ${match.score.toFixed(4)}`}
          >
            {scorePct}%
          </span>
        </div>
      }
    />
  );
}
