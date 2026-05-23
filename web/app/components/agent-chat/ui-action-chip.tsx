import type { UIAction } from "~/lib/agent-ui-actions";

export function describeUIAction(action: UIAction): string {
  switch (action.type) {
    case "navigate_recipe":
      return "Open recipe";
    case "navigate_recipe_list":
      return "Open recipe list";
    case "navigate_trace":
      return "Open trace";
    case "navigate_traces_list":
      return "Open traces list";
    case "refresh_current_screen":
      return "Refresh current screen";
  }
}

export function UIActionChips({ actions }: { actions: UIAction[] }) {
  if (actions.length === 0) return null;
  return (
    <div className="flex flex-wrap gap-1.5 border-t border-zinc-100 pt-2 dark:border-zinc-800">
      {actions.map((action, index) => (
        <span
          key={`${action.type}-${index}`}
          className="rounded-full bg-amber-50 px-2 py-0.5 text-[0.6875rem] font-medium text-amber-800 dark:bg-amber-950/70 dark:text-amber-200"
        >
          {describeUIAction(action)}
        </span>
      ))}
    </div>
  );
}
