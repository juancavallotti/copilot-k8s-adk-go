import { AlertCircle, CheckCircle2, Loader2, Wrench } from "lucide-react";

import type { ToolCall } from "~/lib/agent-ui-actions";
import { useToolCall } from "~/lib/tool-context";

function describeToolCall(call: ToolCall): string {
  if (call.summary != null && call.summary !== "") return call.summary;
  return call.name;
}

export function ToolCallWidget({
  id,
  onSelect,
}: {
  id: string;
  onSelect: (id: string) => void;
}) {
  const call = useToolCall(id);
  if (call == null) return null;

  const label = describeToolCall(call);
  const styles =
    call.status === "error"
      ? "bg-red-50 text-red-800 hover:bg-red-100 dark:bg-red-950/60 dark:text-red-200 dark:hover:bg-red-900/60"
      : call.status === "pending"
        ? "bg-zinc-100 text-zinc-700 hover:bg-zinc-200 dark:bg-zinc-800 dark:text-zinc-200 dark:hover:bg-zinc-700"
        : "bg-emerald-50 text-emerald-800 hover:bg-emerald-100 dark:bg-emerald-950/60 dark:text-emerald-200 dark:hover:bg-emerald-900/60";
  const icon =
    call.status === "error" ? (
      <AlertCircle className="size-3" aria-hidden />
    ) : call.status === "pending" ? (
      <Loader2 className="size-3 animate-spin" aria-hidden />
    ) : (
      <CheckCircle2 className="size-3" aria-hidden />
    );

  return (
    <button
      type="button"
      title={`${call.name}${call.summary ? ` — ${call.summary}` : ""}`}
      onClick={(event) => {
        event.stopPropagation();
        onSelect(call.id);
      }}
      className={`inline-flex max-w-full items-center gap-1 rounded-full px-2 py-0.5 text-[0.6875rem] font-medium transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-amber-500 ${styles}`}
    >
      <Wrench className="size-3 shrink-0 opacity-60" aria-hidden />
      {icon}
      <span className="truncate font-mono">{label}</span>
    </button>
  );
}
