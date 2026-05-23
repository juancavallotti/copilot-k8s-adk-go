import { useEffect } from "react";
import { Link, useLoaderData, useNavigation, useRevalidator } from "react-router";

import type { Event } from "~/lib/traces-api";
import {
  TracesListProvider,
  useTracesListState,
} from "~/state/traces-list/context";
import { TracesListActionType } from "~/state/traces-list/types";

import type { Route } from "./+types/traces._index";

export async function loader({ request }: Route.LoaderArgs) {
  const { listEvents } = await import("~/lib/traces-http.server");
  try {
    const events = await listEvents(request);
    return { events, listError: null as string | null };
  } catch (err) {
    return {
      events: null as Event[] | null,
      listError:
        err instanceof Error ? err.message : "Something went wrong",
    };
  }
}

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Traces · Recipe manager" },
    { name: "description", content: "Recent agent events" },
  ];
}

function formatRange(startedAt: string, endedAt: string): string {
  const start = new Date(startedAt);
  const end = new Date(endedAt);
  return `${start.toLocaleString()} → ${end.toLocaleString()}`;
}

function formatDuration(startedAt: string, endedAt: string): string {
  const ms = new Date(endedAt).getTime() - new Date(startedAt).getTime();
  if (ms < 0) return "";
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60_000) return `${(ms / 1000).toFixed(2)}s`;
  const m = Math.floor(ms / 60_000);
  const s = Math.round((ms % 60_000) / 1000);
  return `${m}m ${s}s`;
}

function TracesIndexContent() {
  const loaderData = useLoaderData<typeof loader>();
  const { state, dispatch } = useTracesListState();
  const { events, listError } = state;
  const navigation = useNavigation();
  const revalidator = useRevalidator();

  const isLoadingList =
    navigation.state === "loading" &&
    navigation.location?.pathname === "/traces" &&
    navigation.formMethod == null;

  useEffect(() => {
    if (loaderData.listError != null) {
      dispatch({
        type: TracesListActionType.FETCH_FAILED,
        data: loaderData.listError,
      });
    } else if (loaderData.events != null) {
      dispatch({
        type: TracesListActionType.FETCH_SUCCESS,
        data: loaderData.events,
      });
    }
  }, [loaderData, dispatch]);

  function retryList() {
    dispatch({ type: TracesListActionType.FETCH_STARTED });
    void revalidator.revalidate();
  }

  return (
    <div className="mx-auto max-w-3xl">
      <h2 className="text-base font-medium text-zinc-900 dark:text-zinc-50">
        Traces
      </h2>
      <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
        Recent agent events, newest first.
      </p>

      {listError ? (
        <div
          className="mt-8 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-800 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-200"
          role="alert"
        >
          <p>{listError}</p>
          <button
            type="button"
            className="mt-3 text-sm font-medium text-red-900 underline-offset-2 hover:underline dark:text-red-100"
            onClick={retryList}
          >
            Try again
          </button>
        </div>
      ) : null}

      {!listError && (events === null || isLoadingList) ? (
        <p className="mt-8 text-sm text-zinc-500 dark:text-zinc-400">Loading…</p>
      ) : null}

      {!listError && events !== null && !isLoadingList && events.length === 0 ? (
        <div className="mt-8 rounded-xl border border-dashed border-zinc-300 bg-zinc-50/80 p-8 text-center dark:border-zinc-700 dark:bg-zinc-900/40">
          <p className="text-sm text-zinc-600 dark:text-zinc-400">
            No events yet.
          </p>
        </div>
      ) : null}

      {!listError && events !== null && !isLoadingList && events.length > 0 ? (
        <ul className="mt-6 divide-y divide-zinc-200 overflow-hidden rounded-xl border border-zinc-200 bg-white dark:divide-zinc-800 dark:border-zinc-800 dark:bg-zinc-900">
          {events.map((evt) => (
            <li key={evt.event_id}>
              <Link
                to={`/traces/${evt.event_id}`}
                className="flex items-center justify-between gap-4 px-4 py-3 transition-colors hover:bg-zinc-50 dark:hover:bg-zinc-800/50"
              >
                <div className="min-w-0">
                  <p className="truncate font-mono text-sm text-zinc-900 dark:text-zinc-100">
                    {evt.event_id}
                  </p>
                  <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
                    {formatRange(evt.started_at, evt.ended_at)}
                  </p>
                </div>
                <div className="flex shrink-0 items-center gap-2">
                  <span className="rounded-full bg-zinc-100 px-2 py-0.5 text-xs font-medium text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300">
                    {formatDuration(evt.started_at, evt.ended_at)}
                  </span>
                  <span className="rounded-full bg-zinc-100 px-2 py-0.5 text-xs font-medium text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300">
                    {evt.trace_count} traces
                  </span>
                </div>
              </Link>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}

export default function TracesIndex() {
  return (
    <TracesListProvider>
      <TracesIndexContent />
    </TracesListProvider>
  );
}
