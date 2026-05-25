import type { RecipeMatch } from "~/lib/recipe-api";

if (typeof window !== "undefined") {
  throw new Error(
    "search-http.server.ts is server-only; it must not run in the browser.",
  );
}

function normalizeApiBase(raw: string): string {
  return raw.replace(/\/$/, "");
}

function getApiBase(request: Request): string {
  const fromEnv =
    typeof process !== "undefined" && process.env.RECIPES_API_BASE != null
      ? process.env.RECIPES_API_BASE.trim()
      : "";
  if (fromEnv !== "") return normalizeApiBase(fromEnv);
  return normalizeApiBase(new URL("/api", request.url).toString());
}

/** Distinct error so the route can render a "configure an API key" hint
 * instead of the generic failure message. Matches the 503 the backend
 * returns when no GEMINI_API_KEY / OPENAI_API_KEY is set. */
export class SearchDisabledError extends Error {
  constructor() {
    super("search disabled: no embedding API key configured");
    this.name = "SearchDisabledError";
  }
}

async function readJsonError(res: Response): Promise<Error> {
  if (res.status === 503) return new SearchDisabledError();
  const err = (await res.json().catch(() => null)) as { error?: string } | null;
  const msg =
    err != null && typeof err.error === "string" && err.error.length > 0
      ? err.error
      : `Request failed (${res.status})`;
  return new Error(msg);
}

export async function searchRecipes(
  request: Request,
  query: string,
  opts?: { limit?: number },
): Promise<RecipeMatch[]> {
  const base = getApiBase(request);
  const params = new URLSearchParams({ q: query });
  if (opts?.limit != null) params.set("limit", String(opts.limit));
  const res = await fetch(`${base}/search/recipes?${params.toString()}`);
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res.json() as Promise<RecipeMatch[]>;
}
