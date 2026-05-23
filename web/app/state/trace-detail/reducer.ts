import {
  TraceDetailActionType,
  type TraceDetailAction,
  type TraceDetailState,
  traceDetailInitialState,
} from "./types";

export function traceDetailReducer(
  state: TraceDetailState = traceDetailInitialState,
  action: TraceDetailAction,
): TraceDetailState {
  switch (action.type) {
    case TraceDetailActionType.LOAD_RESET:
      return { traces: null, error: null };
    case TraceDetailActionType.MISSING_ID:
      return { traces: null, error: action.data };
    case TraceDetailActionType.LOAD_SUCCESS:
      return { traces: action.data, error: null };
    case TraceDetailActionType.LOAD_FAILED:
      return { traces: null, error: action.data };
    default:
      return state;
  }
}
