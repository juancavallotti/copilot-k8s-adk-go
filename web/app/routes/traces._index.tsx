import { Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import {
  Link,
  useFetcher,
  useLoaderData,
  useNavigation,
  useRevalidator,
} from "react-router";

import type { Event } from "~/lib/traces-api";
import {
  TracesListProvider,
  useTracesListState,
} from "~/state/traces-list/context";
import { TracesListActionType } from "~/state/traces-list/types";

import type { Route } from "./+types/traces._index";

type TracesActionResult =
  | { ok: true; intent: "clear" }
  | { ok: true; intent: "delete-event"; eventId: string }
  | { ok: false; intent: "clear" | "delete-event"; eventId?: string; error: string };

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

export async function action({ request }: Route.ActionArgs) {
  const formData = await request.formData();
  const intent = formData.get("intent");

  if (intent === "clear") {
    const { deleteAllEvents } = await import("~/lib/traces-http.server");
    try {
      await deleteAllEvents(request);
      return { ok: true as const, intent: "clear" as const };
    } catch (err) {
      return {
        ok: false as const,
        intent: "clear" as const,
        error: err instanceof Error ? err.message : "Could not clear events",
      };
    }
  }

  if (intent === "delete-event") {
    const { deleteEvent } = await import("~/lib/traces-http.server");
    const eventId = formData.get("event_id");
    if (typeof eventId !== "string" || eventId === "") {
      return {
        ok: false as const,
        intent: "delete-event" as const,
        error: "Missing event id.",
      };
    }
    try {
      await deleteEvent(request, eventId);
      return {
        ok: true as const,
        intent: "delete-event" as const,
        eventId,
      };
    } catch (err) {
      return {
        ok: false as const,
        intent: "delete-event" as const,
        eventId,
        error: err instanceof Error ? err.message : "Could not delete event",
      };
    }
  }

  return {
    ok: false as const,
    intent: "clear" as const,
    error: "Unsupported action.",
  };
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
  const { events, listError, deletingEventId, isClearing, mutationError } =
    state;
  const navigation = useNavigation();
  const revalidator = useRevalidator();
  const clearFetcher = useFetcher<TracesActionResult>();
  const deleteFetcher = useFetcher<TracesActionResult>();
  const [isConfirmingClear, setIsConfirmingClear] = useState(false);
  const [confirmingEventId, setConfirmingEventId] = useState<string | null>(
    null,
  );

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

  useEffect(() => {
    if (clearFetcher.state !== "idle" || clearFetcher.data == null) return;
    const data = clearFetcher.data;
    if (data.intent !== "clear") return;
    if (data.ok) {
      dispatch({ type: TracesListActionType.CLEAR_ALL_SUCCEEDED });
    } else {
      dispatch({
        type: TracesListActionType.CLEAR_ALL_FAILED,
        data: data.error,
      });
    }
  }, [clearFetcher.state, clearFetcher.data, dispatch]);

  useEffect(() => {
    if (deleteFetcher.state !== "idle" || deleteFetcher.data == null) return;
    const data = deleteFetcher.data;
    if (data.intent !== "delete-event") return;
    if (data.ok) {
      dispatch({
        type: TracesListActionType.DELETE_EVENT_SUCCEEDED,
        data: data.eventId,
      });
    } else {
      dispatch({
        type: TracesListActionType.DELETE_EVENT_FAILED,
        data: data.error,
      });
    }
  }, [deleteFetcher.state, deleteFetcher.data, dispatch]);

  function retryList() {
    dispatch({ type: TracesListActionType.FETCH_STARTED });
    void revalidator.revalidate();
  }

  function dismissMutationError() {
    dispatch({ type: TracesListActionType.MUTATION_DISMISS });
  }

  const hasEvents = events != null && events.length > 0;
  const anyMutation = isClearing || deletingEventId !== null;

  return (
    <div className="mx-auto max-w-3xl">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-base font-medium text-zinc-900 dark:text-zinc-50">
            Traces
          </h2>
          <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
            Recent agent events, newest first.
          </p>
        </div>
        {hasEvents ? (
          <div className="shrink-0">
            {isConfirmingClear ? (
              <clearFetcher.Form
                method="post"
                className="flex items-center gap-2"
                onSubmit={() => {
                  setIsConfirmingClear(false);
                  dispatch({
                    type: TracesListActionType.CLEAR_ALL_STARTED,
                  });
                }}
              >
                <input type="hidden" name="intent" value="clear" />
                <span className="text-xs font-medium text-zinc-700 dark:text-zinc-300">
                  Delete all events?
                </span>
                <button
                  type="button"
                  className="rounded-md px-2 py-1 text-xs font-medium text-zinc-600 transition-colors hover:bg-zinc-100 hover:text-zinc-900 disabled:pointer-events-none disabled:opacity-40 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
                  onClick={() => setIsConfirmingClear(false)}
                  disabled={anyMutation}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="rounded-md bg-red-600 px-2 py-1 text-xs font-medium text-white transition-colors hover:bg-red-700 disabled:pointer-events-none disabled:opacity-40 dark:bg-red-700 dark:hover:bg-red-600"
                  disabled={anyMutation}
                >
                  Clear all
                </button>
              </clearFetcher.Form>
            ) : (
              <button
                type="button"
                onClick={() => setIsConfirmingClear(true)}
                disabled={anyMutation}
                className="inline-flex items-center gap-1.5 rounded-lg border border-zinc-200 bg-white px-3 py-1.5 text-sm font-medium text-zinc-700 transition-colors hover:bg-red-50 hover:text-red-700 disabled:pointer-events-none disabled:opacity-40 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:bg-red-950/40 dark:hover:text-red-300"
              >
                <Trash2 className="size-4 stroke-[2]" aria-hidden />
                Clear all
              </button>
            )}
          </div>
        ) : null}
      </div>

      {mutationError ? (
        <div
          className="mt-4 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-200"
          role="alert"
        >
          <p>{mutationError}</p>
          <button
            type="button"
            className="mt-2 text-sm font-medium text-red-900 underline-offset-2 hover:underline dark:text-red-100"
            onClick={dismissMutationError}
          >
            Dismiss
          </button>
        </div>
      ) : null}

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
          {events.map((evt) => {
            const isConfirming = confirmingEventId === evt.event_id;
            const isDeleting = deletingEventId === evt.event_id;
            return (
              <li
                key={evt.event_id}
                className="flex items-stretch transition-colors hover:bg-zinc-50 dark:hover:bg-zinc-800/50"
              >
                <Link
                  to={`/traces/${evt.event_id}`}
                  className="flex min-w-0 flex-1 items-center justify-between gap-4 px-4 py-3"
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
                <div className="flex shrink-0 items-stretch border-l border-zinc-100 dark:border-zinc-800">
                  {isConfirming ? (
                    <deleteFetcher.Form
                      method="post"
                      className="flex items-center gap-2 px-3 py-2"
                      onSubmit={() => {
                        setConfirmingEventId(null);
                        dispatch({
                          type: TracesListActionType.DELETE_EVENT_STARTED,
                          data: evt.event_id,
                        });
                      }}
                    >
                      <input
                        type="hidden"
                        name="intent"
                        value="delete-event"
                      />
                      <input
                        type="hidden"
                        name="event_id"
                        value={evt.event_id}
                      />
                      <span className="text-xs font-medium text-zinc-700 dark:text-zinc-300">
                        Delete?
                      </span>
                      <button
                        type="button"
                        className="rounded-md px-2 py-1 text-xs font-medium text-zinc-600 transition-colors hover:bg-zinc-100 hover:text-zinc-900 disabled:pointer-events-none disabled:opacity-40 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
                        onClick={() => setConfirmingEventId(null)}
                        disabled={anyMutation}
                      >
                        Cancel
                      </button>
                      <button
                        type="submit"
                        className="rounded-md bg-red-600 px-2 py-1 text-xs font-medium text-white transition-colors hover:bg-red-700 disabled:pointer-events-none disabled:opacity-40 dark:bg-red-700 dark:hover:bg-red-600"
                        disabled={anyMutation}
                      >
                        Confirm
                      </button>
                    </deleteFetcher.Form>
                  ) : (
                    <button
                      type="button"
                      className="flex items-center justify-center px-3 text-zinc-400 transition-colors hover:bg-red-50 hover:text-red-700 disabled:pointer-events-none disabled:opacity-40 dark:hover:bg-red-950/40 dark:hover:text-red-300"
                      aria-label={`Delete event ${evt.event_id}`}
                      disabled={anyMutation}
                      onClick={() => setConfirmingEventId(evt.event_id)}
                    >
                      {isDeleting ? (
                        <span className="text-xs font-medium text-zinc-500 dark:text-zinc-400">
                          ...
                        </span>
                      ) : (
                        <Trash2 className="size-4 stroke-[2]" aria-hidden />
                      )}
                    </button>
                  )}
                </div>
              </li>
            );
          })}
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
