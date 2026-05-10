import {
  isRouteErrorResponse,
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
  useLoaderData,
} from "react-router";

import { resolveRecipesApiBaseFromEnv } from "~/lib/recipe-api";

import type { Route } from "./+types/root";
import "./app.css";

export function loader(_args: Route.LoaderArgs) {
  return { recipesApiBase: resolveRecipesApiBaseFromEnv() };
}

/** Keep server-resolved API base; client reload has no `process.env.RECIPES_API_BASE`. */
export function shouldRevalidate() {
  return false;
}

export const links: Route.LinksFunction = () => [
  { rel: "preconnect", href: "https://fonts.googleapis.com" },
  {
    rel: "preconnect",
    href: "https://fonts.gstatic.com",
    crossOrigin: "anonymous",
  },
  {
    rel: "stylesheet",
    href: "https://fonts.googleapis.com/css2?family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap",
  },
];

export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <Meta />
        <Links />
      </head>
      <body>
        {children}
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  );
}

export default function App() {
  const { recipesApiBase } = useLoaderData<typeof loader>();
  const g = globalThis as typeof globalThis & { __RECIPES_API_BASE__?: string };
  g.__RECIPES_API_BASE__ = recipesApiBase;
  return <Outlet />;
}

export function ErrorBoundary({ error }: Route.ErrorBoundaryProps) {
  let message = "Oops!";
  let details = "An unexpected error occurred.";
  let stack: string | undefined;

  if (isRouteErrorResponse(error)) {
    message = error.status === 404 ? "404" : "Error";
    details =
      error.status === 404
        ? "The requested page could not be found."
        : error.statusText || details;
  } else if (import.meta.env.DEV && error && error instanceof Error) {
    details = error.message;
    stack = error.stack;
  }

  return (
    <main className="min-h-screen bg-zinc-100 p-8 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
      <div className="mx-auto max-w-lg rounded-xl border border-zinc-200 bg-white p-6 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
        <h1 className="text-lg font-semibold">{message}</h1>
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">{details}</p>
        {stack && (
          <pre className="mt-4 max-h-64 overflow-auto rounded-lg bg-zinc-50 p-3 text-xs dark:bg-zinc-950">
            <code>{stack}</code>
          </pre>
        )}
      </div>
    </main>
  );
}
