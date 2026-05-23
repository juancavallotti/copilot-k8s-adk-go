import { useEffect } from "react";
import { Wrench, X } from "lucide-react";

import type { ToolCall } from "~/lib/agent-ui-actions";

import { HighlightedJSON } from "./highlighted-json";

function formatJSON(value: unknown): string {
  if (value === undefined) return "";
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return String(value);
  }
}

export function ToolCallDialog({
  call,
  onClose,
}: {
  call: ToolCall;
  onClose: () => void;
}) {
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") onClose();
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  const argsBody = formatJSON(call.args);
  const responseBody = formatJSON(call.response);
  const statusLabel =
    call.status === "pending"
      ? "Running"
      : call.status === "error"
        ? "Failed"
        : "Succeeded";
  const statusStyles =
    call.status === "error"
      ? "bg-red-100 text-red-800 dark:bg-red-950/70 dark:text-red-200"
      : call.status === "pending"
        ? "bg-zinc-200 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200"
        : "bg-emerald-100 text-emerald-800 dark:bg-emerald-950/70 dark:text-emerald-200";

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={`Tool call: ${call.name}`}
      className="fixed inset-0 z-[60] flex items-center justify-center bg-zinc-950/85 p-4"
      onClick={onClose}
    >
      <div
        className="flex max-h-[85vh] w-full max-w-3xl flex-col overflow-hidden rounded-2xl border border-zinc-800 bg-zinc-950 text-zinc-100 shadow-2xl"
        onClick={(event) => event.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-3 border-b border-zinc-800 px-4 py-3">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <Wrench className="size-4 text-zinc-400" aria-hidden />
              <h3 className="truncate font-mono text-sm font-semibold">
                {call.name}
              </h3>
              <span
                className={`shrink-0 rounded-full px-2 py-0.5 text-[0.6875rem] font-medium ${statusStyles}`}
              >
                {statusLabel}
              </span>
            </div>
            {call.summary ? (
              <p className="mt-1 break-all font-mono text-xs text-zinc-400">
                {call.summary}
              </p>
            ) : null}
          </div>
          <button
            type="button"
            className="rounded-full p-2 text-zinc-400 transition-colors hover:bg-zinc-800 hover:text-zinc-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-zinc-500"
            onClick={onClose}
            aria-label="Close tool call dialog"
          >
            <X className="size-4" aria-hidden />
          </button>
        </header>
        <div className="flex-1 space-y-4 overflow-auto px-4 py-3">
          <section>
            <h4 className="mb-1.5 text-[0.6875rem] font-semibold uppercase tracking-wide text-zinc-500">
              Arguments
            </h4>
            <pre className="whitespace-pre-wrap break-words rounded-lg bg-zinc-900 px-3 py-2 font-mono text-xs leading-relaxed text-zinc-300">
              {argsBody === "" ? (
                <span className="text-zinc-500">(no arguments)</span>
              ) : (
                <HighlightedJSON source={argsBody} />
              )}
            </pre>
          </section>
          <section>
            <h4 className="mb-1.5 text-[0.6875rem] font-semibold uppercase tracking-wide text-zinc-500">
              Response
            </h4>
            <pre className="whitespace-pre-wrap break-words rounded-lg bg-zinc-900 px-3 py-2 font-mono text-xs leading-relaxed text-zinc-300">
              {responseBody === "" ? (
                <span className="text-zinc-500">
                  {call.status === "pending"
                    ? "(awaiting response)"
                    : "(no response data)"}
                </span>
              ) : (
                <HighlightedJSON source={responseBody} />
              )}
            </pre>
          </section>
        </div>
      </div>
    </div>
  );
}
