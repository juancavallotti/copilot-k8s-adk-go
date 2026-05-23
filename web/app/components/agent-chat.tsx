import {
  AlertCircle,
  Bot,
  CheckCircle2,
  ImageIcon,
  Loader2,
  MessageCircle,
  RotateCcw,
  Send,
  Sparkles,
  Wrench,
  X,
} from "lucide-react";
import ReactMarkdown from "react-markdown";
import { useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { useLocation, useNavigate, useRevalidator } from "react-router";
import {
  applyToolResponse,
  extractToolEventsFromEvent,
  extractUIActionsFromEvent,
  isInternalToolName,
  makePendingToolCall,
  parseAssistantResponse,
  uniqueUIActions,
  type ToolCall,
  type UIAction,
} from "~/lib/agent-ui-actions";
import { ModelDropdown } from "~/components/model-dropdown";
import { getAgentBaseURL } from "~/lib/agent-base-url";
import { useAgentPreferencesState } from "~/state/agent-preferences/context";
import {
  AgentPreferencesActionType,
  type ProvidersResponse,
} from "~/state/agent-preferences/types";

const agentAppName = "recipe_copilot";
const modelPrefsStorageKey = "recipes-agent-model-prefs";

// Session ID lives in memory only — reloading the page generates a fresh
// session so the agent does not carry over context from a prior visit.
let currentSessionID: string | null = null;

function saveSavedPrefs(prefs: { agentModel: string | null; imageModel: string | null }) {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(modelPrefsStorageKey, JSON.stringify(prefs));
  } catch {
    // ignore quota / disabled storage
  }
}

function buildModelContext(prefs: {
  options: ProvidersResponse | null;
  agentModel: string | null;
  imageModel: string | null;
}): { agentModel: string; imageModel: string } | null {
  if (prefs.options == null || prefs.agentModel == null || prefs.imageModel == null) {
    return null;
  }
  if (
    prefs.agentModel === prefs.options.defaultAgentModel &&
    prefs.imageModel === prefs.options.defaultImageModel
  ) {
    return null;
  }
  return { agentModel: prefs.agentModel, imageModel: prefs.imageModel };
}

type ChatMessage = {
  id: string;
  role: "user" | "assistant";
  content: string;
  uiActions?: UIAction[];
  toolCalls?: ToolCall[];
  rawDebug?: string;
};

type AppContext = {
  screen: "recipe_list" | "specific_recipe" | "create_recipe" | "other";
  path: string;
  recipeId?: string;
  highlightedText?: string;
};

type AgentEvent = {
  author?: string;
  partial?: boolean;
  turnComplete?: boolean;
  errorMessage?: string;
  content?: {
    parts?: Array<{
      text?: string;
      functionResponse?: unknown;
      function_response?: unknown;
    }>;
  };
};

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

function randomID(prefix: string): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return `${prefix}-${crypto.randomUUID()}`;
  }
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}


function getUserID(): string {
  const key = "recipes-agent-user-id";
  const existing = window.localStorage.getItem(key);
  if (existing != null && existing !== "") return existing;
  const next = randomID("web-user");
  window.localStorage.setItem(key, next);
  return next;
}

function getSessionID(): string {
  if (currentSessionID != null && currentSessionID !== "") return currentSessionID;
  currentSessionID = randomID("session");
  return currentSessionID;
}

function startNewSession(): string {
  currentSessionID = randomID("session");
  return currentSessionID;
}

async function ensureSession(baseURL: string, userID: string, sessionID: string) {
  const sessionURL = `${baseURL}/apps/${encodeURIComponent(agentAppName)}/users/${encodeURIComponent(
    userID,
  )}/sessions/${encodeURIComponent(sessionID)}`;

  const existing = await fetch(sessionURL);
  if (existing.ok) return;
  if (existing.status !== 404 && existing.status !== 500) {
    throw new Error(`Could not check agent session (${existing.status})`);
  }

  const res = await fetch(sessionURL, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: "{}",
  });
  if (!res.ok && res.status !== 409) {
    throw new Error(`Could not start agent session (${res.status})`);
  }
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

function extractText(event: AgentEvent): string {
  return (
    event.content?.parts
      ?.map((part) => part.text ?? "")
      .filter((text) => text !== "")
      .join("") ?? ""
  );
}

type ToolEventBatch = {
  calls: Array<{ id: string; name: string; args?: Record<string, unknown> }>;
  responses: Array<{ id: string; name: string; response: unknown }>;
};

type BubbleStreamHandlers = {
  onProgress: (text: string, uiActions: UIAction[]) => void;
  onBubbleFinish: (
    text: string,
    uiActions: UIAction[],
    rawEvents: string[],
  ) => void;
  onToolEvents: (events: ToolEventBatch) => void;
};

async function readAgentStream(
  response: Response,
  handlers: BubbleStreamHandlers,
): Promise<void> {
  if (response.body == null) {
    throw new Error("Agent response did not include a stream.");
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let accumulatedText = "";
  let accumulatedUIActions: UIAction[] = [];
  let bubbleRawEvents: string[] = [];
  let bubbleHasContent = false;

  function finalizeBubble() {
    handlers.onBubbleFinish(
      accumulatedText,
      accumulatedUIActions,
      bubbleRawEvents,
    );
    accumulatedText = "";
    accumulatedUIActions = [];
    bubbleRawEvents = [];
    bubbleHasContent = false;
  }

  function processRawEvent(rawEvent: string) {
    const data = rawEvent
      .split("\n")
      .filter((line) => line.startsWith("data:"))
      .map((line) => line.slice(5).trimStart())
      .join("\n");
    if (data === "") return;
    bubbleRawEvents.push(data);

    const parsed = JSON.parse(data) as AgentEvent | { error?: string };
    if ("error" in parsed && typeof parsed.error === "string") {
      throw new Error(parsed.error);
    }
    const event = parsed as AgentEvent;
    const eventUIActions = extractUIActionsFromEvent(event);
    if (eventUIActions.length > 0) {
      accumulatedUIActions = uniqueUIActions([
        ...accumulatedUIActions,
        ...eventUIActions,
      ]);
      bubbleHasContent = true;
    }
    const toolEvents = extractToolEventsFromEvent(event);
    // Tool events are managed by the caller (assigned to the bubble that
    // created the call, then updated in place when the response arrives).
    // They do not flip bubbleHasContent: a functionCall often carries
    // turnComplete=true, and finalizing on that would split the bubble
    // before the assistant's text response arrives.
    if (toolEvents.calls.length > 0 || toolEvents.responses.length > 0) {
      handlers.onToolEvents(toolEvents);
    }
    const text = extractText(event);
    if (text !== "") {
      if (event.partial) {
        accumulatedText = text.startsWith(accumulatedText)
          ? text
          : `${accumulatedText}${text}`;
      } else if (
        accumulatedText === "" ||
        text.startsWith(accumulatedText)
      ) {
        accumulatedText = text;
      } else if (!accumulatedText.endsWith(text)) {
        accumulatedText = `${accumulatedText}${text}`;
      }
      bubbleHasContent = true;
    }

    if (text !== "" || eventUIActions.length > 0) {
      handlers.onProgress(accumulatedText, accumulatedUIActions);
    }

    if (event.turnComplete === true && bubbleHasContent) {
      finalizeBubble();
    }
  }

  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      buffer += decoder.decode();
      break;
    }

    buffer += decoder.decode(value, { stream: true });
    const events = buffer.split(/\n\n/);
    buffer = events.pop() ?? "";

    for (const rawEvent of events) {
      processRawEvent(rawEvent);
    }
  }
  if (buffer.trim() !== "") {
    processRawEvent(buffer);
  }

  if (bubbleHasContent || bubbleRawEvents.length > 0) {
    finalizeBubble();
  }
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

function getHighlightedText(): string | undefined {
  const text = window.getSelection()?.toString().replace(/\s+/g, " ").trim();
  if (text == null || text === "") return undefined;
  return text.slice(0, 2000);
}

function getAppContext(path: string): AppContext {
  if (path === "/") {
    return { screen: "recipe_list", path };
  }
  const recipeMatch = path.match(/^\/recipe\/([^/]+)(?:\/edit)?\/?$/);
  if (recipeMatch != null) {
    return {
      screen: "specific_recipe",
      path,
      recipeId: decodeURIComponent(recipeMatch[1]),
    };
  }
  if (path === "/create") {
    return { screen: "create_recipe", path };
  }
  return { screen: "other", path };
}

function buildAgentMessage(userMessage: string, appContext: AppContext): string {
  return JSON.stringify(
    {
      appContext,
      userMessage,
    },
    null,
    2,
  );
}

function describeToolCall(call: ToolCall): string {
  if (call.summary != null && call.summary !== "") return call.summary;
  return call.name;
}

function ToolCallChip({ call }: { call: ToolCall }) {
  const label = describeToolCall(call);
  const styles =
    call.status === "error"
      ? "bg-red-50 text-red-800 dark:bg-red-950/60 dark:text-red-200"
      : call.status === "pending"
        ? "bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200"
        : "bg-emerald-50 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-200";
  const icon =
    call.status === "error" ? (
      <AlertCircle className="size-3" aria-hidden />
    ) : call.status === "pending" ? (
      <Loader2 className="size-3 animate-spin" aria-hidden />
    ) : (
      <CheckCircle2 className="size-3" aria-hidden />
    );
  return (
    <span
      title={`${call.name}${call.summary ? ` — ${call.summary}` : ""}`}
      className={`inline-flex max-w-full items-center gap-1 rounded-full px-2 py-0.5 text-[0.6875rem] font-medium ${styles}`}
    >
      <Wrench className="size-3 shrink-0 opacity-60" aria-hidden />
      {icon}
      <span className="truncate font-mono">{label}</span>
    </span>
  );
}

function describeUIAction(action: UIAction): string {
  switch (action.type) {
    case "navigate_recipe":
      return "Open recipe";
    case "navigate_recipe_list":
      return "Open recipe list";
    case "refresh_current_screen":
      return "Refresh current screen";
  }
}

function MarkdownMessage({ content }: { content: string }) {
  return (
    <ReactMarkdown
      components={{
        a: ({ children, ...props }) => (
          <a
            {...props}
            className="font-medium text-amber-700 underline-offset-2 hover:underline dark:text-amber-300"
            target="_blank"
            rel="noreferrer"
          >
            {children}
          </a>
        ),
        code: ({ children }) => (
          <code className="rounded bg-zinc-100 px-1 py-0.5 text-[0.8125em] text-zinc-900 dark:bg-zinc-800 dark:text-zinc-100">
            {children}
          </code>
        ),
        ol: ({ children }) => (
          <ol className="list-decimal space-y-1 pl-5">{children}</ol>
        ),
        p: ({ children }) => <p className="leading-relaxed">{children}</p>,
        pre: ({ children }) => (
          <pre className="overflow-auto rounded-lg bg-zinc-100 p-3 text-xs dark:bg-zinc-950">
            {children}
          </pre>
        ),
        ul: ({ children }) => <ul className="list-disc space-y-1 pl-5">{children}</ul>,
      }}
    >
      {content}
    </ReactMarkdown>
  );
}

function HighlightedJSON({ source }: { source: string }) {
  const tokens: ReactNode[] = [];
  const regex =
    /("(?:\\.|[^"\\])*")(\s*:)?|\b(true|false|null)\b|(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)/g;
  let lastIndex = 0;
  let key = 0;
  let match: RegExpExecArray | null;
  while ((match = regex.exec(source)) !== null) {
    if (match.index > lastIndex) {
      tokens.push(source.slice(lastIndex, match.index));
    }
    if (match[1] != null) {
      const isKey = match[2] != null;
      tokens.push(
        <span
          key={key++}
          className={isKey ? "text-amber-300" : "text-emerald-300"}
        >
          {match[1]}
        </span>,
      );
      if (isKey) tokens.push(match[2]);
    } else if (match[3] != null) {
      tokens.push(
        <span
          key={key++}
          className={match[3] === "null" ? "text-zinc-500" : "text-rose-300"}
        >
          {match[3]}
        </span>,
      );
    } else if (match[4] != null) {
      tokens.push(
        <span key={key++} className="text-sky-300">
          {match[4]}
        </span>,
      );
    }
    lastIndex = regex.lastIndex;
  }
  if (lastIndex < source.length) {
    tokens.push(source.slice(lastIndex));
  }
  return <>{tokens}</>;
}

function DebugDialog({
  message,
  onClose,
}: {
  message: ChatMessage;
  onClose: () => void;
}) {
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") onClose();
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  const title = message.role === "user" ? "Raw request" : "Raw response";
  const body = message.rawDebug ?? "";

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={title}
      className="fixed inset-0 z-[60] flex items-center justify-center bg-zinc-950/85 p-4"
      onClick={onClose}
    >
      <div
        className="flex max-h-[85vh] w-full max-w-3xl flex-col overflow-hidden rounded-2xl border border-zinc-800 bg-zinc-950 text-zinc-100 shadow-2xl"
        onClick={(event) => event.stopPropagation()}
      >
        <header className="flex items-center justify-between border-b border-zinc-800 px-4 py-3">
          <h3 className="text-sm font-semibold">{title}</h3>
          <div className="flex items-center gap-1">
            <button
              type="button"
              className="rounded-full px-2.5 py-1 text-xs font-medium text-zinc-400 transition-colors hover:bg-zinc-800 hover:text-zinc-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-zinc-500"
              onClick={() => {
                if (typeof navigator !== "undefined" && navigator.clipboard) {
                  void navigator.clipboard.writeText(body);
                }
              }}
            >
              Copy
            </button>
            <button
              type="button"
              className="rounded-full p-2 text-zinc-400 transition-colors hover:bg-zinc-800 hover:text-zinc-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-zinc-500"
              onClick={onClose}
              aria-label="Close debug dialog"
            >
              <X className="size-4" aria-hidden />
            </button>
          </div>
        </header>
        <pre className="flex-1 overflow-auto whitespace-pre-wrap break-words px-4 py-3 font-mono text-xs leading-relaxed text-zinc-300">
          {body === "" ? (
            "(no debug data captured)"
          ) : (
            <HighlightedJSON source={body} />
          )}
        </pre>
      </div>
    </div>
  );
}

export function AgentChat() {
  const [isOpen, setIsOpen] = useState(false);
  const [messages, setMessages] = useState<ChatMessage[]>(initialMessages);
  const [draft, setDraft] = useState("");
  const [isSending, setIsSending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [debugMessage, setDebugMessage] = useState<ChatMessage | null>(null);
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
    inputRef.current?.focus();
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
    // Tool call ids → the bubble (message id) that owns them. Each call
    // is created in exactly one bubble, and its response updates that
    // same bubble's chip in place — even if the response arrives in a
    // later bubble.
    const toolCallOwners = new Map<string, string>();

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
        onProgress: (chunk, streamUIActions) => {
          ensureCurrentBubble();
          const id = currentAssistantID;
          const parsed = parseAssistantResponse(chunk);
          setMessages((current) =>
            replaceMessageContent(
              current,
              id,
              parsed.content,
              uniqueUIActions([...streamUIActions, ...parsed.uiActions]),
            ),
          );
        },
        onToolEvents: (toolEvents) => {
          for (const call of toolEvents.calls) {
            if (isInternalToolName(call.name)) continue;
            if (toolCallOwners.has(call.id)) continue;
            ensureCurrentBubble();
            const ownerId = currentAssistantID;
            toolCallOwners.set(call.id, ownerId);
            const newCall = makePendingToolCall(call);
            setMessages((current) =>
              current.map((message) =>
                message.id === ownerId
                  ? {
                      ...message,
                      toolCalls: [...(message.toolCalls ?? []), newCall],
                    }
                  : message,
              ),
            );
          }
          for (const response of toolEvents.responses) {
            if (isInternalToolName(response.name)) continue;
            const ownerId = toolCallOwners.get(response.id);
            if (ownerId == null) continue;
            setMessages((current) =>
              current.map((message) => {
                if (message.id !== ownerId) return message;
                const toolCalls = (message.toolCalls ?? []).map((tc) =>
                  tc.id === response.id
                    ? applyToolResponse(tc, response.response)
                    : tc,
                );
                return { ...message, toolCalls };
              }),
            );
          }
        },
        onBubbleFinish: (chunk, streamUIActions, rawEvents) => {
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
                    rawDebug,
                  }
                : message,
            ),
          );
          for (const action of uiActions) {
            if (action.type === "navigate_recipe") {
              void navigate(`/recipe/${encodeURIComponent(action.recipeId)}`);
            } else if (action.type === "navigate_recipe_list") {
              void navigate("/");
            } else if (action.type === "refresh_current_screen") {
              void revalidator.revalidate();
            }
          }
          pendingNextBubble = true;
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
            {messages.map((message) => {
              const isClickable =
                message.rawDebug != null && message.rawDebug !== "";
              return (
              <div
                key={message.id}
                className={[
                  "flex",
                  message.role === "user" ? "justify-end" : "justify-start",
                ].join(" ")}
              >
                <div
                  role={isClickable ? "button" : undefined}
                  tabIndex={isClickable ? 0 : undefined}
                  title={isClickable ? "View raw debug payload" : undefined}
                  onClick={
                    isClickable
                      ? (event) => {
                          const target = event.target as HTMLElement;
                          if (target.closest("a, button")) return;
                          setDebugMessage(message);
                        }
                      : undefined
                  }
                  onKeyDown={
                    isClickable
                      ? (event) => {
                          if (event.key === "Enter" || event.key === " ") {
                            event.preventDefault();
                            setDebugMessage(message);
                          }
                        }
                      : undefined
                  }
                  className={[
                    "max-w-[85%] rounded-2xl px-3 py-2 text-sm shadow-sm",
                    isClickable
                      ? "cursor-pointer transition hover:shadow-md focus:outline-none focus-visible:ring-2 focus-visible:ring-amber-500"
                      : "",
                    message.role === "user"
                      ? "bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900"
                      : "border border-zinc-200 bg-white text-zinc-800 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-100",
                  ].join(" ")}
                >
                  {message.role === "assistant" ? (
                    <div className="space-y-2">
                      {message.content === "" &&
                      (message.toolCalls == null ||
                        message.toolCalls.length === 0) ? (
                        <span className="text-zinc-500 dark:text-zinc-400">
                          Thinking...
                        </span>
                      ) : message.content !== "" ? (
                        <MarkdownMessage content={message.content} />
                      ) : null}
                      {message.toolCalls != null && message.toolCalls.length > 0 ? (
                        <div
                          className={[
                            "flex flex-wrap gap-1.5",
                            message.content !== ""
                              ? "border-t border-zinc-100 pt-2 dark:border-zinc-800"
                              : "",
                          ].join(" ")}
                        >
                          {message.toolCalls.map((call) => (
                            <ToolCallChip key={call.id} call={call} />
                          ))}
                        </div>
                      ) : null}
                      {message.uiActions != null && message.uiActions.length > 0 ? (
                        <div className="flex flex-wrap gap-1.5 border-t border-zinc-100 pt-2 dark:border-zinc-800">
                          {message.uiActions.map((action, index) => (
                            <span
                              key={`${action.type}-${index}`}
                              className="rounded-full bg-amber-50 px-2 py-0.5 text-[0.6875rem] font-medium text-amber-800 dark:bg-amber-950/70 dark:text-amber-200"
                            >
                              {describeUIAction(action)}
                            </span>
                          ))}
                        </div>
                      ) : null}
                    </div>
                  ) : message.content === "" ? (
                    <span className="text-zinc-500 dark:text-zinc-400">
                      Thinking...
                    </span>
                  ) : (
                    <p className="whitespace-pre-wrap">{message.content}</p>
                  )}
                </div>
              </div>
              );
            })}
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
          message={debugMessage}
          onClose={() => setDebugMessage(null)}
        />
      ) : null}
    </div>
  );
}
