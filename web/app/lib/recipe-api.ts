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

export async function createRecipe(body: CreateRecipeBody): Promise<Recipe> {
  const res = await fetch(`${getApiBase()}/recipes`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const err = (await res.json().catch(() => null)) as { error?: string } | null;
    const msg =
      err != null && typeof err.error === "string" && err.error.length > 0
        ? err.error
        : `Request failed (${res.status})`;
    throw new Error(msg);
  }
  return res.json() as Promise<Recipe>;
}
