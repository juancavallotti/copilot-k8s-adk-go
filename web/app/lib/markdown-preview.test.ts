import { describe, expect, it } from "vitest";

import { toMarkdownPreview } from "~/lib/markdown-preview";

describe("toMarkdownPreview", () => {
  it("strips markdown table rows", () => {
    const input = [
      "Quick weeknight pasta.",
      "",
      "| Ingredient | Amount |",
      "| --- | --- |",
      "| Pasta | 200g |",
      "| Salt | 1 tsp |",
    ].join("\n");

    expect(toMarkdownPreview(input)).toBe("Quick weeknight pasta.");
  });

  it("flattens bulleted lists into inline text without markers", () => {
    const input = [
      "Tips for success:",
      "- Use cold butter",
      "- Don't overmix",
      "- Rest the dough",
    ].join("\n");

    expect(toMarkdownPreview(input)).toBe(
      "Tips for success: Use cold butter Don't overmix Rest the dough",
    );
  });

  it("flattens numbered lists", () => {
    const input = ["Steps:", "1. Mix", "2. Bake", "3. Cool"].join("\n");
    expect(toMarkdownPreview(input)).toBe("Steps: Mix Bake Cool");
  });

  it("strips headings while keeping their text", () => {
    const input = ["# Big title", "Body text follows."].join("\n");
    expect(toMarkdownPreview(input)).toBe("Big title Body text follows.");
  });

  it("strips blockquote markers", () => {
    const input = ["> Pro tip:", "Always preheat."].join("\n");
    expect(toMarkdownPreview(input)).toBe("Pro tip: Always preheat.");
  });

  it("truncates to the first two sentences when they fit within the limit", () => {
    const input =
      "First sentence here. Second sentence here. Third sentence here.";
    expect(toMarkdownPreview(input)).toBe(
      "First sentence here. Second sentence here…",
    );
  });

  it("returns the original text when it has at most two sentences", () => {
    const input = "Only one sentence here.";
    expect(toMarkdownPreview(input)).toBe(input);
  });

  it("truncates at a word boundary with ellipsis when over the char limit", () => {
    const input = "a".repeat(150) + " " + "b".repeat(150);
    const result = toMarkdownPreview(input, 200);
    expect(result.endsWith("…")).toBe(true);
    expect(result.length).toBeLessThanOrEqual(201);
  });

  it("returns an empty string for empty input", () => {
    expect(toMarkdownPreview("")).toBe("");
  });

  it("preserves inline markdown like bold and links", () => {
    const input = "Use **fresh** basil and [olive oil](https://example.com).";
    expect(toMarkdownPreview(input)).toBe(
      "Use **fresh** basil and [olive oil](https://example.com).",
    );
  });
});
