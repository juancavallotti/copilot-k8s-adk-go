import {
  TracesListActionType,
  type TracesListAction,
  type TracesListState,
  tracesListInitialState,
} from "./types";

export function tracesListReducer(
  state: TracesListState = tracesListInitialState,
  action: TracesListAction,
): TracesListState {
  switch (action.type) {
    case TracesListActionType.FETCH_STARTED:
      return { events: null, listError: null };
    case TracesListActionType.FETCH_SUCCESS:
      return { events: action.data, listError: null };
    case TracesListActionType.FETCH_FAILED:
      return { events: null, listError: action.data };
    default:
      return state;
  }
}
