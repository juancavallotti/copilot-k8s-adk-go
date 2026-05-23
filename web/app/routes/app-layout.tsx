import { Activity, BookOpen, ChefHat, CirclePlus } from "lucide-react";
import { useEffect } from "react";
import { NavLink, Outlet, useLoaderData } from "react-router";

import { AgentChat } from "~/components/agent-chat";
import { getAgentBaseURL } from "~/lib/agent-base-url";
import {
  AgentPreferencesProvider,
  useAgentPreferencesState,
} from "~/state/agent-preferences/context";
import {
  AgentPreferencesActionType,
  type ProvidersResponse,
} from "~/state/agent-preferences/types";

import type { Route } from "./+types/app-layout";

// clientLoader: runs in the browser, so the fetch reaches the agent the
// same way the chat does — bypasses the dev-container / SSR network gap.
export async function clientLoader(_args: Route.ClientLoaderArgs) {
  const baseURL = getAgentBaseURL();
  try {
    const res = await fetch(`${baseURL}/providers`);
    if (!res.ok) {
      throw new Error(`providers request failed (${res.status})`);
    }
    const options = (await res.json()) as ProvidersResponse;
    return { options, loadError: null as string | null };
  } catch (err) {
    return {
      options: null as ProvidersResponse | null,
      loadError:
        err instanceof Error ? err.message : "Could not load agent providers",
    };
  }
}


export function meta({}: Route.MetaArgs) {
  return [{ title: "Recipes" }];
}

function navClass(isActive: boolean) {
  return [
    "flex items-center gap-2.5 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors",
    isActive
      ? "bg-zinc-900 text-white shadow-sm dark:bg-zinc-100 dark:text-zinc-900"
      : "text-zinc-700 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800",
  ].join(" ");
}

function navIconClass(isActive: boolean) {
  return [
    "size-[1.125rem] shrink-0 stroke-[2]",
    isActive
      ? "text-white dark:text-zinc-900"
      : "text-zinc-500 dark:text-zinc-500 group-hover:text-zinc-700 dark:group-hover:text-zinc-300",
  ].join(" ");
}

const modelPrefsStorageKey = "recipes-agent-model-prefs";

function loadSavedPrefs():
  | { agentModel: string | null; imageModel: string | null }
  | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = window.localStorage.getItem(modelPrefsStorageKey);
    if (raw == null || raw === "") return null;
    const parsed = JSON.parse(raw) as Partial<{
      agentModel: string;
      imageModel: string;
    }>;
    return {
      agentModel: typeof parsed.agentModel === "string" ? parsed.agentModel : null,
      imageModel: typeof parsed.imageModel === "string" ? parsed.imageModel : null,
    };
  } catch {
    return null;
  }
}

export default function AppLayout() {
  return (
    <AgentPreferencesProvider>
      <AppLayoutContents />
    </AgentPreferencesProvider>
  );
}

function AppLayoutContents() {
  const loaderData = useLoaderData<typeof clientLoader>();
  const { dispatch } = useAgentPreferencesState();

  useEffect(() => {
    if (loaderData.loadError != null) {
      dispatch({
        type: AgentPreferencesActionType.OPTIONS_FAILED,
        data: loaderData.loadError,
      });
      return;
    }
    if (loaderData.options != null) {
      dispatch({
        type: AgentPreferencesActionType.OPTIONS_LOADED,
        data: { options: loaderData.options, saved: loadSavedPrefs() },
      });
    }
  }, [loaderData, dispatch]);

  return (
    <div className="flex h-screen bg-zinc-100 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
      <aside className="flex w-56 shrink-0 flex-col border-r border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-900">
        <div className="border-b border-zinc-200 px-4 py-4 dark:border-zinc-800">
          <p className="text-xs font-semibold uppercase tracking-wide text-zinc-500 dark:text-zinc-400">
            Tool
          </p>
          <div className="mt-1 flex items-center gap-2">
            <span className="flex size-8 items-center justify-center rounded-lg bg-amber-100 text-amber-800 dark:bg-amber-950/80 dark:text-amber-200">
              <ChefHat className="size-4 stroke-[2]" aria-hidden />
            </span>
            <p className="truncate text-sm font-semibold text-zinc-900 dark:text-zinc-50">
              Recipe manager
            </p>
          </div>
        </div>
        <nav className="flex flex-1 flex-col gap-1 p-3">
          <NavLink
            to="/create"
            className="group block rounded-lg outline-none focus-visible:ring-2 focus-visible:ring-zinc-400 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:focus-visible:ring-offset-zinc-900"
          >
            {({ isActive }) => (
              <span className={navClass(isActive)}>
                <CirclePlus className={navIconClass(isActive)} aria-hidden />
                Create recipe
              </span>
            )}
          </NavLink>
          <NavLink
            to="/"
            end
            className="group block rounded-lg outline-none focus-visible:ring-2 focus-visible:ring-zinc-400 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:focus-visible:ring-offset-zinc-900"
          >
            {({ isActive }) => (
              <span className={navClass(isActive)}>
                <BookOpen className={navIconClass(isActive)} aria-hidden />
                Recipes
              </span>
            )}
          </NavLink>
          <div className="mt-auto">
            <NavLink
              to="/traces"
              className="group block rounded-lg outline-none focus-visible:ring-2 focus-visible:ring-zinc-400 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:focus-visible:ring-offset-zinc-900"
            >
              {({ isActive }) => (
                <span className={navClass(isActive)}>
                  <Activity className={navIconClass(isActive)} aria-hidden />
                  Traces
                </span>
              )}
            </NavLink>
          </div>
        </nav>
      </aside>
      <div className="flex min-h-0 min-w-0 flex-1 flex-col">
        <header className="border-b border-zinc-200 bg-white px-6 py-4 dark:border-zinc-800 dark:bg-zinc-900">
          <h1 className="text-lg font-semibold tracking-tight text-zinc-900 dark:text-zinc-50">
            Workspace
          </h1>
        </header>
        <main className="flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>
      <AgentChat />
    </div>
  );
}
