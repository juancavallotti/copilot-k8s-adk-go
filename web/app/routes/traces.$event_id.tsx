import { ArrowLeft, Bot, Wrench } from "lucide-react";
import { useEffect, type ReactNode } from "react";
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

function getTraceMsg(data: unknown): string {
  if (data != null && typeof data === "object" && "msg" in data) {
    const m = (data as { msg?: unknown }).msg;
    if (typeof m === "string") return m;
  }
  return "";
}

function getToolName(data: unknown): string {
  if (data != null && typeof data === "object" && "tool" in data) {
    const t = (data as { tool?: unknown }).tool;
    if (typeof t === "string") return t;
  }
  return "";
}

type TraceItem =
  | { kind: "single"; trace: Trace }
  | { kind: "toolGroup"; traces: Trace[] };

function groupTraces(traces: Trace[]): TraceItem[] {
  const items: TraceItem[] = [];
  let pending: Trace[] = [];
  const flush = () => {
    if (pending.length > 0) {
      items.push({ kind: "toolGroup", traces: pending });
      pending = [];
    }
  };
  for (const trace of traces) {
    if (getTraceMsg(trace.data).startsWith("tool.")) {
      pending.push(trace);
    } else {
      flush();
      items.push({ kind: "single", trace });
    }
  }
  flush();
  return items;
}

function findContainingItem(
  items: TraceItem[],
  traceId: string | null,
): TraceItem | null {
  if (traceId == null) return null;
  for (const item of items) {
    if (item.kind === "single") {
      if (item.trace.id === traceId) return item;
    } else if (item.traces.some((t) => t.id === traceId)) {
      return item;
    }
  }
  return null;
}

function itemKey(item: TraceItem): string {
  return item.kind === "single" ? item.trace.id : item.traces[0].id;
}

function itemTime(item: TraceItem): string {
  return item.kind === "single"
    ? item.trace.occurred_at
    : item.traces[0].occurred_at;
}

function durationMs(from: string, to: string): number {
  return new Date(to).getTime() - new Date(from).getTime();
}

function formatDuration(ms: number): string {
  if (ms < 0) return "";
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60_000) return `${(ms / 1000).toFixed(2)}s`;
  const m = Math.floor(ms / 60_000);
  const s = Math.round((ms % 60_000) / 1000);
  return `${m}m ${s}s`;
}

function itemToolName(item: Extract<TraceItem, { kind: "toolGroup" }>): string {
  for (const t of item.traces) {
    const name = getToolName(t.data);
    if (name !== "") return name;
  }
  return "tool";
}

const jsonTokenRe =
  /("(?:\\.|[^"\\])*")\s*:|"(?:\\.|[^"\\])*"|\btrue\b|\bfalse\b|\bnull\b|-?\d+\.?\d*(?:[eE][+-]?\d+)?/g;

function highlightJson(source: string): ReactNode[] {
  const out: ReactNode[] = [];
  let last = 0;
  let key = 0;
  let m: RegExpExecArray | null;
  jsonTokenRe.lastIndex = 0;
  while ((m = jsonTokenRe.exec(source)) !== null) {
    if (m.index > last) {
      out.push(source.slice(last, m.index));
    }
    const matched = m[0];
    if (m[1] != null) {
      out.push(
        <span key={key++} className="text-sky-700 dark:text-sky-300">
          {m[1]}
        </span>,
      );
      out.push(matched.slice(m[1].length));
    } else if (matched.startsWith('"')) {
      out.push(
        <span key={key++} className="text-emerald-700 dark:text-emerald-300">
          {matched}
        </span>,
      );
    } else if (matched === "true" || matched === "false") {
      out.push(
        <span key={key++} className="text-purple-700 dark:text-purple-300">
          {matched}
        </span>,
      );
    } else if (matched === "null") {
      out.push(
        <span key={key++} className="text-zinc-500">
          {matched}
        </span>,
      );
    } else {
      out.push(
        <span key={key++} className="text-amber-700 dark:text-amber-400">
          {matched}
        </span>,
      );
    }
    last = m.index + matched.length;
  }
  if (last < source.length) out.push(source.slice(last));
  return out;
}

function renderTraceJson(data: unknown): ReactNode[] {
  return highlightJson(JSON.stringify(data, null, 2) ?? "(no data)");
}

function TraceRow({
  trace,
  eventId,
  isSelected,
  label,
  icon,
  duration,
}: {
  trace: Trace;
  eventId: string;
  isSelected: boolean;
  label: string;
  icon: ReactNode;
  duration: number | null;
}) {
  return (
    <Link
      replace
      to={`/traces/${eventId}?trace=${encodeURIComponent(trace.id)}`}
      className={
        isSelected
          ? "flex items-start gap-2 rounded-md bg-zinc-900 px-3 py-2 text-sm text-white dark:bg-zinc-100 dark:text-zinc-900"
          : "flex items-start gap-2 rounded-md px-3 py-2 text-sm text-zinc-700 transition-colors hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800/50"
      }
    >
      <span className="mt-0.5 flex size-3.5 shrink-0 items-center justify-center">
        {icon}
      </span>
      <span className="flex min-w-0 flex-col gap-0.5">
        <span className="truncate font-medium">{label}</span>
        <span className="text-xs opacity-70">
          {new Date(trace.occurred_at).toLocaleTimeString()}
          {duration != null ? ` · ${formatDuration(duration)}` : ""}
        </span>
      </span>
    </Link>
  );
}

function SelectedItemView({
  item,
  selectionMissed,
}: {
  item: TraceItem;
  selectionMissed: boolean;
}) {
  const missedHint = selectionMissed ? (
    <p className="border-b border-amber-200 bg-amber-50 px-4 py-2 text-xs text-amber-800 dark:border-amber-900/50 dark:bg-amber-950/40 dark:text-amber-200">
      Selected trace not found in this event — showing the first trace.
    </p>
  ) : null;

  if (item.kind === "single") {
    return (
      <>
        <div className="flex items-baseline justify-between gap-3 border-b border-zinc-200 px-4 py-2 dark:border-zinc-800">
          <p className="truncate font-mono text-xs text-zinc-600 dark:text-zinc-400">
            {item.trace.id}
          </p>
          <p className="shrink-0 text-xs text-zinc-500 dark:text-zinc-400">
            {new Date(item.trace.occurred_at).toLocaleString()}
          </p>
        </div>
        {missedHint}
        <pre className="m-0 flex-1 overflow-auto whitespace-pre-wrap break-words bg-zinc-50 p-4 font-mono text-xs leading-relaxed text-zinc-800 dark:bg-zinc-950/40 dark:text-zinc-200">
          {renderTraceJson(item.trace.data)}
        </pre>
      </>
    );
  }

  const toolName = itemToolName(item);
  const first = item.traces[0];
  const last = item.traces[item.traces.length - 1];
  const totalDuration =
    item.traces.length > 1
      ? durationMs(first.occurred_at, last.occurred_at)
      : null;
  return (
    <>
      <div className="flex items-baseline justify-between gap-3 border-b border-zinc-200 px-4 py-2 dark:border-zinc-800">
        <p className="truncate text-xs text-zinc-600 dark:text-zinc-400">
          Tool call · <span className="font-mono">{toolName}</span>
          {totalDuration != null ? (
            <span className="ml-2 text-zinc-500 dark:text-zinc-500">
              · {formatDuration(totalDuration)} total
            </span>
          ) : null}
        </p>
        <p className="shrink-0 text-xs text-zinc-500 dark:text-zinc-400">
          {new Date(first.occurred_at).toLocaleString()}
        </p>
      </div>
      {missedHint}
      <div className="flex-1 overflow-auto bg-zinc-50 dark:bg-zinc-950/40">
        {item.traces.map((trace, i) => {
          const msg = getTraceMsg(trace.data);
          const short = msg.startsWith("tool.")
            ? msg.slice("tool.".length)
            : msg !== ""
              ? msg
              : "event";
          const nextTrace = item.traces[i + 1];
          const sectionDuration = nextTrace
            ? durationMs(trace.occurred_at, nextTrace.occurred_at)
            : null;
          return (
            <div key={trace.id}>
              <div className="flex items-baseline justify-between gap-3 border-y border-zinc-200 bg-white px-4 py-1.5 text-xs font-medium uppercase tracking-wide text-zinc-500 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-400">
                <span>{short}</span>
                <span className="font-normal normal-case tracking-normal">
                  {new Date(trace.occurred_at).toLocaleTimeString()}
                  {sectionDuration != null
                    ? ` · ${formatDuration(sectionDuration)}`
                    : ""}
                </span>
              </div>
              <pre className="m-0 overflow-auto whitespace-pre-wrap break-words p-4 font-mono text-xs leading-relaxed text-zinc-800 dark:text-zinc-200">
                {renderTraceJson(trace.data)}
              </pre>
            </div>
          );
        })}
      </div>
    </>
  );
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

  const items = traces != null ? groupTraces(traces) : [];
  const matchedItem = findContainingItem(items, selectedId);
  const selectionMissed =
    traces != null && selectedId != null && matchedItem == null;
  const selectedItem = matchedItem ?? (items[0] ?? null);

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
          <div className="flex w-72 shrink-0 flex-col gap-1 overflow-y-auto rounded-xl border border-zinc-200 bg-white p-2 dark:border-zinc-800 dark:bg-zinc-900">
            {items.map((item, i) => {
              const next = items[i + 1];
              if (item.kind === "single") {
                const msg = getTraceMsg(item.trace.data);
                const duration = next
                  ? durationMs(item.trace.occurred_at, itemTime(next))
                  : null;
                return (
                  <TraceRow
                    key={item.trace.id}
                    trace={item.trace}
                    eventId={eventId}
                    isSelected={itemKey(selectedItem!) === item.trace.id}
                    label={msg !== "" ? msg : "trace"}
                    icon={
                      msg === "agent.event" ? (
                        <Bot className="size-3.5 shrink-0" aria-hidden />
                      ) : null
                    }
                    duration={duration}
                  />
                );
              }
              const first = item.traces[0];
              const last = item.traces[item.traces.length - 1];
              const totalDuration =
                item.traces.length > 1
                  ? durationMs(first.occurred_at, last.occurred_at)
                  : null;
              return (
                <TraceRow
                  key={first.id}
                  trace={first}
                  eventId={eventId}
                  isSelected={itemKey(selectedItem!) === first.id}
                  label={itemToolName(item)}
                  icon={<Wrench className="size-3.5 shrink-0" aria-hidden />}
                  duration={totalDuration}
                />
              );
            })}
          </div>

          <div className="flex min-w-0 flex-1 flex-col overflow-hidden rounded-xl border border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-900">
            {selectedItem != null ? (
              <SelectedItemView
                item={selectedItem}
                selectionMissed={selectionMissed}
              />
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
