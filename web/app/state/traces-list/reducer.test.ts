import { describe, expect, it } from "vitest";

import type { Event } from "~/lib/traces-api";

import { tracesListReducer } from "./reducer";
import {
  TracesListActionType,
  tracesListInitialState,
} from "./types";

function event(overrides: Partial<Event> = {}): Event {
  return {
    event_id: "evt-1",
    started_at: "2025-01-01T00:00:00Z",
    ended_at: "2025-01-01T00:00:05Z",
    trace_count: 3,
    ...overrides,
  };
}

describe("tracesListReducer", () => {
  it("FETCH_STARTED clears events and error", () => {
    const next = tracesListReducer(
      { ...tracesListInitialState, events: [event()], listError: "boom" },
      { type: TracesListActionType.FETCH_STARTED },
    );
    expect(next.events).toBeNull();
    expect(next.listError).toBeNull();
  });

  it("FETCH_SUCCESS stores events and clears error", () => {
    const next = tracesListReducer(
      { ...tracesListInitialState, listError: "old" },
      {
        type: TracesListActionType.FETCH_SUCCESS,
        data: [event({ event_id: "a" }), event({ event_id: "b" })],
      },
    );
    expect(next.events).toHaveLength(2);
    expect(next.listError).toBeNull();
  });

  it("FETCH_FAILED sets message and clears events", () => {
    const next = tracesListReducer(
      { ...tracesListInitialState, events: [event()] },
      { type: TracesListActionType.FETCH_FAILED, data: "network" },
    );
    expect(next.events).toBeNull();
    expect(next.listError).toBe("network");
  });
});
