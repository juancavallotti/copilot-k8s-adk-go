export type AppContext = {
  screen: "recipe_list" | "specific_recipe" | "create_recipe" | "other";
  path: string;
  recipeId?: string;
  highlightedText?: string;
};

export function getAppContext(path: string): AppContext {
  if (path === "/") {
    return { screen: "recipe_list", path };
  }
  const recipeMatch = path.match(/^\/recipe\/([^/]+)(?:\/edit)?\/?$/);
  if (recipeMatch != null) {
    return {
      screen: "specific_recipe",
      path,
      recipeId: decodeURIComponent(recipeMatch[1]),
    };
  }
  if (path === "/create") {
    return { screen: "create_recipe", path };
  }
  return { screen: "other", path };
}

export function getHighlightedText(): string | undefined {
  if (typeof window === "undefined") return undefined;
  const text = window.getSelection()?.toString().replace(/\s+/g, " ").trim();
  if (text == null || text === "") return undefined;
  return text.slice(0, 2000);
}

export function buildAgentMessage(
  userMessage: string,
  appContext: AppContext,
): string {
  return JSON.stringify(
    {
      appContext,
      userMessage,
    },
    null,
    2,
  );
}
