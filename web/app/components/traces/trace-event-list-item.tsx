import { Trash2 } from "lucide-react";
import type { ElementType } from "react";
import { Link } from "react-router";

import { formatTraceDuration, formatTraceRange } from "~/lib/trace-format";
import type { Event } from "~/lib/traces-api";

export function TraceEventListItem({
  event,
  Form,
  disabled,
  isConfirming,
  isDeleting,
  onCancelDelete,
  onConfirmDelete,
  onDeleteStarted,
}: {
  event: Event;
  Form: ElementType;
  disabled: boolean;
  isConfirming: boolean;
  isDeleting: boolean;
  onCancelDelete: () => void;
  onConfirmDelete: () => void;
  onDeleteStarted: () => void;
}) {
  return (
    <li className="flex items-stretch transition-colors hover:bg-zinc-50 dark:hover:bg-zinc-800/50">
      <Link
        to={`/traces/${event.event_id}`}
        className="flex min-w-0 flex-1 items-center justify-between gap-4 px-4 py-3"
      >
        <div className="min-w-0">
          {event.user_prompt ? (
            <>
              <p className="line-clamp-2 text-sm font-medium text-zinc-900 dark:text-zinc-100">
                {event.user_prompt}
              </p>
              <p className="mt-0.5 truncate font-mono text-xs text-zinc-500 dark:text-zinc-400">
                {event.event_id}
              </p>
            </>
          ) : (
            <p className="truncate font-mono text-sm text-zinc-900 dark:text-zinc-100">
              {event.event_id}
            </p>
          )}
          <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
            {formatTraceRange(event.started_at, event.ended_at)}
          </p>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <span className="rounded-full bg-zinc-100 px-2 py-0.5 text-xs font-medium text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300">
            {formatTraceDuration(event.started_at, event.ended_at)}
          </span>
          <span className="rounded-full bg-zinc-100 px-2 py-0.5 text-xs font-medium text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300">
            {event.trace_count} traces
          </span>
        </div>
      </Link>
      <div className="flex shrink-0 items-stretch border-l border-zinc-100 dark:border-zinc-800">
        {isConfirming ? (
          <Form
            method="post"
            className="flex items-center gap-2 px-3 py-2"
            onSubmit={onDeleteStarted}
          >
            <input type="hidden" name="intent" value="delete-event" />
            <input type="hidden" name="event_id" value={event.event_id} />
            <span className="text-xs font-medium text-zinc-700 dark:text-zinc-300">
              Delete?
            </span>
            <button
              type="button"
              className="rounded-md px-2 py-1 text-xs font-medium text-zinc-600 transition-colors hover:bg-zinc-100 hover:text-zinc-900 disabled:pointer-events-none disabled:opacity-40 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
              onClick={onCancelDelete}
              disabled={disabled}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="rounded-md bg-red-600 px-2 py-1 text-xs font-medium text-white transition-colors hover:bg-red-700 disabled:pointer-events-none disabled:opacity-40 dark:bg-red-700 dark:hover:bg-red-600"
              disabled={disabled}
            >
              Confirm
            </button>
          </Form>
        ) : (
          <button
            type="button"
            className="flex items-center justify-center px-3 text-zinc-400 transition-colors hover:bg-red-50 hover:text-red-700 disabled:pointer-events-none disabled:opacity-40 dark:hover:bg-red-950/40 dark:hover:text-red-300"
            aria-label={`Delete event ${event.event_id}`}
            disabled={disabled}
            onClick={onConfirmDelete}
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
}
