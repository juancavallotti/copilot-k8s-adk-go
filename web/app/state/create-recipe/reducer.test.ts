import { describe, expect, it } from "vitest";

import { emptyRecipeDraft } from "~/lib/recipe-draft";

import { createRecipeReducer } from "./reducer";
import { CreateRecipeActionType, type CreateRecipeState } from "./types";

const base: CreateRecipeState = {
  draft: emptyRecipeDraft(),
  submitting: false,
  error: null,
};

describe("createRecipeReducer", () => {
  it("UPDATE_DRAFT replaces draft", () => {
    const draft = { ...emptyRecipeDraft(), name: "Pasta" };
    const next = createRecipeReducer(base, {
      type: CreateRecipeActionType.UPDATE_DRAFT,
      data: draft,
    });
    expect(next.draft.name).toBe("Pasta");
    expect(next.error).toBeNull();
  });

  it("SUBMIT_START clears error and sets submitting", () => {
    const next = createRecipeReducer(
      { ...base, submitting: false, error: "old" },
      { type: CreateRecipeActionType.SUBMIT_START },
    );
    expect(next.submitting).toBe(true);
    expect(next.error).toBeNull();
  });

  it("SUBMIT_ERROR stores message and stops submitting", () => {
    const next = createRecipeReducer(
      { ...base, submitting: true },
      {
        type: CreateRecipeActionType.SUBMIT_ERROR,
        data: "failed",
      },
    );
    expect(next.submitting).toBe(false);
    expect(next.error).toBe("failed");
  });

  it("RESET_FORM applies fresh draft and clears error", () => {
    const next = createRecipeReducer(
      {
        ...base,
        draft: { ...emptyRecipeDraft(), name: "X" },
        error: "e",
      },
      {
        type: CreateRecipeActionType.RESET_FORM,
        data: emptyRecipeDraft(),
      },
    );
    expect(next.draft.name).toBe("");
    expect(next.error).toBeNull();
  });
});
