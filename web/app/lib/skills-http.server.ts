import type { Skill, SkillCreate, SkillPatch } from "~/lib/skills-api";

if (typeof window !== "undefined") {
  throw new Error(
    "skills-http.server.ts is server-only; it must not run in the browser.",
  );
}

function normalizeApiBase(raw: string): string {
  return raw.replace(/\/$/, "");
}

/**
 * Base URL for skills HTTP calls on the **Node server** (loaders/actions).
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

export async function listSkills(request: Request): Promise<Skill[]> {
  const base = getApiBase(request);
  const res = await fetch(`${base}/skills`);
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res.json() as Promise<Skill[]>;
}

export async function getSkill(request: Request, id: string): Promise<Skill> {
  const base = getApiBase(request);
  const res = await fetch(`${base}/skills/${encodeURIComponent(id)}`);
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res.json() as Promise<Skill>;
}

export async function createSkill(
  request: Request,
  body: SkillCreate,
): Promise<Skill> {
  const base = getApiBase(request);
  const res = await fetch(`${base}/skills`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res.json() as Promise<Skill>;
}

export async function updateSkill(
  request: Request,
  id: string,
  body: SkillPatch,
): Promise<Skill> {
  const base = getApiBase(request);
  const res = await fetch(`${base}/skills/${encodeURIComponent(id)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res.json() as Promise<Skill>;
}

export async function deleteSkill(
  request: Request,
  id: string,
): Promise<void> {
  const base = getApiBase(request);
  const res = await fetch(`${base}/skills/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (!res.ok) {
    throw await readJsonError(res);
  }
}
