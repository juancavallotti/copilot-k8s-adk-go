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
      return { ...state, events: null, listError: null };
    case TracesListActionType.FETCH_SUCCESS:
      return { ...state, events: action.data, listError: null };
    case TracesListActionType.FETCH_FAILED:
      return { ...state, events: null, listError: action.data };
    case TracesListActionType.DELETE_EVENT_STARTED:
      return {
        ...state,
        deletingEventId: action.data,
        mutationError: null,
      };
    case TracesListActionType.DELETE_EVENT_SUCCEEDED:
      return {
        ...state,
        deletingEventId: null,
        events:
          state.events == null
            ? state.events
            : state.events.filter((e) => e.event_id !== action.data),
      };
    case TracesListActionType.DELETE_EVENT_FAILED:
      return { ...state, deletingEventId: null, mutationError: action.data };
    case TracesListActionType.CLEAR_ALL_STARTED:
      return { ...state, isClearing: true, mutationError: null };
    case TracesListActionType.CLEAR_ALL_SUCCEEDED:
      return { ...state, isClearing: false, events: [] };
    case TracesListActionType.CLEAR_ALL_FAILED:
      return { ...state, isClearing: false, mutationError: action.data };
    case TracesListActionType.MUTATION_DISMISS:
      return { ...state, mutationError: null };
    default:
      return state;
  }
}
