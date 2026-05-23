import { describe, expect, it, vi } from "vitest";

import { createToolContextStore } from "./tool-context";

describe("tool context store", () => {
  it("registers a new tool call and exposes it via get/list", () => {
    const store = createToolContextStore();
    const isNew = store.register({ id: "call-1", name: "search", args: { q: "x" } });
    expect(isNew).toBe(true);
    const call = store.get("call-1");
    expect(call).toMatchObject({ id: "call-1", name: "search", status: "pending" });
    expect(store.list()).toHaveLength(1);
  });

  it("returns false on duplicate register without overwriting state", () => {
    const store = createToolContextStore();
    store.register({ id: "call-1", name: "search" });
    store.applyResponse("call-1", { ok: true });
    const isNewAgain = store.register({ id: "call-1", name: "search" });
    expect(isNewAgain).toBe(false);
    expect(store.get("call-1")?.status).toBe("success");
  });

  it("ignores applyResponse for unknown ids", () => {
    const store = createToolContextStore();
    store.applyResponse("ghost", { ok: true });
    expect(store.get("ghost")).toBeUndefined();
  });

  it("notifies subscribers when their tool changes and only their tool changes", () => {
    const store = createToolContextStore();
    store.register({ id: "a", name: "n" });
    store.register({ id: "b", name: "n" });

    const listenerA = vi.fn();
    const listenerB = vi.fn();
    const unsubA = store.subscribeTool("a", listenerA);
    const unsubB = store.subscribeTool("b", listenerB);

    store.applyResponse("a", { ok: true });
    expect(listenerA).toHaveBeenCalledTimes(1);
    expect(listenerB).not.toHaveBeenCalled();

    store.applyResponse("b", { ok: true });
    expect(listenerA).toHaveBeenCalledTimes(1);
    expect(listenerB).toHaveBeenCalledTimes(1);

    unsubA();
    unsubB();
  });

  it("unsubscribe stops further notifications", () => {
    const store = createToolContextStore();
    store.register({ id: "a", name: "n" });
    const listener = vi.fn();
    const unsub = store.subscribeTool("a", listener);
    store.applyResponse("a", { ok: true });
    expect(listener).toHaveBeenCalledTimes(1);
    unsub();
    store.applyResponse("a", { ok: false, error: "boom" });
    expect(listener).toHaveBeenCalledTimes(1);
  });

  it("reset clears tools and notifies every subscribed listener", () => {
    const store = createToolContextStore();
    store.register({ id: "a", name: "n" });
    store.register({ id: "b", name: "n" });
    const listenerA = vi.fn();
    const listenerB = vi.fn();
    store.subscribeTool("a", listenerA);
    store.subscribeTool("b", listenerB);

    store.reset();
    expect(store.list()).toHaveLength(0);
    expect(store.get("a")).toBeUndefined();
    expect(listenerA).toHaveBeenCalledTimes(1);
    expect(listenerB).toHaveBeenCalledTimes(1);
  });

  it("infers error status from response", () => {
    const store = createToolContextStore();
    store.register({ id: "a", name: "n" });
    store.applyResponse("a", { successful: false });
    expect(store.get("a")?.status).toBe("error");
  });
});
