import { Bot, MessageCircle, RotateCcw, Send, X } from "lucide-react";
import ReactMarkdown from "react-markdown";
import { useEffect, useMemo, useRef, useState } from "react";
import { useLocation, useNavigate, useRevalidator } from "react-router";

const agentAppName = "recipe_copilot";
const sessionStorageKey = "recipes-agent-session-id";

type ChatMessage = {
  id: string;
  role: "user" | "assistant";
  content: string;
  uiActions?: UIAction[];
};

type UIAction =
  | { type: "navigate_recipe"; recipeId: string }
  | { type: "navigate_recipe_list" }
  | { type: "refresh_current_screen" };

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

function normalizeBaseURL(raw: string): string {
  return raw.replace(/\/+$/, "");
}

function getAgentBaseURL(): string {
  const fromEnv = import.meta.env.VITE_AGENT_API_BASE_URL;
  if (typeof fromEnv === "string" && fromEnv.trim() !== "") {
    return normalizeBaseURL(fromEnv.trim());
  }
  return import.meta.env.DEV ? "http://localhost:4100/agent" : "/agent";
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
  const existing = window.localStorage.getItem(sessionStorageKey);
  if (existing != null && existing !== "") return existing;
  const next = randomID("session");
  window.localStorage.setItem(sessionStorageKey, next);
  return next;
}

function startNewSession(): string {
  const next = randomID("session");
  window.localStorage.setItem(sessionStorageKey, next);
  return next;
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

async function readAgentStream(
  response: Response,
  onText: (text: string) => void,
): Promise<string> {
  if (response.body == null) {
    throw new Error("Agent response did not include a stream.");
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let accumulatedText = "";

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });
    const events = buffer.split(/\n\n/);
    buffer = events.pop() ?? "";

    for (const rawEvent of events) {
      const data = rawEvent
        .split("\n")
        .filter((line) => line.startsWith("data:"))
        .map((line) => line.slice(5).trimStart())
        .join("\n");
      if (data === "") continue;

      const parsed = JSON.parse(data) as AgentEvent | { error?: string };
      if ("error" in parsed && typeof parsed.error === "string") {
        throw new Error(parsed.error);
      }
      const event = parsed as AgentEvent;
      const text = extractText(event);
      if (text === "") continue;

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

      onText(accumulatedText);
    }
  }

  return accumulatedText;
}

function parseAssistantResponse(raw: string): {
  content: string;
  uiActions: UIAction[];
} {
  const uiActions: UIAction[] = [];
  let hasCompleteActionBlock = false;
  const completeBlock = /<ui_actions>\s*([\s\S]*?)\s*<\/ui_actions>/gi;
  for (const match of raw.matchAll(completeBlock)) {
    hasCompleteActionBlock = true;
    try {
      uiActions.push(...normalizeUIActions(JSON.parse(match[1])));
    } catch {
      // Ignore malformed action directives and keep the chat output usable.
    }
  }

  const content = raw
    .replace(/\s*<ui_actions>[\s\S]*?(?:<\/ui_actions>|$)/gi, "")
    .trim();

  return {
    content: content === "" && hasCompleteActionBlock ? "Done." : content,
    uiActions,
  };
}

function normalizeUIActions(value: unknown): UIAction[] {
  const rawActions =
    value != null &&
    typeof value === "object" &&
    "actions" in value &&
    Array.isArray((value as { actions?: unknown }).actions)
      ? (value as { actions: unknown[] }).actions
      : Array.isArray(value)
        ? value
        : [];

  return rawActions.flatMap((rawAction): UIAction[] => {
    if (rawAction == null || typeof rawAction !== "object") return [];
    const action = rawAction as Record<string, unknown>;
    if (action.type === "navigate_recipe") {
      const recipeId = action.recipeId ?? action.recipe_id;
      return typeof recipeId === "string" && recipeId.trim() !== ""
        ? [{ type: "navigate_recipe", recipeId: recipeId.trim() }]
        : [];
    }
    if (action.type === "navigate_recipe_list") {
      return [{ type: "navigate_recipe_list" }];
    }
    if (action.type === "refresh_current_screen") {
      return [{ type: "refresh_current_screen" }];
    }
    return [];
  });
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

export function AgentChat() {
  const [isOpen, setIsOpen] = useState(false);
  const [messages, setMessages] = useState<ChatMessage[]>(initialMessages);
  const [draft, setDraft] = useState("");
  const [isSending, setIsSending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const bottomRef = useRef<HTMLDivElement | null>(null);
  const inputRef = useRef<HTMLTextAreaElement | null>(null);
  const baseURL = useMemo(getAgentBaseURL, []);
  const location = useLocation();
  const navigate = useNavigate();
  const revalidator = useRevalidator();

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

    const userMessage: ChatMessage = {
      id: randomID("user"),
      role: "user",
      content: text,
    };
    const assistantID = randomID("assistant");
    setMessages((current) => [
      ...current,
      userMessage,
      { id: assistantID, role: "assistant", content: "" },
    ]);
    setDraft("");
    setError(null);
    setIsSending(true);

    try {
      const userID = getUserID();
      const sessionID = getSessionID();
      await ensureSession(baseURL, userID, sessionID);

      const res = await fetch(`${baseURL}/run_sse`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          appName: agentAppName,
          userId: userID,
          sessionId: sessionID,
          streaming: true,
          newMessage: {
            role: "user",
            parts: [{ text: buildAgentMessage(text, appContext) }],
          },
        }),
      });

      if (!res.ok) {
        throw new Error(`Agent request failed (${res.status})`);
      }

      const rawResponse = await readAgentStream(res, (chunk) => {
        const parsed = parseAssistantResponse(chunk);
        setMessages((current) =>
          replaceMessageContent(
            current,
            assistantID,
            parsed.content,
            parsed.uiActions,
          ),
        );
      });
      const parsed = parseAssistantResponse(rawResponse);
      setMessages((current) =>
        replaceMessageContent(
          current,
          assistantID,
          parsed.content,
          parsed.uiActions,
        ),
      );
      for (const action of parsed.uiActions) {
        if (action.type === "navigate_recipe") {
          void navigate(`/recipe/${encodeURIComponent(action.recipeId)}`);
        } else if (action.type === "navigate_recipe_list") {
          void navigate("/");
        } else if (action.type === "refresh_current_screen") {
          void revalidator.revalidate();
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Agent request failed.");
      setMessages((current) =>
        current.filter(
          (message) => message.id !== assistantID || message.content !== "",
        ),
      );
    } finally {
      setIsSending(false);
    }
  }

  return (
    <div className="fixed bottom-5 right-5 z-50">
      {isOpen ? (
        <section className="flex h-[min(36rem,calc(100vh-2.5rem))] w-[min(24rem,calc(100vw-2.5rem))] flex-col overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-2xl dark:border-zinc-800 dark:bg-zinc-900">
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
              <div
                key={message.id}
                className={[
                  "flex",
                  message.role === "user" ? "justify-end" : "justify-start",
                ].join(" ")}
              >
                <div
                  className={[
                    "max-w-[85%] rounded-2xl px-3 py-2 text-sm shadow-sm",
                    message.role === "user"
                      ? "bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900"
                      : "border border-zinc-200 bg-white text-zinc-800 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-100",
                  ].join(" ")}
                >
                  {message.content === "" ? (
                    <span className="text-zinc-500 dark:text-zinc-400">
                      Thinking...
                    </span>
                  ) : message.role === "assistant" ? (
                    <div className="space-y-2">
                      <MarkdownMessage content={message.content} />
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
                  ) : (
                    <p className="whitespace-pre-wrap">{message.content}</p>
                  )}
                </div>
              </div>
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
            <div className="flex items-end gap-2">
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
                className="min-h-10 flex-1 resize-none rounded-xl border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition focus:border-zinc-400 focus:ring-2 focus:ring-zinc-200 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-100 dark:focus:border-zinc-500 dark:focus:ring-zinc-800"
              />
              <button
                type="submit"
                disabled={draft.trim() === "" || isSending}
                className="flex size-10 shrink-0 items-center justify-center rounded-full bg-amber-600 text-white shadow-sm transition hover:bg-amber-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-amber-500 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 dark:focus-visible:ring-offset-zinc-900"
                aria-label="Send message"
              >
                <Send className="size-4" aria-hidden />
              </button>
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
    </div>
  );
}
