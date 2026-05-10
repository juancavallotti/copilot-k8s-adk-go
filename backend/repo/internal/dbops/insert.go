package dbops

import (
	"context"

	types "juancavallotti.com/recipe-types"
)

func (s *Store) CreateRecipe(ctx context.Context, recipe types.Recipe) (string, error) {
	if s.db == nil {
		return "", errNilDB
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	var id string
	err = tx.QueryRowContext(ctx, `
INSERT INTO recipes (name, description, category, image)
VALUES ($1, $2, $3, $4)
RETURNING id::text`,
		recipe.Name, recipe.Description, recipe.Category, recipe.Image).Scan(&id)
	if err != nil {
		return "", err
	}
	if err := s.insertIngredientsAndSteps(ctx, tx, id, recipe); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return id, nil
}
