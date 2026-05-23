import {
  Bot,
  ImageIcon,
  Loader2,
  MessageCircle,
  RotateCcw,
  Send,
  Sparkles,
  X,
} from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { useLocation, useNavigate, useRevalidator } from "react-router";

import { ModelDropdown } from "~/components/model-dropdown";
import { BubbleRow, type BubbleMessage } from "~/components/agent-chat/bubble-row";
import { DebugDialog } from "~/components/agent-chat/debug-dialog";
import { ToolCallDialog } from "~/components/agent-chat/tool-call-dialog";
import { getAgentBaseURL } from "~/lib/agent-base-url";
import {
  isInternalToolName,
  parseAssistantResponse,
  uniqueUIActions,
  type UIAction,
} from "~/lib/agent-ui-actions";
import {
  agentAppName,
  buildModelContext,
  ensureSession,
  getSessionID,
  getUserID,
  randomID,
  saveSavedPrefs,
  startNewSession,
} from "~/lib/agent-session";
import {
  buildAgentMessage,
  getAppContext,
  getHighlightedText,
} from "~/lib/app-context";
import { readAgentStream } from "~/lib/agent-stream";
import {
  ToolContextProvider,
  createToolContextStore,
  useToolCall,
  type ToolContextStore,
} from "~/lib/tool-context";
import { useAgentPreferencesState } from "~/state/agent-preferences/context";
import { AgentPreferencesActionType } from "~/state/agent-preferences/types";

type ChatMessage = BubbleMessage;

function initialMessages(): ChatMessage[] {
  return [
    {
      id: "welcome",
      role: "assistant",
      content:
        "Hi, I can help manage recipes, inspect the current recipe list, or create and update recipes.",
    },
  ];
}

function replaceMessageContent(
  messages: ChatMessage[],
  messageID: string,
  content: string,
  uiActions?: UIAction[],
): ChatMessage[] {
  return messages.map((message) =>
    message.id === messageID ? { ...message, content, uiActions } : message,
  );
}

function formatRawEvents(rawEvents: string[]): string {
  return rawEvents
    .map((event) => {
      try {
        return JSON.stringify(JSON.parse(event), null, 2);
      } catch {
        return event;
      }
    })
    .join("\n\n");
}

function ToolCallDialogContainer({
  id,
  onClose,
}: {
  id: string;
  onClose: () => void;
}) {
  const call = useToolCall(id);
  if (call == null) return null;
  return <ToolCallDialog call={call} onClose={onClose} />;
}

export function AgentChat() {
  const toolStore = useMemo<ToolContextStore>(() => createToolContextStore(), []);
  const [isOpen, setIsOpen] = useState(false);
  const [messages, setMessages] = useState<ChatMessage[]>(initialMessages);
  const [draft, setDraft] = useState("");
  const [isSending, setIsSending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [debugMessage, setDebugMessage] = useState<ChatMessage | null>(null);
  const [selectedToolCallID, setSelectedToolCallID] = useState<string | null>(
    null,
  );
  const bottomRef = useRef<HTMLDivElement | null>(null);
  const inputRef = useRef<HTMLTextAreaElement | null>(null);
  const baseURL = useMemo(getAgentBaseURL, []);
  const location = useLocation();
  const navigate = useNavigate();
  const revalidator = useRevalidator();
  const { state: prefs, dispatch: prefsDispatch } = useAgentPreferencesState();

  useEffect(() => {
    if (prefs.options == null) return;
    if (prefs.agentModel == null && prefs.imageModel == null) return;
    saveSavedPrefs({
      agentModel: prefs.agentModel,
      imageModel: prefs.imageModel,
    });
  }, [prefs.options, prefs.agentModel, prefs.imageModel]);

  useEffect(() => {
    if (!isOpen) return;
    bottomRef.current?.scrollIntoView({ block: "end" });
    inputRef.current?.focus();
  }, [isOpen, messages]);

  function startNewChat() {
    if (isSending) return;
    startNewSession();
    setMessages(initialMessages());
    setDraft("");
    setError(null);
    toolStore.reset();
    inputRef.current?.focus();
  }

  function runSideEffects(uiActions: UIAction[]) {
    for (const action of uiActions) {
      if (action.type === "navigate_recipe") {
        void navigate(`/recipe/${encodeURIComponent(action.recipeId)}`);
      } else if (action.type === "navigate_recipe_list") {
        void navigate("/");
      } else if (action.type === "refresh_current_screen") {
        void revalidator.revalidate();
      }
    }
  }

  async function sendMessage() {
    const text = draft.trim();
    if (text === "" || isSending) return;
    const appContext = {
      ...getAppContext(location.pathname),
      highlightedText: getHighlightedText(),
    };

    const userID = getUserID();
    const sessionID = getSessionID();
    const body: Record<string, unknown> = {
      appName: agentAppName,
      userId: userID,
      sessionId: sessionID,
      streaming: true,
      newMessage: {
        role: "user",
        parts: [{ text: buildAgentMessage(text, appContext) }],
      },
    };
    const modelContext = buildModelContext(prefs);
    if (modelContext != null) {
      body.modelContext = modelContext;
    }

    const userMessage: ChatMessage = {
      id: randomID("user"),
      role: "user",
      content: text,
      rawDebug: JSON.stringify(body, null, 2),
    };
    let currentAssistantID = randomID("assistant");
    setMessages((current) => [
      ...current,
      userMessage,
      { id: currentAssistantID, role: "assistant", content: "" },
    ]);
    setDraft("");
    setError(null);
    setIsSending(true);

    let pendingNextBubble = false;
    let lastFinalizedBubbleID: string | null = null;

    function ensureCurrentBubble() {
      if (!pendingNextBubble) return;
      currentAssistantID = randomID("assistant");
      const newID = currentAssistantID;
      setMessages((current) => [
        ...current,
        { id: newID, role: "assistant", content: "" },
      ]);
      pendingNextBubble = false;
    }

    try {
      await ensureSession(baseURL, userID, sessionID);

      const res = await fetch(`${baseURL}/run_sse`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      if (!res.ok) {
        throw new Error(`Agent request failed (${res.status})`);
      }

      await readAgentStream(res, {
        onTextProgress: (chunk, streamUIActions) => {
          ensureCurrentBubble();
          const id = currentAssistantID;
          const parsed = parseAssistantResponse(chunk);
          const uiActions = uniqueUIActions([
            ...streamUIActions,
            ...parsed.uiActions,
          ]);
          setMessages((current) =>
            replaceMessageContent(current, id, parsed.content, uiActions),
          );
        },

        onToolEvents: (toolEvents) => {
          for (const call of toolEvents.calls) {
            if (isInternalToolName(call.name)) continue;
            const isNew = toolStore.register(call);
            if (!isNew) continue;
            ensureCurrentBubble();
            const ownerId = currentAssistantID;
            setMessages((current) =>
              current.map((message) =>
                message.id === ownerId
                  ? {
                      ...message,
                      toolCallIds: [...(message.toolCallIds ?? []), call.id],
                    }
                  : message,
              ),
            );
          }
          for (const response of toolEvents.responses) {
            if (isInternalToolName(response.name)) continue;
            toolStore.applyResponse(response.id, response.response);
          }
        },

        onTurnFinalize: (chunk, streamUIActions, rawEvents) => {
          const id = currentAssistantID;
          const parsed = parseAssistantResponse(chunk);
          const uiActions = uniqueUIActions([
            ...streamUIActions,
            ...parsed.uiActions,
          ]);
          const rawDebug = formatRawEvents(rawEvents);
          setMessages((current) =>
            current.map((message) =>
              message.id === id
                ? {
                    ...message,
                    content: parsed.content,
                    uiActions,
                    rawDebug:
                      message.rawDebug != null && message.rawDebug !== ""
                        ? `${message.rawDebug}\n\n${rawDebug}`
                        : rawDebug,
                  }
                : message,
            ),
          );
          runSideEffects(uiActions);
          lastFinalizedBubbleID = id;
          pendingNextBubble = true;
        },

        onTurnAuthoritativeText: (chunk, streamUIActions, rawEvents) => {
          const targetId = lastFinalizedBubbleID ?? currentAssistantID;
          const parsed = parseAssistantResponse(chunk);
          const uiActions = uniqueUIActions([
            ...streamUIActions,
            ...parsed.uiActions,
          ]);
          const rawDebugAppend = formatRawEvents(rawEvents);
          setMessages((current) =>
            current.map((message) =>
              message.id === targetId
                ? {
                    ...message,
                    content: parsed.content,
                    uiActions,
                    rawDebug:
                      message.rawDebug != null && message.rawDebug !== ""
                        ? `${message.rawDebug}\n\n${rawDebugAppend}`
                        : rawDebugAppend,
                  }
                : message,
            ),
          );
          // Side effects already fired in onTurnFinalize; do not re-fire.
        },
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Agent request failed.");
      const lastID = currentAssistantID;
      setMessages((current) =>
        current.filter(
          (message) => message.id !== lastID || message.content !== "",
        ),
      );
    } finally {
      setIsSending(false);
    }
  }

  return (
    <ToolContextProvider store={toolStore}>
      <div className="fixed bottom-5 right-5 z-50">
        {isOpen ? (
          <section className="flex h-[min(36rem,calc(100vh-2.5rem))] w-[min(28rem,calc(100vw-2.5rem))] flex-col overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-2xl dark:border-zinc-800 dark:bg-zinc-900">
            <header className="flex items-center justify-between border-b border-zinc-200 px-4 py-3 dark:border-zinc-800">
              <div className="flex items-center gap-2">
                <span className="flex size-8 items-center justify-center rounded-full bg-amber-100 text-amber-800 dark:bg-amber-950/80 dark:text-amber-200">
                  <Bot className="size-4" aria-hidden />
                </span>
                <div>
                  <h2 className="text-sm font-semibold text-zinc-900 dark:text-zinc-50">
                    Recipe copilot
                  </h2>
                  <p className="text-xs text-zinc-500 dark:text-zinc-400">
                    Powered by the agent API
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-1">
                <button
                  type="button"
                  className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1.5 text-xs font-medium text-zinc-500 transition-colors hover:bg-zinc-100 hover:text-zinc-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-zinc-400 disabled:pointer-events-none disabled:opacity-50 dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
                  onClick={startNewChat}
                  disabled={isSending}
                  aria-label="Start new chat"
                >
                  <RotateCcw className="size-3.5" aria-hidden />
                  New
                </button>
                <button
                  type="button"
                  className="rounded-full p-2 text-zinc-500 transition-colors hover:bg-zinc-100 hover:text-zinc-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
                  onClick={() => setIsOpen(false)}
                  aria-label="Close recipe copilot"
                >
                  <X className="size-4" aria-hidden />
                </button>
              </div>
            </header>

            <div className="flex-1 space-y-3 overflow-y-auto bg-zinc-50/80 px-4 py-4 dark:bg-zinc-950/40">
              {messages.map((message) => (
                <BubbleRow
                  key={message.id}
                  message={message}
                  onDebugClick={setDebugMessage}
                  onToolSelect={setSelectedToolCallID}
                />
              ))}
              {error ? (
                <div className="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800 dark:border-red-900/60 dark:bg-red-950/40 dark:text-red-200">
                  {error}
                </div>
              ) : null}
              <div ref={bottomRef} />
            </div>

            <form
              className="border-t border-zinc-200 bg-white p-3 dark:border-zinc-800 dark:bg-zinc-900"
              onSubmit={(event) => {
                event.preventDefault();
                void sendMessage();
              }}
            >
              <div className="flex flex-col gap-2">
                <textarea
                  ref={inputRef}
                  value={draft}
                  onChange={(event) => setDraft(event.target.value)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" && !event.shiftKey) {
                      event.preventDefault();
                      void sendMessage();
                    }
                  }}
                  rows={2}
                  placeholder="Ask the copilot..."
                  className="min-h-10 w-full resize-none rounded-xl border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition focus:border-zinc-400 focus:ring-2 focus:ring-zinc-200 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-100 dark:focus:border-zinc-500 dark:focus:ring-zinc-800"
                />
                <div className="flex items-center gap-2">
                  {prefs.options != null &&
                  prefs.options.imageOptions.length > 1 ? (
                    <ModelDropdown
                      ariaLabel="Image model"
                      icon={<ImageIcon className="size-3.5" />}
                      options={prefs.options.imageOptions.map((opt) => ({
                        id: opt.id,
                        label: opt.model,
                      }))}
                      value={prefs.imageModel}
                      disabled={isSending}
                      onChange={(id) =>
                        prefsDispatch({
                          type: AgentPreferencesActionType.SELECT_IMAGE,
                          data: id,
                        })
                      }
                    />
                  ) : null}
                  {prefs.options != null &&
                  prefs.options.agentOptions.length > 1 ? (
                    <ModelDropdown
                      ariaLabel="Agent model"
                      icon={<Sparkles className="size-3.5" />}
                      options={prefs.options.agentOptions.map((opt) => ({
                        id: opt.id,
                        label: opt.model,
                      }))}
                      value={prefs.agentModel}
                      disabled={isSending}
                      onChange={(id) =>
                        prefsDispatch({
                          type: AgentPreferencesActionType.SELECT_AGENT,
                          data: id,
                        })
                      }
                    />
                  ) : null}
                  <div className="flex-1" />
                  <button
                    type="submit"
                    disabled={draft.trim() === "" || isSending}
                    className="flex size-10 shrink-0 items-center justify-center rounded-full bg-amber-600 text-white shadow-sm transition hover:bg-amber-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-amber-500 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-70 dark:focus-visible:ring-offset-zinc-900"
                    aria-label={isSending ? "Agent is working" : "Send message"}
                    aria-busy={isSending}
                  >
                    {isSending ? (
                      <Loader2 className="size-4 animate-spin" aria-hidden />
                    ) : (
                      <Send className="size-4" aria-hidden />
                    )}
                  </button>
                </div>
              </div>
            </form>
          </section>
        ) : (
          <button
            type="button"
            className="flex size-14 items-center justify-center rounded-full bg-amber-600 text-white shadow-xl transition hover:scale-105 hover:bg-amber-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-amber-500 focus-visible:ring-offset-2 dark:focus-visible:ring-offset-zinc-950"
            onClick={() => setIsOpen(true)}
            aria-label="Open recipe copilot"
          >
            <MessageCircle className="size-6" aria-hidden />
          </button>
        )}
        {debugMessage != null ? (
          <DebugDialog
            title={debugMessage.role === "user" ? "Raw request" : "Raw response"}
            body={debugMessage.rawDebug ?? ""}
            onClose={() => setDebugMessage(null)}
          />
        ) : null}
        {selectedToolCallID != null ? (
          <ToolCallDialogContainer
            id={selectedToolCallID}
            onClose={() => setSelectedToolCallID(null)}
          />
        ) : null}
      </div>
    </ToolContextProvider>
  );
}
