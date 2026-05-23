import { describe, expect, it } from "vitest";

import type { Trace } from "~/lib/traces-api";

import { traceDetailReducer } from "./reducer";
import {
  TraceDetailActionType,
  traceDetailInitialState,
} from "./types";

function sampleTrace(overrides: Partial<Trace> = {}): Trace {
  return {
    id: "t-1",
    event_id: "evt-1",
    occurred_at: "2025-01-01T00:00:00Z",
    data: { hello: "world" },
    ...overrides,
  };
}

describe("traceDetailReducer", () => {
  it("LOAD_RESET clears traces and error", () => {
    const next = traceDetailReducer(
      { traces: [sampleTrace()], error: "x" },
      { type: TraceDetailActionType.LOAD_RESET },
    );
    expect(next.traces).toBeNull();
    expect(next.error).toBeNull();
  });

  it("MISSING_ID sets error only", () => {
    const next = traceDetailReducer(traceDetailInitialState, {
      type: TraceDetailActionType.MISSING_ID,
      data: "Missing event id.",
    });
    expect(next.error).toBe("Missing event id.");
    expect(next.traces).toBeNull();
  });

  it("LOAD_SUCCESS sets traces and clears error", () => {
    const traces = [sampleTrace(), sampleTrace({ id: "t-2" })];
    const next = traceDetailReducer(
      { traces: null, error: "old" },
      { type: TraceDetailActionType.LOAD_SUCCESS, data: traces },
    );
    expect(next.traces).toEqual(traces);
    expect(next.error).toBeNull();
  });

  it("LOAD_FAILED sets error and clears traces", () => {
    const next = traceDetailReducer(
      { traces: [sampleTrace()], error: null },
      { type: TraceDetailActionType.LOAD_FAILED, data: "Not found" },
    );
    expect(next.traces).toBeNull();
    expect(next.error).toBe("Not found");
  });
});
