import type { Event } from "~/lib/traces-api";

export type TracesListState = {
  events: Event[] | null;
  listError: string | null;
};

export const TracesListActionType = {
  FETCH_STARTED: "FETCH_STARTED",
  FETCH_SUCCESS: "FETCH_SUCCESS",
  FETCH_FAILED: "FETCH_FAILED",
} as const;

export type TracesListAction =
  | { type: typeof TracesListActionType.FETCH_STARTED }
  | { type: typeof TracesListActionType.FETCH_SUCCESS; data: Event[] }
  | { type: typeof TracesListActionType.FETCH_FAILED; data: string };

export const tracesListInitialState: TracesListState = {
  events: null,
  listError: null,
};
