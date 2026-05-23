import type { Event } from "~/lib/traces-api";

export type TracesListState = {
  events: Event[] | null;
  listError: string | null;
  deletingEventId: string | null;
  isClearing: boolean;
  mutationError: string | null;
};

export const TracesListActionType = {
  FETCH_STARTED: "FETCH_STARTED",
  FETCH_SUCCESS: "FETCH_SUCCESS",
  FETCH_FAILED: "FETCH_FAILED",
  DELETE_EVENT_STARTED: "DELETE_EVENT_STARTED",
  DELETE_EVENT_SUCCEEDED: "DELETE_EVENT_SUCCEEDED",
  DELETE_EVENT_FAILED: "DELETE_EVENT_FAILED",
  CLEAR_ALL_STARTED: "CLEAR_ALL_STARTED",
  CLEAR_ALL_SUCCEEDED: "CLEAR_ALL_SUCCEEDED",
  CLEAR_ALL_FAILED: "CLEAR_ALL_FAILED",
  MUTATION_DISMISS: "MUTATION_DISMISS",
} as const;

export type TracesListAction =
  | { type: typeof TracesListActionType.FETCH_STARTED }
  | { type: typeof TracesListActionType.FETCH_SUCCESS; data: Event[] }
  | { type: typeof TracesListActionType.FETCH_FAILED; data: string }
  | { type: typeof TracesListActionType.DELETE_EVENT_STARTED; data: string }
  | { type: typeof TracesListActionType.DELETE_EVENT_SUCCEEDED; data: string }
  | { type: typeof TracesListActionType.DELETE_EVENT_FAILED; data: string }
  | { type: typeof TracesListActionType.CLEAR_ALL_STARTED }
  | { type: typeof TracesListActionType.CLEAR_ALL_SUCCEEDED }
  | { type: typeof TracesListActionType.CLEAR_ALL_FAILED; data: string }
  | { type: typeof TracesListActionType.MUTATION_DISMISS };

export const tracesListInitialState: TracesListState = {
  events: null,
  listError: null,
  deletingEventId: null,
  isClearing: false,
  mutationError: null,
};
