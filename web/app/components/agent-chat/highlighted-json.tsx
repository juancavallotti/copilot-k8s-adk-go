import type { ReactNode } from "react";

export function HighlightedJSON({ source }: { source: string }) {
  const tokens: ReactNode[] = [];
  const regex =
    /("(?:\\.|[^"\\])*")(\s*:)?|\b(true|false|null)\b|(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)/g;
  let lastIndex = 0;
  let key = 0;
  let match: RegExpExecArray | null;
  while ((match = regex.exec(source)) !== null) {
    if (match.index > lastIndex) {
      tokens.push(source.slice(lastIndex, match.index));
    }
    if (match[1] != null) {
      const isKey = match[2] != null;
      tokens.push(
        <span
          key={key++}
          className={isKey ? "text-amber-300" : "text-emerald-300"}
        >
          {match[1]}
        </span>,
      );
      if (isKey) tokens.push(match[2]);
    } else if (match[3] != null) {
      tokens.push(
        <span
          key={key++}
          className={match[3] === "null" ? "text-zinc-500" : "text-rose-300"}
        >
          {match[3]}
        </span>,
      );
    } else if (match[4] != null) {
      tokens.push(
        <span key={key++} className="text-sky-300">
          {match[4]}
        </span>,
      );
    }
    lastIndex = regex.lastIndex;
  }
  if (lastIndex < source.length) {
    tokens.push(source.slice(lastIndex));
  }
  return <>{tokens}</>;
}
