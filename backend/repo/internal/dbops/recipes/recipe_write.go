package recipes

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	types "juancavallotti.com/recipe-types"
)

func (s *Store) upsertIngredientID(ctx context.Context, tx *sql.Tx, name string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("%w: empty ingredient name", ErrParseIngredient)
	}
	const q = `
INSERT INTO ingredients (name) VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = ingredients.name
RETURNING id`
	var id int64
	if err := tx.QueryRowContext(ctx, q, name).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Store) insertIngredientsAndSteps(ctx context.Context, tx *sql.Tx, recipeID string, recipe types.Recipe) error {
	parsed := make([]ParsedIngredient, 0, len(recipe.Ingredients))
	for _, line := range recipe.Ingredients {
		p, err := ParseIngredientLine(line)
		if err != nil {
			return err
		}
		parsed = append(parsed, p)
	}
	merged, err := MergeParsedIngredients(parsed)
	if err != nil {
		return err
	}
	for _, p := range merged {
		ingID, err := s.upsertIngredientID(ctx, tx, p.Name)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
INSERT INTO recipes_ingredients (recipe_id, ingredient_id, quantity, unit)
VALUES ($1::uuid, $2, $3::numeric, $4)`,
			recipeID, ingID, ratToNumericString(p.Quantity), p.Unit)
		if err != nil {
			return err
		}
	}
	for i, inst := range recipe.Instructions {
		inst = strings.TrimSpace(inst)
		if inst == "" {
			return fmt.Errorf("%w: empty instruction at position %d", ErrParseIngredient, i+1)
		}
		_, err := tx.ExecContext(ctx, `
INSERT INTO steps (recipe_id, sort_order, instruction)
VALUES ($1::uuid, $2, $3)`,
			recipeID, i+1, inst)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) replaceIngredientsAndSteps(ctx context.Context, tx *sql.Tx, recipeID string, recipe types.Recipe) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM recipes_ingredients WHERE recipe_id = $1::uuid`, recipeID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM steps WHERE recipe_id = $1::uuid`, recipeID); err != nil {
		return err
	}
	return s.insertIngredientsAndSteps(ctx, tx, recipeID, recipe)
}
