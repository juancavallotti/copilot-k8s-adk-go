import type { Recipe } from "~/lib/recipe-api";

/** Editable recipe fields (no id or timestamps). Shared by create and edit flows. */
export type RecipeDraft = {
  name: string;
  description: string;
  category: string;
  image: string;
  ingredients: string[];
  instructions: string[];
};

export function emptyRecipeDraft(): RecipeDraft {
  return {
    name: "",
    description: "",
    category: "",
    image: "",
    ingredients: [""],
    instructions: [""],
  };
}

/** Hydrate the editor from an API recipe (e.g. edit screen). */
export function recipeToDraft(recipe: Recipe): RecipeDraft {
  return {
    name: recipe.name,
    description: recipe.description,
    category: recipe.category,
    image: recipe.image,
    ingredients:
      recipe.ingredients.length > 0 ? [...recipe.ingredients] : [""],
    instructions:
      recipe.instructions.length > 0 ? [...recipe.instructions] : [""],
  };
}

export function draftToCreateBody(draft: RecipeDraft) {
  return {
    name: draft.name.trim(),
    description: draft.description.trim(),
    category: draft.category.trim(),
    image: draft.image.trim(),
    ingredients: draft.ingredients.map((s) => s.trim()).filter(Boolean),
    instructions: draft.instructions.map((s) => s.trim()).filter(Boolean),
  };
}
