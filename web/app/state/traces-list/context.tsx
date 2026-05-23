import type { ReactNode } from "react";

import { bootstrapProvider } from "@eetr/react-reducer-utils";

import { tracesListReducer } from "./reducer";
import { tracesListInitialState } from "./types";
import type { TracesListAction, TracesListState } from "./types";

const { Provider, useContextAccessors } = bootstrapProvider<
  TracesListState,
  TracesListAction
>(tracesListReducer, tracesListInitialState);

export function TracesListProvider({ children }: { children: ReactNode }) {
  return <Provider>{children}</Provider>;
}

export { useContextAccessors as useTracesListState };
