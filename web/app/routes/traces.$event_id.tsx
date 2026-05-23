import { ArrowLeft } from "lucide-react";
import { useEffect } from "react";
import {
  Link,
  useLoaderData,
  useNavigation,
  useParams,
  useSearchParams,
} from "react-router";

import type { Trace } from "~/lib/traces-api";
import {
  TraceDetailProvider,
  useTraceDetailState,
} from "~/state/trace-detail/context";
import { TraceDetailActionType } from "~/state/trace-detail/types";

import type { Route } from "./+types/traces.$event_id";

export async function loader({ request, params }: Route.LoaderArgs) {
  const { listEventTraces } = await import("~/lib/traces-http.server");
  const eventId = params.event_id;
  if (eventId == null || eventId === "") {
    return {
      eventId: "",
      traces: null as Trace[] | null,
      error: "Missing event id.",
    };
  }
  try {
    const traces = await listEventTraces(request, eventId);
    return { eventId, traces, error: null as string | null };
  } catch (err) {
    return {
      eventId,
      traces: null as Trace[] | null,
      error: err instanceof Error ? err.message : "Something went wrong",
    };
  }
}

export function meta({ data }: Route.MetaArgs) {
  if (data?.eventId != null && data.eventId !== "") {
    return [{ title: `Event ${data.eventId} · Recipe manager` }];
  }
  return [{ title: "Event · Recipe manager" }];
}

function TraceDetailContent() {
  const params = useParams();
  const eventId = params.event_id ?? "";
  const loaderData = useLoaderData<typeof loader>();
  const { state, dispatch } = useTraceDetailState();
  const { traces, error } = state;
  const navigation = useNavigation();
  const [searchParams] = useSearchParams();
  const selectedId = searchParams.get("trace");

  useEffect(() => {
    if (eventId === "") {
      dispatch({
        type: TraceDetailActionType.MISSING_ID,
        data: "Missing event id.",
      });
      return;
    }
    dispatch({ type: TraceDetailActionType.LOAD_RESET });
    if (loaderData.error) {
      dispatch({
        type: TraceDetailActionType.LOAD_FAILED,
        data: loaderData.error,
      });
    } else if (
      loaderData.traces != null &&
      loaderData.eventId === eventId
    ) {
      dispatch({
        type: TraceDetailActionType.LOAD_SUCCESS,
        data: loaderData.traces,
      });
    }
  }, [eventId, loaderData, dispatch]);

  const isPending =
    eventId !== "" &&
    navigation.state === "loading" &&
    navigation.location?.pathname === `/traces/${eventId}` &&
    traces == null;

  const matchedTrace =
    traces != null && selectedId != null
      ? traces.find((t) => t.id === selectedId) ?? null
      : null;
  const selectionMissed =
    traces != null && selectedId != null && matchedTrace == null;
  const selectedTrace = matchedTrace ?? (traces?.[0] ?? null);

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Link
          to="/traces"
          className="inline-flex items-center gap-2 text-sm font-medium text-zinc-600 underline-offset-2 hover:text-zinc-900 hover:underline dark:text-zinc-400 dark:hover:text-zinc-100"
        >
          <ArrowLeft className="size-4 stroke-[2]" aria-hidden />
          All events
        </Link>
        {eventId !== "" ? (
          <p className="truncate font-mono text-xs text-zinc-500 dark:text-zinc-400">
            {eventId}
          </p>
        ) : null}
      </div>

      {error ? (
        <div
          className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-800 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-200"
          role="alert"
        >
          <p>{error}</p>
        </div>
      ) : null}

      {!error && (traces === null || isPending) ? (
        <p className="mt-8 text-sm text-zinc-500 dark:text-zinc-400">Loading…</p>
      ) : null}

      {!error && traces !== null && !isPending && traces.length === 0 ? (
        <div className="mt-8 rounded-xl border border-dashed border-zinc-300 bg-zinc-50/80 p-8 text-center dark:border-zinc-700 dark:bg-zinc-900/40">
          <p className="text-sm text-zinc-600 dark:text-zinc-400">
            This event has no traces.
          </p>
        </div>
      ) : null}

      {!error && traces !== null && !isPending && traces.length > 0 ? (
        <div className="mt-4 flex min-h-0 flex-1 gap-4">
          <ul className="flex w-72 shrink-0 flex-col divide-y divide-zinc-200 overflow-y-auto rounded-xl border border-zinc-200 bg-white dark:divide-zinc-800 dark:border-zinc-800 dark:bg-zinc-900">
            {traces.map((trace) => {
              const isSelected = selectedTrace?.id === trace.id;
              return (
                <li key={trace.id}>
                  <Link
                    replace
                    to={`/traces/${eventId}?trace=${encodeURIComponent(trace.id)}`}
                    className={
                      isSelected
                        ? "flex flex-col gap-0.5 bg-zinc-900 px-4 py-3 text-sm text-white dark:bg-zinc-100 dark:text-zinc-900"
                        : "flex flex-col gap-0.5 px-4 py-3 text-sm text-zinc-700 transition-colors hover:bg-zinc-50 dark:text-zinc-300 dark:hover:bg-zinc-800/50"
                    }
                  >
                    <span>{new Date(trace.occurred_at).toLocaleString()}</span>
                    <span className="font-mono text-xs opacity-70">
                      {trace.id.slice(0, 12)}
                    </span>
                  </Link>
                </li>
              );
            })}
          </ul>

          <div className="flex min-w-0 flex-1 flex-col overflow-hidden rounded-xl border border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-900">
            {selectedTrace != null ? (
              <>
                <div className="flex items-baseline justify-between gap-3 border-b border-zinc-200 px-4 py-2 dark:border-zinc-800">
                  <p className="truncate font-mono text-xs text-zinc-600 dark:text-zinc-400">
                    {selectedTrace.id}
                  </p>
                  <p className="shrink-0 text-xs text-zinc-500 dark:text-zinc-400">
                    {new Date(selectedTrace.occurred_at).toLocaleString()}
                  </p>
                </div>
                {selectionMissed ? (
                  <p className="border-b border-amber-200 bg-amber-50 px-4 py-2 text-xs text-amber-800 dark:border-amber-900/50 dark:bg-amber-950/40 dark:text-amber-200">
                    Selected trace not found in this event — showing the first
                    trace.
                  </p>
                ) : null}
                <pre className="m-0 flex-1 overflow-auto bg-zinc-50 p-4 font-mono text-xs leading-relaxed text-zinc-800 dark:bg-zinc-950/40 dark:text-zinc-200">
                  {JSON.stringify(selectedTrace.data, null, 2) ?? "(no data)"}
                </pre>
              </>
            ) : null}
          </div>
        </div>
      ) : null}
    </div>
  );
}

export default function TraceDetailRoute() {
  return (
    <TraceDetailProvider>
      <TraceDetailContent />
    </TraceDetailProvider>
  );
}
