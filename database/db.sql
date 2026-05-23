-- PostgreSQL schema for recipes (normalized ingredients and steps).
--
-- Idempotent: safe to run repeatedly (e.g. on every container start) without
-- dropping existing data. Uses CREATE IF NOT EXISTS / DROP IF EXISTS patterns.
--
-- Apply manually, for example:
--   createdb recipes
--   psql -v ON_ERROR_STOP=1 -d recipes -f database/db.sql
--
-- Requires PostgreSQL 13+ (uses gen_random_uuid() without an extension).

BEGIN;

CREATE TABLE IF NOT EXISTS recipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT '',
    image TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT recipes_name_nonempty CHECK (length(trim(name)) > 0)
);

CREATE TABLE IF NOT EXISTS ingredients (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    CONSTRAINT ingredients_name_unique UNIQUE (name),
    CONSTRAINT ingredients_name_nonempty CHECK (length(trim(name)) > 0)
);

-- Links a recipe to catalog ingredients with amount per line.
CREATE TABLE IF NOT EXISTS recipes_ingredients (
    id BIGSERIAL PRIMARY KEY,
    recipe_id UUID NOT NULL REFERENCES recipes (id) ON DELETE CASCADE,
    ingredient_id BIGINT NOT NULL REFERENCES ingredients (id) ON DELETE RESTRICT,
    quantity NUMERIC(14, 4) NOT NULL,
    unit TEXT NOT NULL,
    CONSTRAINT recipes_ingredients_unit_nonempty CHECK (length(trim(unit)) > 0),
    CONSTRAINT recipes_ingredients_quantity_nonnegative CHECK (quantity >= 0),
    CONSTRAINT recipes_ingredients_recipe_ingredient_unique UNIQUE (recipe_id, ingredient_id)
);

CREATE TABLE IF NOT EXISTS steps (
    id BIGSERIAL PRIMARY KEY,
    recipe_id UUID NOT NULL REFERENCES recipes (id) ON DELETE CASCADE,
    sort_order INT NOT NULL,
    instruction TEXT NOT NULL,
    CONSTRAINT steps_instruction_nonempty CHECK (length(trim(instruction)) > 0),
    CONSTRAINT steps_sort_positive CHECK (sort_order > 0),
    CONSTRAINT steps_recipe_order_unique UNIQUE (recipe_id, sort_order)
);

CREATE TABLE IF NOT EXISTS recipe_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    image_base64 TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT recipe_images_base64_nonempty CHECK (length(trim(image_base64)) > 0)
);

CREATE TABLE IF NOT EXISTS recipes_images (
    recipe_id UUID NOT NULL REFERENCES recipes (id) ON DELETE CASCADE,
    image_id UUID NOT NULL REFERENCES recipe_images (id) ON DELETE CASCADE,
    is_featured BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (recipe_id, image_id)
);

CREATE INDEX IF NOT EXISTS recipes_ingredients_recipe_id_idx ON recipes_ingredients (recipe_id);
CREATE INDEX IF NOT EXISTS recipes_ingredients_ingredient_id_idx ON recipes_ingredients (ingredient_id);
CREATE INDEX IF NOT EXISTS steps_recipe_id_idx ON steps (recipe_id);
CREATE INDEX IF NOT EXISTS recipes_images_recipe_id_idx ON recipes_images (recipe_id);
CREATE INDEX IF NOT EXISTS recipes_images_image_id_idx ON recipes_images (image_id);
CREATE UNIQUE INDEX IF NOT EXISTS recipes_images_one_featured_idx
    ON recipes_images (recipe_id)
    WHERE is_featured;

CREATE TABLE IF NOT EXISTS events (
    event_id TEXT PRIMARY KEY,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ NOT NULL,
    trace_count INT NOT NULL DEFAULT 0,
    CONSTRAINT events_event_id_nonempty CHECK (length(trim(event_id)) > 0),
    CONSTRAINT events_time_order CHECK (ended_at >= started_at)
);

CREATE INDEX IF NOT EXISTS events_ended_at_idx ON events (ended_at DESC);

CREATE TABLE IF NOT EXISTS traces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id TEXT NOT NULL REFERENCES events (event_id) ON DELETE CASCADE,
    occurred_at TIMESTAMPTZ NOT NULL,
    data JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS traces_event_id_idx ON traces (event_id);
CREATE INDEX IF NOT EXISTS traces_event_id_occurred_at_idx ON traces (event_id, occurred_at);

CREATE TABLE IF NOT EXISTS skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT skills_name_unique UNIQUE (name),
    CONSTRAINT skills_name_slug CHECK (name ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$')
);

CREATE INDEX IF NOT EXISTS skills_name_idx ON skills (name);

INSERT INTO skills (name, description, content) VALUES
    ('recipe-management',
     'Create, patch, delete recipes, manage photos, import/export, and inspect recipe data. Load this skill for any user request involving recipes.',
     $skill$# Recipe Management

This skill covers every recipe-related operation: listing, inspection, creation, patching, deletion, importing, exporting, schema discovery, and photo management.

## Discover the CLI before acting

The embedded help in the system prompt is the source of truth for available commands, but commands and flags are added over time. If a user request might be served by a command you have not explicitly seen, call call_recipes_cli with args ["--help"] to re-check the help text. Do not guess unsupported CLI flags or commands.

## Fetch the JSON Schema before constructing payloads

Before constructing any JSON payload for create or patch — that is, before the first call_recipes_cli invocation that would send recipe JSON over stdin — call call_recipes_cli with args ["schema"] to fetch the current JSON Schema. Use the schema as the source of truth for field names, types, required fields, and nested shapes. Doing this once up front prevents repetitive failed tool calls from guessing the payload shape. You may skip the schema fetch only if you have already fetched it earlier in the same conversation and are confident it has not changed.

## Passing JSON to the CLI

When a command needs JSON input, prefer passing "-" as the CLI path and provide the JSON through the tool''s stdin field. Keep JSON minimal and aligned with the schema output. Report command failures clearly, including stderr when it helps the user recover.

## Generated photo rule

generate_recipe_photos returns filesystem paths, not base64. For every generated photo, use the photo.filePath string as the image-path argument to recipes-cli. Never use "-" or stdin for generated photos. Never copy the handle, path, or filePath value into stdin. Never construct base64 from a generated photo result.

When attaching a generated photo to an existing recipe, call recipes-cli as add-photo <recipe-id> <filePath> [--featured] through call_recipes_cli, where <filePath> is the photo.filePath returned by generate_recipe_photos. Use --featured only when the user asks to feature the photo or when it should replace the current featured image.

## Photo generation UX

Before every generate_recipe_photos tool call, first stream a short user-visible status sentence, for example "I''ll generate the photo now; it may take a little while." Do not call generate_recipe_photos silently. After generate_recipe_photos returns, stream another brief status sentence before attaching files, for example "The photo is ready; I''ll attach it to the recipe now." Keep these progress messages short and do not wait until all tools are done to show the first progress message.

When the user asks you to generate photos for an existing recipe, first stream the photo generation status sentence. If the current appContext does not include enough recipe details for a good image prompt, export or inspect the recipe first. Then use generate_recipe_photos, stream an attachment status sentence, and attach each returned photo with call_recipes_cli add-photo.

## Creation order

When creating a recipe, create the recipe first and generate photos afterward, unless the user explicitly asks for no generated photos. The order is:

1. Call recipes-cli create - through call_recipes_cli with a JSON payload for the recipe without any photos.
2. Stream the photo generation status sentence.
3. Call generate_recipe_photos using the finalized recipe details (title, ingredients, plating, cuisine) as the image prompt so the photo matches the actual recipe.
4. Stream the attachment status sentence.
5. Attach each returned photo with recipes-cli add-photo <recipe-id> <filePath> [--featured].

Do not include generated photos in the create JSON payload. Do not ask the user for images first. If image generation fails, the recipe was already created successfully; explain the photo warning briefly while still treating recipe creation as successful.

## Photo count limit

Never generate more than four photos for a single user request. If the user asks for more than four, generate at most four, explain that four is the maximum per request, and ask whether they want more afterward.

## UI actions for recipe operations

After successfully creating a recipe, call issue_ui_actions and make the final response include a navigate_recipe action with the newly created recipe''s ID, even if generated photos were attached afterward. Prefer navigate_recipe over refresh_current_screen for successful recipe creation.

After any successful change to existing recipe data, call issue_ui_actions and make the final response include refresh_current_screen unless you also need to navigate to the changed recipe. This includes recipe patch/update operations, delete operations, imports, add-photo operations, replacing or featuring photos, and attaching generated photos to an existing recipe.

Generated photo actions refresh the UI only after the generated photo is successfully attached to a recipe or otherwise changes recipe data. If photo generation fails, or if no recipe data changes, explain the result briefly and use an empty actions array unless another UI action is useful.

If you created a new recipe and then attached generated photos to it, use navigate_recipe for the created recipe instead of refresh_current_screen.
$skill$)
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    content     = EXCLUDED.content;

INSERT INTO skills (name, description, content) VALUES
    ('trace-analysis',
     'Investigate agent traces and events. Load this skill when the user asks why the agent did something, wants to inspect a session, or wants to understand recent agent activity.',
     $skill$# Trace Analysis

This skill covers investigating agent runs through stored traces and events. Each agent invocation produces an event row plus one or more trace rows. Events aggregate by invocation_id; traces are the individual structured log entries.

## Listing events

Call recipes-cli list-events to see recent agent invocations, newest first. Each line is one JSON object: event_id, started_at, ended_at, trace_count. Use --limit and --offset for pagination. Limit defaults to 50 and is capped at 200.

When the user asks "what did the agent do recently" or "show me the latest sessions", start here.

## Listing traces for one event

Call recipes-cli list-traces <event-id> to see all trace rows for one event, oldest first. Each line is one JSON object whose data field is the original slog record (msg, level, time, plus any structured attributes the agent or tools attached).

Common trace messages to look for:
- agent.starting / agent.event — top-level agent lifecycle
- tool call_recipes_cli: start / tool call_recipes_cli: done — CLI tool invocations with args, exit code, duration
- Model and image-generation events

## Investigation workflow

1. Ask the user which event or time range to inspect. If they cannot name an event_id, list recent events first and present a short summary.
2. Pull the full trace list for the chosen event.
3. Read the traces top-to-bottom. Pay attention to: which tools were called, in what order, with what args; exit codes; durations; any error or failure messages.
4. Summarize what happened in plain prose. Refer to events and traces by what they did (e.g. "the agent called the CLI''s add-photo command and got exit code 0"), not by raw ids.

## Pagination

Both list-events and list-traces accept --limit and --offset. Use them when the result might exceed 50 rows. If the user asks about activity over a longer time window, paginate explicitly rather than truncating silently.

## UI actions for trace analysis

Trace inspection is read-only, so no refresh action is needed. Use these navigation actions to drop the user on the right page after your reply:

- When the user asks to see recent events, the traces list, or "what the agent did", call issue_ui_actions with navigate_traces_list and include the same action in the <ui_actions> directive at the end of your response.
- When the user asks about a specific event_id or you have a single event in focus and the user wants to inspect it, call issue_ui_actions with navigate_trace and the event_id, and include the same action in the <ui_actions> directive.
- If you are only answering a question about traces without an obvious page to land on, return an empty actions array.
$skill$)
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    content     = EXCLUDED.content;

CREATE OR REPLACE FUNCTION recipes_set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS recipes_updated_at ON recipes;
CREATE TRIGGER recipes_updated_at
    BEFORE UPDATE ON recipes
    FOR EACH ROW
    EXECUTE FUNCTION recipes_set_updated_at();

DROP TRIGGER IF EXISTS skills_updated_at ON skills;
CREATE TRIGGER skills_updated_at
    BEFORE UPDATE ON skills
    FOR EACH ROW
    EXECUTE FUNCTION recipes_set_updated_at();

COMMIT;
