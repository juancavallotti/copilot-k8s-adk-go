/**
 * Shared trace types (safe for client bundles).
 * All HTTP calls to the traces endpoints live in `traces-http.server.ts` (Node only).
 */
export type Event = {
  event_id: string;
  started_at: string;
  ended_at: string;
  trace_count: number;
};

export type Trace = {
  id: string;
  event_id: string;
  occurred_at: string;
  data: unknown;
};
