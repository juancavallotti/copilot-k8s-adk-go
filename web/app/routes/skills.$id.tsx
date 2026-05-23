import { useEffect, useState } from "react";
import {
  Link,
  redirect,
  useActionData,
  useLoaderData,
  useNavigation,
  useSubmit,
} from "react-router";

import { MarkdownMessage } from "~/components/agent-chat/markdown-message";
import type {
  Skill,
  SkillCreate,
  SkillPatch,
} from "~/lib/skills-api";

import type { Route } from "./+types/skills.$id";

const NEW_ID = "new";

type LoaderData =
  | { skill: Skill; isNew: false; loadError: null }
  | { skill: null; isNew: true; loadError: null }
  | { skill: null; isNew: false; loadError: string };

type ActionResult =
  | { ok: true }
  | { ok: false; error: string };

export async function loader({
  request,
  params,
}: Route.LoaderArgs): Promise<LoaderData> {
  const id = params.id ?? "";
  if (id === NEW_ID) {
    return { skill: null, isNew: true, loadError: null };
  }
  const { getSkill } = await import("~/lib/skills-http.server");
  try {
    const skill = await getSkill(request, id);
    return { skill, isNew: false, loadError: null };
  } catch (err) {
    return {
      skill: null,
      isNew: false,
      loadError:
        err instanceof Error ? err.message : "Something went wrong",
    };
  }
}

export async function action({
  request,
  params,
}: Route.ActionArgs): Promise<ActionResult | Response> {
  const id = params.id ?? "";
  const { createSkill, updateSkill, deleteSkill } = await import(
    "~/lib/skills-http.server"
  );

  if (request.method === "DELETE") {
    if (id === NEW_ID || id === "") {
      return { ok: false, error: "Cannot delete an unsaved skill." };
    }
    try {
      await deleteSkill(request, id);
      return redirect("/skills");
    } catch (err) {
      return {
        ok: false,
        error: err instanceof Error ? err.message : "Delete failed",
      };
    }
  }

  if (request.method === "POST") {
    let body: SkillCreate;
    try {
      body = (await request.json()) as SkillCreate;
    } catch {
      return { ok: false, error: "Invalid request body." };
    }
    try {
      const created = await createSkill(request, body);
      return redirect(`/skills/${created.id}`);
    } catch (err) {
      return {
        ok: false,
        error: err instanceof Error ? err.message : "Create failed",
      };
    }
  }

  if (request.method === "PATCH") {
    if (id === NEW_ID || id === "") {
      return { ok: false, error: "Missing skill id." };
    }
    let body: SkillPatch;
    try {
      body = (await request.json()) as SkillPatch;
    } catch {
      return { ok: false, error: "Invalid request body." };
    }
    try {
      await updateSkill(request, id, body);
      return { ok: true };
    } catch (err) {
      return {
        ok: false,
        error: err instanceof Error ? err.message : "Save failed",
      };
    }
  }

  return { ok: false, error: `Unsupported method ${request.method}.` };
}

export function meta({ data }: Route.MetaArgs) {
  if (data?.isNew) {
    return [{ title: "New skill · Recipe manager" }];
  }
  if (data?.skill != null) {
    return [{ title: `${data.skill.name} · Skills · Recipe manager` }];
  }
  return [{ title: "Skill · Recipe manager" }];
}

const NAME_RE = /^[a-z0-9][a-z0-9-]*[a-z0-9]$/;

export default function SkillDetail() {
  const loaderData = useLoaderData<typeof loader>();
  const actionData = useActionData<typeof action>();
  const submit = useSubmit();
  const navigation = useNavigation();

  const { skill, isNew, loadError } = loaderData;

  const [name, setName] = useState(skill?.name ?? "");
  const [description, setDescription] = useState(skill?.description ?? "");
  const [content, setContent] = useState(skill?.content ?? "");
  const [localError, setLocalError] = useState<string | null>(null);

  // Re-seed local state when the loaded skill changes (e.g. navigation).
  useEffect(() => {
    setName(skill?.name ?? "");
    setDescription(skill?.description ?? "");
    setContent(skill?.content ?? "");
    setLocalError(null);
  }, [skill?.id, skill?.updated_at]);

  const busy = navigation.state === "submitting";

  if (loadError != null) {
    return (
      <div className="mx-auto max-w-5xl">
        <Link
          to="/skills"
          className="text-sm font-medium text-zinc-600 underline-offset-2 hover:text-zinc-900 hover:underline dark:text-zinc-400 dark:hover:text-zinc-100"
        >
          ← All skills
        </Link>
        <div
          className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-800 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-200"
          role="alert"
        >
          {loadError}
        </div>
      </div>
    );
  }

  function handleSave() {
    setLocalError(null);
    if (isNew) {
      const trimmedName = name.trim();
      if (!NAME_RE.test(trimmedName)) {
        setLocalError(
          "Name must be lowercase letters, digits, and hyphens (e.g. recipe-management).",
        );
        return;
      }
      if (description.trim() === "") {
        setLocalError("Description is required.");
        return;
      }
      if (content.trim() === "") {
        setLocalError("Content is required.");
        return;
      }
      const body: SkillCreate = {
        name: trimmedName,
        description,
        content,
      };
      submit(body, { method: "POST", encType: "application/json" });
      return;
    }
    const body: SkillPatch = { description, content };
    submit(body, { method: "PATCH", encType: "application/json" });
  }

  function handleDelete() {
    if (isNew || skill == null) return;
    if (!window.confirm(`Delete skill "${skill.name}"? This cannot be undone.`)) {
      return;
    }
    submit(null, { method: "DELETE" });
  }

  const submitError =
    actionData != null && "ok" in actionData && actionData.ok === false
      ? actionData.error
      : null;
  const errorMessage = localError ?? submitError;

  const savedToast =
    !isNew &&
    actionData != null &&
    "ok" in actionData &&
    actionData.ok === true;

  return (
    <div className="mx-auto max-w-5xl">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="min-w-0">
          <Link
            to="/skills"
            className="text-sm font-medium text-zinc-600 underline-offset-2 hover:text-zinc-900 hover:underline dark:text-zinc-400 dark:hover:text-zinc-100"
          >
            ← All skills
          </Link>
          <h2 className="mt-2 truncate text-base font-medium text-zinc-900 dark:text-zinc-50">
            {isNew ? "New skill" : skill?.name}
          </h2>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {!isNew ? (
            <button
              type="button"
              onClick={handleDelete}
              disabled={busy}
              className="inline-flex items-center justify-center rounded-lg border border-red-200 bg-white px-3 py-1.5 text-sm font-medium text-red-700 transition-colors hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-red-900/60 dark:bg-zinc-900 dark:text-red-300 dark:hover:bg-red-950/40"
            >
              Delete
            </button>
          ) : null}
          <Link
            to="/skills"
            className="inline-flex items-center justify-center rounded-lg border border-zinc-300 bg-white px-3 py-1.5 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800"
          >
            Cancel
          </Link>
          <button
            type="button"
            onClick={handleSave}
            disabled={busy}
            className="inline-flex items-center justify-center rounded-lg bg-zinc-900 px-3 py-1.5 text-sm font-medium text-white shadow-sm transition-colors hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200"
          >
            {busy ? "Saving…" : "Save"}
          </button>
        </div>
      </div>

      {errorMessage ? (
        <div
          className="mt-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-200"
          role="alert"
        >
          {errorMessage}
        </div>
      ) : null}
      {savedToast ? (
        <div className="mt-4 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-800 dark:border-emerald-900/50 dark:bg-emerald-950/40 dark:text-emerald-200">
          Saved.
        </div>
      ) : null}

      <div className="mt-6 grid gap-4">
        {isNew ? (
          <label className="block">
            <span className="text-xs font-medium uppercase tracking-wide text-zinc-500 dark:text-zinc-400">
              Name (slug)
            </span>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={busy}
              placeholder="recipe-management"
              className="mt-1 block w-full rounded-lg border border-zinc-300 bg-white px-3 py-2 font-mono text-sm shadow-sm focus:border-zinc-500 focus:outline-none focus:ring-1 focus:ring-zinc-500 dark:border-zinc-700 dark:bg-zinc-900 dark:focus:border-zinc-400"
            />
            <span className="mt-1 block text-xs text-zinc-500 dark:text-zinc-400">
              Lowercase letters, digits, and hyphens. Used as the key the agent
              passes to <code>load-skill</code>.
            </span>
          </label>
        ) : null}

        <label className="block">
          <span className="text-xs font-medium uppercase tracking-wide text-zinc-500 dark:text-zinc-400">
            Description
          </span>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            disabled={busy}
            rows={2}
            placeholder="One- or two-sentence summary the agent reads from the catalog to decide whether to load this skill."
            className="mt-1 block w-full rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-zinc-500 focus:outline-none focus:ring-1 focus:ring-zinc-500 dark:border-zinc-700 dark:bg-zinc-900 dark:focus:border-zinc-400"
          />
        </label>
      </div>

      <div className="mt-6">
        <span className="text-xs font-medium uppercase tracking-wide text-zinc-500 dark:text-zinc-400">
          Content (markdown)
        </span>
        <div className="mt-1 grid gap-4 lg:grid-cols-2">
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            disabled={busy}
            spellCheck={false}
            rows={28}
            placeholder="# Skill name&#10;&#10;Detailed instructions the agent loads on demand…"
            className="block min-h-[28rem] w-full rounded-lg border border-zinc-300 bg-white px-3 py-2 font-mono text-xs leading-relaxed shadow-sm focus:border-zinc-500 focus:outline-none focus:ring-1 focus:ring-zinc-500 dark:border-zinc-700 dark:bg-zinc-900 dark:focus:border-zinc-400"
          />
          <div className="overflow-auto rounded-lg border border-zinc-200 bg-white px-4 py-3 text-sm leading-relaxed text-zinc-800 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-200">
            {content.trim() === "" ? (
              <p className="text-xs text-zinc-400 dark:text-zinc-500">
                Markdown preview appears here.
              </p>
            ) : (
              <MarkdownMessage content={content} />
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
