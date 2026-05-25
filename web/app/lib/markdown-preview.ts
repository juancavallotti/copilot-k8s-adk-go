const DEFAULT_MAX_CHARS = 200;

/**
 * Prepare a markdown string for an inline preview (e.g. a list row): strip
 * block-level constructs that don't make sense in a one-or-two-line snippet
 * (tables, lists, headings, blockquotes) and truncate to the first ~2
 * sentences or `maxChars` characters, whichever comes first.
 */
export function toMarkdownPreview(
  source: string,
  maxChars: number = DEFAULT_MAX_CHARS,
): string {
  const stripped = stripBlockMarkdown(source);
  return truncate(stripped, maxChars);
}

function stripBlockMarkdown(source: string): string {
  const lines = source.split(/\r?\n/);
  const kept: string[] = [];
  for (const raw of lines) {
    const line = raw.trim();
    if (line === "") continue;
    if (isTableLine(line)) continue;
    if (isHeading(line)) {
      kept.push(line.replace(/^#+\s*/, ""));
      continue;
    }
    if (isBlockquote(line)) {
      kept.push(line.replace(/^>\s?/, ""));
      continue;
    }
    const withoutListMarker = line
      .replace(/^[-*+]\s+/, "")
      .replace(/^\d+\.\s+/, "");
    kept.push(withoutListMarker);
  }
  return kept.join(" ").replace(/\s+/g, " ").trim();
}

function isTableLine(line: string): boolean {
  if (!line.startsWith("|")) return false;
  return line.endsWith("|") || /^\|[\s:|-]+$/.test(line);
}

function isHeading(line: string): boolean {
  return /^#{1,6}\s+/.test(line);
}

function isBlockquote(line: string): boolean {
  return line.startsWith(">");
}

function truncate(text: string, maxChars: number): string {
  if (text.length === 0) return text;

  const sentenceEnd = findSecondSentenceEnd(text);
  if (sentenceEnd !== -1 && sentenceEnd <= maxChars) {
    if (sentenceEnd >= text.length) return text;
    return appendEllipsis(text.slice(0, sentenceEnd));
  }

  if (text.length <= maxChars) return text;

  const lastSpace = text.lastIndexOf(" ", maxChars);
  const cutoff = lastSpace > maxChars * 0.6 ? lastSpace : maxChars;
  return appendEllipsis(text.slice(0, cutoff));
}

function appendEllipsis(text: string): string {
  return text.trimEnd().replace(/[.!?,;:]+$/, "") + "…";
}

function findSecondSentenceEnd(text: string): number {
  const regex = /[.!?](?=\s|$)/g;
  let count = 0;
  let match: RegExpExecArray | null;
  while ((match = regex.exec(text)) !== null) {
    count++;
    if (count === 2) return match.index + 1;
  }
  return -1;
}
