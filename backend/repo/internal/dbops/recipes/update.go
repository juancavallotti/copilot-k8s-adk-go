package recipes

import (
	"context"
	"strings"

	"github.com/google/uuid"
	types "juancavallotti.com/recipe-types"
)

func (s *Store) UpdateRecipe(ctx context.Context, recipe types.Recipe) error {
	if s.db == nil {
		return errNilDB
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	id := strings.TrimSpace(recipe.ID)
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
UPDATE recipes
SET name = $2, description = $3, category = $4, image = $5
WHERE id = $1::uuid`,
		id, recipe.Name, recipe.Description, recipe.Category, recipe.Image)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrRecipeNotFound
	}
	if err := s.replaceIngredientsAndSteps(ctx, tx, id, recipe); err != nil {
		return err
	}
	if err := s.replaceRecipePhotos(ctx, tx, id, recipe.Photos); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
