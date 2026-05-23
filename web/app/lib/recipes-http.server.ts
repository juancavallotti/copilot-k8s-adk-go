import type {
  CreateRecipeBody,
  Recipe,
  RecipePatchBody,
} from "~/lib/recipe-api";

if (typeof window !== "undefined") {
  throw new Error(
    "recipes-http.server.ts is server-only; it must not run in the browser.",
  );
}

function normalizeApiBase(raw: string): string {
  return raw.replace(/\/$/, "");
}

/**
 * Base URL for recipes HTTP calls on the **Node server** (loaders/actions).
 *
 * - Prefer `RECIPES_API_BASE` (cluster / Docker), so the browser never needs that value.
 * - If unset, use same-origin `/api` relative to the incoming request (Vite dev proxies
 *   `/api` → the API; production needs a reverse proxy or RECIPES_API_BASE).
 */
function getApiBase(request: Request): string {
  const fromEnv =
    typeof process !== "undefined" && process.env.RECIPES_API_BASE != null
      ? process.env.RECIPES_API_BASE.trim()
      : "";
  if (fromEnv !== "") return normalizeApiBase(fromEnv);
  return normalizeApiBase(new URL("/api", request.url).toString());
}

async function readJsonError(res: Response): Promise<Error> {
  const err = (await res.json().catch(() => null)) as { error?: string } | null;
  const msg =
    err != null && typeof err.error === "string" && err.error.length > 0
      ? err.error
      : `Request failed (${res.status})`;
  return new Error(msg);
}

export async function listRecipes(request: Request): Promise<Recipe[]> {
  const base = getApiBase(request);
  const res = await fetch(`${base}/recipes`);
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res.json() as Promise<Recipe[]>;
}

export async function getRecipe(
  request: Request,
  id: string,
): Promise<Recipe> {
  const base = getApiBase(request);
  const res = await fetch(`${base}/recipes/${encodeURIComponent(id)}`);
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res.json() as Promise<Recipe>;
}

export async function createRecipe(
  request: Request,
  body: CreateRecipeBody,
): Promise<Recipe> {
  const base = getApiBase(request);
  const res = await fetch(`${base}/recipes`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res.json() as Promise<Recipe>;
}

export async function deleteRecipe(
  request: Request,
  id: string,
): Promise<void> {
  const base = getApiBase(request);
  const res = await fetch(`${base}/recipes/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (!res.ok) {
    throw await readJsonError(res);
  }
}

export async function replaceRecipe(
  request: Request,
  recipe: Recipe,
): Promise<Recipe> {
  const base = getApiBase(request);
  const res = await fetch(
    `${base}/recipes/${encodeURIComponent(recipe.id)}`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(recipe),
    },
  );
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res.json() as Promise<Recipe>;
}

export async function patchRecipe(
  request: Request,
  id: string,
  patch: RecipePatchBody,
): Promise<Recipe> {
  const base = getApiBase(request);
  const res = await fetch(`${base}/recipes/${encodeURIComponent(id)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(patch),
  });
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res.json() as Promise<Recipe>;
}

export type RestoreResult = {
  imported: number;
  failed: Array<{ id: string; error: string }>;
};

export async function downloadBackup(request: Request): Promise<Response> {
  const base = getApiBase(request);
  const res = await fetch(`${base}/recipes/export`);
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res;
}

export async function uploadBackup(
  request: Request,
  file: File,
): Promise<RestoreResult> {
  const base = getApiBase(request);
  const fd = new FormData();
  fd.set("file", file);
  const res = await fetch(`${base}/recipes/import`, {
    method: "POST",
    body: fd,
  });
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return (await res.json()) as RestoreResult;
}
