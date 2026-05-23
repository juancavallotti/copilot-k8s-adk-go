import {
  createContext,
  useContext,
  useMemo,
  useSyncExternalStore,
  type ReactNode,
} from "react";

import {
  applyToolResponse,
  makePendingToolCall,
  type ToolCall,
} from "./agent-ui-actions";

export type ToolContextStore = {
  get(id: string): ToolCall | undefined;
  list(): ToolCall[];
  /** Returns true when the id was newly registered, false if already present. */
  register(call: { id: string; name: string; args?: Record<string, unknown> }): boolean;
  applyResponse(id: string, response: unknown): void;
  reset(): void;
  subscribeTool(id: string, listener: () => void): () => void;
};

export function createToolContextStore(): ToolContextStore {
  const tools = new Map<string, ToolCall>();
  const listeners = new Map<string, Set<() => void>>();

  function notify(id: string) {
    const set = listeners.get(id);
    if (set == null) return;
    for (const listener of set) listener();
  }

  return {
    get(id) {
      return tools.get(id);
    },
    list() {
      return Array.from(tools.values());
    },
    register(call) {
      if (tools.has(call.id)) return false;
      tools.set(call.id, makePendingToolCall(call));
      notify(call.id);
      return true;
    },
    applyResponse(id, response) {
      const current = tools.get(id);
      if (current == null) return;
      tools.set(id, applyToolResponse(current, response));
      notify(id);
    },
    reset() {
      const ids = Array.from(tools.keys());
      tools.clear();
      for (const id of ids) notify(id);
    },
    subscribeTool(id, listener) {
      let set = listeners.get(id);
      if (set == null) {
        set = new Set();
        listeners.set(id, set);
      }
      set.add(listener);
      return () => {
        const current = listeners.get(id);
        if (current == null) return;
        current.delete(listener);
        if (current.size === 0) listeners.delete(id);
      };
    },
  };
}

const ToolContextContext = createContext<ToolContextStore | null>(null);

export function ToolContextProvider({
  store,
  children,
}: {
  store: ToolContextStore;
  children: ReactNode;
}) {
  return (
    <ToolContextContext.Provider value={store}>
      {children}
    </ToolContextContext.Provider>
  );
}

export function useToolContext(): ToolContextStore {
  const ctx = useContext(ToolContextContext);
  if (ctx == null) {
    throw new Error("useToolContext must be used within a ToolContextProvider");
  }
  return ctx;
}

export function useToolCall(id: string): ToolCall | undefined {
  const store = useToolContext();
  const subscribe = useMemo(
    () => (listener: () => void) => store.subscribeTool(id, listener),
    [store, id],
  );
  const getSnapshot = useMemo(() => () => store.get(id), [store, id]);
  return useSyncExternalStore(subscribe, getSnapshot, getSnapshot);
}
