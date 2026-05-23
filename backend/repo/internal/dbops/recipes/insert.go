package recipes

import (
	"context"
	"strings"

	"github.com/google/uuid"
	types "juancavallotti.com/recipe-types"
)

func (s *Store) CreateRecipe(ctx context.Context, recipe types.Recipe) (string, error) {
	if s.db == nil {
		return "", errNilDB
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return s.insertRecipe(ctx, recipe, "")
}

// CreateRecipeWithID inserts a recipe using recipe.ID as the primary key. The id must be a valid UUID.
func (s *Store) CreateRecipeWithID(ctx context.Context, recipe types.Recipe) error {
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
	_, err := s.insertRecipe(ctx, recipe, id)
	return err
}

// insertRecipe inserts the recipe row and linked ingredients/steps in one transaction.
// If explicitID is empty, the database assigns a new UUID; otherwise explicitID must be a valid UUID string.
func (s *Store) insertRecipe(ctx context.Context, recipe types.Recipe, explicitID string) (id string, err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	if explicitID == "" {
		err = tx.QueryRowContext(ctx, `
INSERT INTO recipes (name, description, category, image)
VALUES ($1, $2, $3, $4)
RETURNING id::text`,
			recipe.Name, recipe.Description, recipe.Category, recipe.Image).Scan(&id)
	} else {
		_, err = tx.ExecContext(ctx, `
INSERT INTO recipes (id, name, description, category, image)
VALUES ($1::uuid, $2, $3, $4, $5)`,
			explicitID, recipe.Name, recipe.Description, recipe.Category, recipe.Image)
		id = explicitID
	}
	if err != nil {
		return "", err
	}
	if err := s.insertIngredientsAndSteps(ctx, tx, id, recipe); err != nil {
		return "", err
	}
	if err := s.replaceRecipePhotos(ctx, tx, id, recipe.Photos); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return id, nil
}
