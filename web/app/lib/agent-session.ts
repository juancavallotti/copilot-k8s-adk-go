import type { ProvidersResponse } from "~/state/agent-preferences/types";

export const agentAppName = "recipe_copilot";
export const modelPrefsStorageKey = "recipes-agent-model-prefs";

// Session ID lives in memory only — reloading the page generates a fresh
// session so the agent does not carry over context from a prior visit.
let currentSessionID: string | null = null;

export function saveSavedPrefs(prefs: {
  agentModel: string | null;
  imageModel: string | null;
}) {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(modelPrefsStorageKey, JSON.stringify(prefs));
  } catch {
    // ignore quota / disabled storage
  }
}

export function buildModelContext(prefs: {
  options: ProvidersResponse | null;
  agentModel: string | null;
  imageModel: string | null;
}): { agentModel: string; imageModel: string } | null {
  if (
    prefs.options == null ||
    prefs.agentModel == null ||
    prefs.imageModel == null
  ) {
    return null;
  }
  if (
    prefs.agentModel === prefs.options.defaultAgentModel &&
    prefs.imageModel === prefs.options.defaultImageModel
  ) {
    return null;
  }
  return { agentModel: prefs.agentModel, imageModel: prefs.imageModel };
}

export function randomID(prefix: string): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return `${prefix}-${crypto.randomUUID()}`;
  }
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

export function getUserID(): string {
  const key = "recipes-agent-user-id";
  const existing = window.localStorage.getItem(key);
  if (existing != null && existing !== "") return existing;
  const next = randomID("web-user");
  window.localStorage.setItem(key, next);
  return next;
}

export function getSessionID(): string {
  if (currentSessionID != null && currentSessionID !== "") return currentSessionID;
  currentSessionID = randomID("session");
  return currentSessionID;
}

export function startNewSession(): string {
  currentSessionID = randomID("session");
  return currentSessionID;
}

export async function ensureSession(
  baseURL: string,
  userID: string,
  sessionID: string,
) {
  const sessionURL = `${baseURL}/apps/${encodeURIComponent(
    agentAppName,
  )}/users/${encodeURIComponent(userID)}/sessions/${encodeURIComponent(
    sessionID,
  )}`;

  const existing = await fetch(sessionURL);
  if (existing.ok) return;
  if (existing.status !== 404 && existing.status !== 500) {
    throw new Error(`Could not check agent session (${existing.status})`);
  }

  const res = await fetch(sessionURL, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: "{}",
  });
  if (!res.ok && res.status !== 409) {
    throw new Error(`Could not start agent session (${res.status})`);
  }
}
