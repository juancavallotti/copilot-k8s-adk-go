import type { ReactNode } from "react";

import { bootstrapProvider } from "@eetr/react-reducer-utils";

import { traceDetailReducer } from "./reducer";
import { traceDetailInitialState } from "./types";
import type { TraceDetailAction, TraceDetailState } from "./types";

const { Provider, useContextAccessors } = bootstrapProvider<
  TraceDetailState,
  TraceDetailAction
>(traceDetailReducer, traceDetailInitialState);

export function TraceDetailProvider({ children }: { children: ReactNode }) {
  return <Provider>{children}</Provider>;
}

export { useContextAccessors as useTraceDetailState };
