package recipes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
	types "juancavallotti.com/recipe-types"
)

func (s *Store) GetRecipes(ctx context.Context) ([]types.Recipe, error) {
	if s.db == nil {
		return nil, errNilDB
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id::text, name, description, category, image, created_at, updated_at
FROM recipes
ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]types.Recipe, 0)
	ids := make([]string, 0)
	for rows.Next() {
		var r types.Recipe
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Category, &r.Image, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		r.Ingredients = []string{}
		r.Instructions = []string{}
		r.Photos = []types.Photo{}
		out = append(out, r)
		ids = append(ids, r.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	photosByRecipeID, err := s.loadPhotosForRecipes(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range out {
		if photos, ok := photosByRecipeID[out[i].ID]; ok {
			out[i].Photos = photos
		}
	}
	return out, nil
}

func (s *Store) GetRecipe(ctx context.Context, id string) (types.Recipe, error) {
	if s.db == nil {
		return types.Recipe{}, errNilDB
	}
	if err := ctx.Err(); err != nil {
		return types.Recipe{}, err
	}
	id = strings.TrimSpace(id)
	if _, err := uuid.Parse(id); err != nil {
		return types.Recipe{}, ErrInvalidID
	}

	var r types.Recipe
	err := s.db.QueryRowContext(ctx, `
SELECT id::text, name, description, category, image, created_at, updated_at
FROM recipes WHERE id = $1::uuid`, id).Scan(
		&r.ID, &r.Name, &r.Description, &r.Category, &r.Image, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return types.Recipe{}, ErrRecipeNotFound
	}
	if err != nil {
		return types.Recipe{}, err
	}

	ingRows, err := s.db.QueryContext(ctx, `
SELECT ri.quantity::text, ri.unit, i.name
FROM recipes_ingredients ri
JOIN ingredients i ON i.id = ri.ingredient_id
WHERE ri.recipe_id = $1::uuid
ORDER BY ri.id`, id)
	if err != nil {
		return types.Recipe{}, err
	}
	defer ingRows.Close()

	r.Ingredients = []string{}
	for ingRows.Next() {
		var qtyStr, unit, name string
		if err := ingRows.Scan(&qtyStr, &unit, &name); err != nil {
			return types.Recipe{}, err
		}
		q := new(big.Rat)
		if _, ok := q.SetString(qtyStr); !ok {
			return types.Recipe{}, fmt.Errorf("invalid quantity in database: %q", qtyStr)
		}
		line := FormatIngredientLine(ParsedIngredient{Quantity: q, Unit: unit, Name: name})
		r.Ingredients = append(r.Ingredients, line)
	}
	if err := ingRows.Err(); err != nil {
		return types.Recipe{}, err
	}

	stepRows, err := s.db.QueryContext(ctx, `
SELECT instruction FROM steps WHERE recipe_id = $1::uuid ORDER BY sort_order`, id)
	if err != nil {
		return types.Recipe{}, err
	}
	defer stepRows.Close()

	r.Instructions = []string{}
	for stepRows.Next() {
		var inst string
		if err := stepRows.Scan(&inst); err != nil {
			return types.Recipe{}, err
		}
		r.Instructions = append(r.Instructions, inst)
	}
	if err := stepRows.Err(); err != nil {
		return types.Recipe{}, err
	}

	r.Photos, err = s.loadRecipePhotos(ctx, id)
	if err != nil {
		return types.Recipe{}, err
	}

	return r, nil
}
