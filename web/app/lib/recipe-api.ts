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

export function getApiBase(): string {
  const v = import.meta.env.VITE_API_ORIGIN as string | undefined;
  if (v != null && v !== "") return v.replace(/\/$/, "");
  return "/api";
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
