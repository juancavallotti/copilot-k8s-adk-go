import {
  extractToolEventsFromEvent,
  extractUIActionsFromEvent,
  uniqueUIActions,
  type UIAction,
} from "./agent-ui-actions";

export type AgentEvent = {
  author?: string;
  partial?: boolean;
  turnComplete?: boolean;
  errorMessage?: string;
  content?: {
    parts?: Array<{
      text?: string;
      functionCall?: unknown;
      function_call?: unknown;
      functionResponse?: unknown;
      function_response?: unknown;
    }>;
  };
};

export type ToolEventBatch = ReturnType<typeof extractToolEventsFromEvent>;

export type BubbleStreamHandlers = {
  /** Tokens streaming into the current bubble. Fires on every partial-text event. */
  onTextProgress: (text: string, uiActions: UIAction[]) => void;
  /** First turnComplete for a bubble. Commit the streamed text and fire side effects (navigate, refresh). */
  onTurnFinalize: (text: string, uiActions: UIAction[], rawEvents: string[]) => void;
  /** Post-turn re-emission of the same content. Silently replaces the previous bubble's text/uiActions (side effects already fired in onTurnFinalize). */
  onTurnAuthoritativeText: (text: string, uiActions: UIAction[], rawEvents: string[]) => void;
  /** Tool calls / responses. Routed straight through; orthogonal to turn-text state. */
  onToolEvents: (events: ToolEventBatch) => void;
};

function extractText(event: AgentEvent): string {
  return (
    event.content?.parts
      ?.map((part) => part.text ?? "")
      .filter((text) => text !== "")
      .join("") ?? ""
  );
}

function isReemitMatch(candidate: string, finalized: string): boolean {
  if (candidate === "" || finalized === "") return false;
  return candidate === finalized
    || candidate.startsWith(finalized)
    || finalized.startsWith(candidate);
}

export async function readAgentStream(
  response: Response,
  handlers: BubbleStreamHandlers,
): Promise<void> {
  if (response.body == null) {
    throw new Error("Agent response did not include a stream.");
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  // Current (in-progress) bubble accumulators.
  let accumulatedText = "";
  let accumulatedUIActions: UIAction[] = [];
  let bubbleRawEvents: string[] = [];
  let bubbleHasContent = false;

  // Post-finalize state — held until we either confirm a re-emit or detect a new turn.
  let awaitingAuthoritative = false;
  let finalizedTurnText = "";
  let finalizedTurnUIActions: UIAction[] = [];
  let authoritativeBuffer = "";
  let authoritativeUIActions: UIAction[] = [];
  let authoritativeRawEvents: string[] = [];

  function resetBubble() {
    accumulatedText = "";
    accumulatedUIActions = [];
    bubbleRawEvents = [];
    bubbleHasContent = false;
  }

  function clearAuthoritative() {
    awaitingAuthoritative = false;
    finalizedTurnText = "";
    finalizedTurnUIActions = [];
    authoritativeBuffer = "";
    authoritativeUIActions = [];
    authoritativeRawEvents = [];
  }

  function emitAuthoritative() {
    const text =
      authoritativeBuffer.length >= finalizedTurnText.length
        ? authoritativeBuffer
        : finalizedTurnText;
    const merged = uniqueUIActions([
      ...finalizedTurnUIActions,
      ...authoritativeUIActions,
    ]);
    const events = authoritativeRawEvents;
    handlers.onTurnAuthoritativeText(text, merged, events);
    clearAuthoritative();
  }

  function processRawEvent(rawEvent: string) {
    const data = rawEvent
      .split("\n")
      .filter((line) => line.startsWith("data:"))
      .map((line) => line.slice(5).trimStart())
      .join("\n");
    if (data === "") return;

    const parsed = JSON.parse(data) as AgentEvent | { error?: string };
    if ("error" in parsed && typeof parsed.error === "string") {
      throw new Error(parsed.error);
    }
    const event = parsed as AgentEvent;

    const toolEvents = extractToolEventsFromEvent(event);
    if (toolEvents.calls.length > 0 || toolEvents.responses.length > 0) {
      handlers.onToolEvents(toolEvents);
    }

    const text = extractText(event);
    const eventUIActions = extractUIActionsFromEvent(event);

    // ── Post-turn window: is this event part of the re-emit, or the start of a new turn? ──
    if (awaitingAuthoritative) {
      const couldBeReemit = text === "" || isReemitMatch(text, finalizedTurnText);
      if (couldBeReemit) {
        authoritativeRawEvents.push(data);
        if (text !== "" && text.length > authoritativeBuffer.length) {
          authoritativeBuffer = text;
        }
        if (eventUIActions.length > 0) {
          authoritativeUIActions = uniqueUIActions([
            ...authoritativeUIActions,
            ...eventUIActions,
          ]);
        }
        if (event.turnComplete === true) {
          emitAuthoritative();
        }
        return;
      }
      // Divergent text → no re-emit coming. Flush any buffered raw events as
      // authoritative (with the finalized text we have) before falling through.
      if (authoritativeRawEvents.length > 0 || authoritativeBuffer !== "") {
        emitAuthoritative();
      } else {
        clearAuthoritative();
      }
      // fall through to handle as new-turn event
    }

    bubbleRawEvents.push(data);

    if (eventUIActions.length > 0) {
      accumulatedUIActions = uniqueUIActions([
        ...accumulatedUIActions,
        ...eventUIActions,
      ]);
      bubbleHasContent = true;
    }

    if (text !== "") {
      if (event.partial === true) {
        accumulatedText = text.startsWith(accumulatedText)
          ? text
          : `${accumulatedText}${text}`;
      } else if (accumulatedText === "" || text.startsWith(accumulatedText)) {
        accumulatedText = text;
      } else if (!accumulatedText.endsWith(text)) {
        accumulatedText = `${accumulatedText}${text}`;
      }
      bubbleHasContent = true;
    }

    if (text !== "" || eventUIActions.length > 0) {
      handlers.onTextProgress(accumulatedText, accumulatedUIActions);
    }

    if (event.turnComplete === true && bubbleHasContent) {
      handlers.onTurnFinalize(accumulatedText, accumulatedUIActions, bubbleRawEvents);
      finalizedTurnText = accumulatedText;
      finalizedTurnUIActions = accumulatedUIActions;
      awaitingAuthoritative = true;
      authoritativeBuffer = "";
      authoritativeUIActions = [];
      resetBubble();
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

  // Stream ended.
  if (awaitingAuthoritative) {
    if (authoritativeBuffer !== "") {
      emitAuthoritative();
    } else {
      clearAuthoritative();
    }
  }
  if (bubbleHasContent || bubbleRawEvents.length > 0) {
    handlers.onTurnFinalize(accumulatedText, accumulatedUIActions, bubbleRawEvents);
    resetBubble();
  }
}
