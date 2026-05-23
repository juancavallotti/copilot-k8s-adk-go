import type { Trace } from "~/lib/traces-api";

export type TraceDetailState = {
  traces: Trace[] | null;
  error: string | null;
};

export const TraceDetailActionType = {
  LOAD_RESET: "LOAD_RESET",
  MISSING_ID: "MISSING_ID",
  LOAD_SUCCESS: "LOAD_SUCCESS",
  LOAD_FAILED: "LOAD_FAILED",
} as const;

export type TraceDetailAction =
  | { type: typeof TraceDetailActionType.LOAD_RESET }
  | { type: typeof TraceDetailActionType.MISSING_ID; data: string }
  | { type: typeof TraceDetailActionType.LOAD_SUCCESS; data: Trace[] }
  | { type: typeof TraceDetailActionType.LOAD_FAILED; data: string };

export const traceDetailInitialState: TraceDetailState = {
  traces: null,
  error: null,
};
