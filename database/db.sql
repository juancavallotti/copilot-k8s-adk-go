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

COMMIT;
