import { Plus, Trash2 } from "lucide-react";
import { useId } from "react";

import { inputClass } from "~/lib/input-styles";
import type { RecipeDraft } from "~/lib/recipe-draft";

export type RecipeEditorProps = {
  value: RecipeDraft;
  onChange: (next: RecipeDraft) => void;
  disabled?: boolean;
};

function labelClass() {
  return "text-xs font-medium uppercase tracking-wide text-zinc-500 dark:text-zinc-400";
}

function updateAt<T>(arr: T[], index: number, item: T): T[] {
  const next = [...arr];
  next[index] = item;
  return next;
}

function StringListEditor({
  label,
  hint,
  values,
  onChange,
  disabled,
  addLabel,
  placeholder,
  multiline,
}: {
  label: string;
  hint?: string;
  values: string[];
  onChange: (next: string[]) => void;
  disabled: boolean;
  addLabel: string;
  placeholder: string;
  multiline: boolean;
}) {
  const baseId = useId();

  function setLine(i: number, v: string) {
    onChange(updateAt(values, i, v));
  }

  function addLine() {
    onChange([...values, ""]);
  }

  function removeLine(i: number) {
    if (values.length <= 1) {
      onChange(updateAt(values, i, ""));
      return;
    }
    onChange(values.filter((_, j) => j !== i));
  }

  return (
    <div className="space-y-2">
      <div>
        <p className={labelClass()}>{label}</p>
        {hint ? (
          <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">{hint}</p>
        ) : null}
      </div>
      <ul className="space-y-2">
        {values.map((line, i) => {
          const fieldId = `${baseId}-${i}`;
          return (
            <li key={fieldId} className="flex gap-2">
              {multiline ? (
                <textarea
                  id={fieldId}
                  name={`${label.toLowerCase()}-${i}`}
                  rows={3}
                  className={`${inputClass(disabled)} min-h-[5rem] resize-y`}
                  value={line}
                  onChange={(e) => setLine(i, e.target.value)}
                  disabled={disabled}
                  placeholder={placeholder}
                />
              ) : (
                <input
                  id={fieldId}
                  name={`${label.toLowerCase()}-${i}`}
                  type="text"
                  className={inputClass(disabled)}
                  value={line}
                  onChange={(e) => setLine(i, e.target.value)}
                  disabled={disabled}
                  placeholder={placeholder}
                />
              )}
              <button
                type="button"
                className="inline-flex size-10 shrink-0 items-center justify-center rounded-lg border border-zinc-200 text-zinc-500 transition-colors hover:border-red-200 hover:bg-red-50 hover:text-red-700 disabled:pointer-events-none disabled:opacity-40 dark:border-zinc-700 dark:hover:border-red-900/60 dark:hover:bg-red-950/40 dark:hover:text-red-300"
                onClick={() => removeLine(i)}
                disabled={disabled}
                aria-label={`Remove ${label} row ${i + 1}`}
              >
                <Trash2 className="size-4 stroke-[2]" aria-hidden />
              </button>
            </li>
          );
        })}
      </ul>
      <button
        type="button"
        className="inline-flex items-center gap-1.5 rounded-lg border border-dashed border-zinc-300 px-3 py-2 text-xs font-medium text-zinc-600 transition-colors hover:border-zinc-400 hover:bg-zinc-50 disabled:pointer-events-none disabled:opacity-40 dark:border-zinc-600 dark:text-zinc-300 dark:hover:border-zinc-500 dark:hover:bg-zinc-800/80"
        onClick={addLine}
        disabled={disabled}
      >
        <Plus className="size-3.5 stroke-[2]" aria-hidden />
        {addLabel}
      </button>
    </div>
  );
}

export function RecipeEditor({ value, onChange, disabled = false }: RecipeEditorProps) {
  const id = useId();

  return (
    <div className="space-y-8">
      <div className="grid gap-6 sm:grid-cols-2">
        <div className="space-y-2 sm:col-span-2">
          <label htmlFor={`${id}-name`} className={labelClass()}>
            Name
          </label>
          <input
            id={`${id}-name`}
            name="name"
            type="text"
            required
            className={inputClass(disabled)}
            value={value.name}
            onChange={(e) => onChange({ ...value, name: e.target.value })}
            disabled={disabled}
            placeholder="e.g. Weeknight lentil soup"
            autoComplete="off"
          />
        </div>
        <div className="space-y-2">
          <label htmlFor={`${id}-category`} className={labelClass()}>
            Category
          </label>
          <input
            id={`${id}-category`}
            name="category"
            type="text"
            className={inputClass(disabled)}
            value={value.category}
            onChange={(e) => onChange({ ...value, category: e.target.value })}
            disabled={disabled}
            placeholder="e.g. Dinner"
            autoComplete="off"
          />
        </div>
        <div className="space-y-2">
          <label htmlFor={`${id}-image`} className={labelClass()}>
            Image URL
          </label>
          <input
            id={`${id}-image`}
            name="image"
            type="url"
            className={inputClass(disabled)}
            value={value.image}
            onChange={(e) => onChange({ ...value, image: e.target.value })}
            disabled={disabled}
            placeholder="https://…"
            autoComplete="off"
          />
        </div>
      </div>

      <div className="space-y-2">
        <label htmlFor={`${id}-description`} className={labelClass()}>
          Description
        </label>
        <textarea
          id={`${id}-description`}
          name="description"
          rows={4}
          className={`${inputClass(disabled)} min-h-[6rem] resize-y`}
          value={value.description}
          onChange={(e) => onChange({ ...value, description: e.target.value })}
          disabled={disabled}
          placeholder="Short summary of the dish"
        />
      </div>

      <StringListEditor
        label="Ingredients"
        hint='One per line. Quantity and unit are parsed, e.g. "2 cups flour", "200 g sugar", or a bare name like "salt".'
        values={value.ingredients}
        onChange={(ingredients) => onChange({ ...value, ingredients })}
        disabled={disabled}
        addLabel="Add ingredient"
        placeholder="2 cups all-purpose flour"
        multiline={false}
      />

      <StringListEditor
        label="Instructions"
        hint="Markdown is supported in each step."
        values={value.instructions}
        onChange={(instructions) => onChange({ ...value, instructions })}
        disabled={disabled}
        addLabel="Add step"
        placeholder="Describe this step…"
        multiline
      />
    </div>
  );
}
