import type { Event, Trace } from "~/lib/traces-api";

if (typeof window !== "undefined") {
  throw new Error(
    "traces-http.server.ts is server-only; it must not run in the browser.",
  );
}

function normalizeApiBase(raw: string): string {
  return raw.replace(/\/$/, "");
}

/**
 * Base URL for traces HTTP calls on the **Node server** (loaders/actions).
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

export async function listEvents(
  request: Request,
  opts?: { limit?: number; offset?: number },
): Promise<Event[]> {
  const base = getApiBase(request);
  const params = new URLSearchParams();
  if (opts?.limit != null) params.set("limit", String(opts.limit));
  if (opts?.offset != null) params.set("offset", String(opts.offset));
  const qs = params.toString();
  const res = await fetch(`${base}/events${qs === "" ? "" : `?${qs}`}`);
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res.json() as Promise<Event[]>;
}

export async function listEventTraces(
  request: Request,
  eventId: string,
): Promise<Trace[]> {
  const base = getApiBase(request);
  const res = await fetch(
    `${base}/events/${encodeURIComponent(eventId)}/traces`,
  );
  if (!res.ok) {
    throw await readJsonError(res);
  }
  return res.json() as Promise<Trace[]>;
}

export async function deleteAllEvents(request: Request): Promise<void> {
  const base = getApiBase(request);
  const res = await fetch(`${base}/events`, { method: "DELETE" });
  if (!res.ok) {
    throw await readJsonError(res);
  }
}

export async function deleteEvent(
  request: Request,
  eventId: string,
): Promise<void> {
  const base = getApiBase(request);
  const res = await fetch(`${base}/events/${encodeURIComponent(eventId)}`, {
    method: "DELETE",
  });
  if (!res.ok) {
    throw await readJsonError(res);
  }
}
