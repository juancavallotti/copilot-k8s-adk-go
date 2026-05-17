import { BookOpen, ChefHat, CirclePlus } from "lucide-react";
import { NavLink, Outlet } from "react-router";

import { AgentChat } from "~/components/agent-chat";

import type { Route } from "./+types/app-layout";

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

export default function AppLayout() {
  return (
    <div className="flex min-h-screen bg-zinc-100 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
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
        </nav>
      </aside>
      <div className="flex min-w-0 flex-1 flex-col">
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
