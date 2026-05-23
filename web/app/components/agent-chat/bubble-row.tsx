import { memo } from "react";

import type { UIAction } from "~/lib/agent-ui-actions";

import { MarkdownMessage } from "./markdown-message";
import { ToolCallWidget } from "./tool-call-widget";
import { UIActionChips } from "./ui-action-chip";

export type BubbleMessage = {
  id: string;
  role: "user" | "assistant";
  content: string;
  uiActions?: UIAction[];
  toolCallIds?: string[];
  rawDebug?: string;
};

function BubbleRowImpl({
  message,
  onDebugClick,
  onToolSelect,
}: {
  message: BubbleMessage;
  onDebugClick: (message: BubbleMessage) => void;
  onToolSelect: (id: string) => void;
}) {
  const isClickable = message.rawDebug != null && message.rawDebug !== "";
  const hasTools = message.toolCallIds != null && message.toolCallIds.length > 0;
  const hasUIActions = message.uiActions != null && message.uiActions.length > 0;

  const containerClass = [
    "flex",
    message.role === "user" ? "justify-end" : "justify-start",
  ].join(" ");

  const bubbleClass = [
    "max-w-[85%] rounded-2xl px-3 py-2 text-sm shadow-sm",
    isClickable
      ? "cursor-pointer transition hover:shadow-md focus:outline-none focus-visible:ring-2 focus-visible:ring-amber-500"
      : "",
    message.role === "user"
      ? "bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900"
      : "border border-zinc-200 bg-white text-zinc-800 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-100",
  ].join(" ");

  return (
    <div className={containerClass}>
      <div
        role={isClickable ? "button" : undefined}
        tabIndex={isClickable ? 0 : undefined}
        title={isClickable ? "View raw debug payload" : undefined}
        onClick={
          isClickable
            ? (event) => {
                const target = event.target as HTMLElement;
                if (target.closest("a, button")) return;
                onDebugClick(message);
              }
            : undefined
        }
        onKeyDown={
          isClickable
            ? (event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  onDebugClick(message);
                }
              }
            : undefined
        }
        className={bubbleClass}
      >
        {message.role === "assistant" ? (
          <div className="space-y-2">
            {message.content === "" && !hasTools ? (
              <span className="text-zinc-500 dark:text-zinc-400">
                Thinking...
              </span>
            ) : message.content !== "" ? (
              <MarkdownMessage content={message.content} />
            ) : null}
            {hasTools ? (
              <div
                className={[
                  "flex flex-wrap gap-1.5",
                  message.content !== ""
                    ? "border-t border-zinc-100 pt-2 dark:border-zinc-800"
                    : "",
                ].join(" ")}
              >
                {message.toolCallIds!.map((id) => (
                  <ToolCallWidget key={id} id={id} onSelect={onToolSelect} />
                ))}
              </div>
            ) : null}
            {hasUIActions ? <UIActionChips actions={message.uiActions!} /> : null}
          </div>
        ) : message.content === "" ? (
          <span className="text-zinc-500 dark:text-zinc-400">Thinking...</span>
        ) : (
          <p className="whitespace-pre-wrap">{message.content}</p>
        )}
      </div>
    </div>
  );
}

export const BubbleRow = memo(BubbleRowImpl);
