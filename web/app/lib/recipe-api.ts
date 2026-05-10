export type Recipe = {
  id: string;
  name: string;
  description: string;
  category: string;
  image: string;
  ingredients: string[];
  instructions: string[];
  created_at: string;
  updated_at: string;
};

export type CreateRecipeBody = {
  name: string;
  description: string;
  category: string;
  image: string;
  ingredients: string[];
  instructions: string[];
};

function normalizeApiBase(raw: string): string {
  return raw.replace(/\/$/, "");
}

/**
 * Resolves the recipes HTTP API base URL for the Node server (e.g. Docker / k8s).
 * Prefer `RECIPES_API_BASE` at runtime so the same image can target different services.
 */
export function resolveRecipesApiBaseFromEnv(): string {
  const fromEnv =
    typeof process !== "undefined" && process.env.RECIPES_API_BASE != null
      ? process.env.RECIPES_API_BASE.trim()
      : "";
  if (fromEnv !== "") return normalizeApiBase(fromEnv);

  const vite =
    (import.meta.env.VITE_RECIPES_API_BASE as string | undefined)?.trim() ||
    (import.meta.env.VITE_API_ORIGIN as string | undefined)?.trim() ||
    "";
  if (vite !== "") return normalizeApiBase(vite);

  return "/api";
}

function getRuntimeApiBaseFromGlobal(): string | undefined {
  const g = globalThis as typeof globalThis & { __RECIPES_API_BASE__?: string };
  const v = g.__RECIPES_API_BASE__;
  if (typeof v === "string" && v !== "") return normalizeApiBase(v);
  return undefined;
}

export function getApiBase(): string {
  return getRuntimeApiBaseFromGlobal() ?? resolveRecipesApiBaseFromEnv();
}

async function readJsonError(res: Response): Promise<Error> {
  const err = (await res.json().catch(() => null)) as { error?: string } | null;
  const msg =
    err != null && typeof err.error === "string" && err.error.length > 0
      ? err.error
      : `Request failed (${res.status})`;
  return new Error(msg);
}

export async function listRecipes(): Promise<Recipe[]> {
  const res = await fetch(`${getApiBase()}/recipes`);
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res.json() as Promise<Recipe[]>;
}

export async function getRecipe(id: string): Promise<Recipe> {
  const res = await fetch(`${getApiBase()}/recipes/${encodeURIComponent(id)}`);
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res.json() as Promise<Recipe>;
}

export async function createRecipe(body: CreateRecipeBody): Promise<Recipe> {
  const res = await fetch(`${getApiBase()}/recipes`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res.json() as Promise<Recipe>;
}

export async function deleteRecipe(id: string): Promise<void> {
  const res = await fetch(`${getApiBase()}/recipes/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (!res.ok) {
    throw await readJsonError(res);
  }
}

export async function replaceRecipe(recipe: Recipe): Promise<Recipe> {
  const res = await fetch(
    `${getApiBase()}/recipes/${encodeURIComponent(recipe.id)}`,
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
