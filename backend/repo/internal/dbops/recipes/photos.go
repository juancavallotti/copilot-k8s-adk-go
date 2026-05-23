package recipes

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
	types "juancavallotti.com/recipe-types"
)

func (s *Store) AddRecipePhoto(ctx context.Context, recipeID string, photo types.Photo) (string, error) {
	if s.db == nil {
		return "", errNilDB
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	recipeID = strings.TrimSpace(recipeID)
	if _, err := uuid.Parse(recipeID); err != nil {
		return "", ErrInvalidID
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM recipes WHERE id = $1::uuid)`, recipeID).Scan(&exists); err != nil {
		return "", err
	}
	if !exists {
		return "", ErrRecipeNotFound
	}

	id, err := s.insertPhoto(ctx, tx, recipeID, photo)
	if err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE recipes SET updated_at = now() WHERE id = $1::uuid`, recipeID); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Store) DeleteRecipePhoto(ctx context.Context, recipeID string, photoID string) error {
	if s.db == nil {
		return errNilDB
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	recipeID = strings.TrimSpace(recipeID)
	photoID = strings.TrimSpace(photoID)
	if _, err := uuid.Parse(recipeID); err != nil {
		return ErrInvalidID
	}
	if _, err := uuid.Parse(photoID); err != nil {
		return ErrInvalidID
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
DELETE FROM recipes_images
WHERE recipe_id = $1::uuid AND image_id = $2::uuid`, recipeID, photoID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrPhotoNotFound
	}
	if _, err := tx.ExecContext(ctx, `
DELETE FROM recipe_images img
WHERE img.id = $1::uuid
  AND NOT EXISTS (SELECT 1 FROM recipes_images ri WHERE ri.image_id = img.id)`, photoID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE recipes SET updated_at = now() WHERE id = $1::uuid`, recipeID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) SetFeaturedRecipePhoto(ctx context.Context, recipeID string, photoID string) error {
	if s.db == nil {
		return errNilDB
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	recipeID = strings.TrimSpace(recipeID)
	photoID = strings.TrimSpace(photoID)
	if _, err := uuid.Parse(recipeID); err != nil {
		return ErrInvalidID
	}
	if _, err := uuid.Parse(photoID); err != nil {
		return ErrInvalidID
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
UPDATE recipes_images
SET is_featured = false
WHERE recipe_id = $1::uuid`, recipeID); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `
UPDATE recipes_images
SET is_featured = true
WHERE recipe_id = $1::uuid AND image_id = $2::uuid`, recipeID, photoID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrPhotoNotFound
	}
	if _, err := tx.ExecContext(ctx, `UPDATE recipes SET updated_at = now() WHERE id = $1::uuid`, recipeID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) insertPhoto(ctx context.Context, tx *sql.Tx, recipeID string, photo types.Photo) (string, error) {
	if photo.Featured {
		if _, err := tx.ExecContext(ctx, `UPDATE recipes_images SET is_featured = false WHERE recipe_id = $1::uuid`, recipeID); err != nil {
			return "", err
		}
	}

	photoID := strings.TrimSpace(photo.ID)
	if photoID == "" {
		if err := tx.QueryRowContext(ctx, `
INSERT INTO recipe_images (image_base64)
VALUES ($1)
RETURNING id::text`, photo.ImageBase64).Scan(&photoID); err != nil {
			return "", err
		}
	} else {
		if _, err := uuid.Parse(photoID); err != nil {
			return "", ErrInvalidID
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO recipe_images (id, image_base64)
VALUES ($1::uuid, $2)
ON CONFLICT (id) DO UPDATE SET image_base64 = EXCLUDED.image_base64`, photoID, photo.ImageBase64); err != nil {
			return "", err
		}
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO recipes_images (recipe_id, image_id, is_featured)
VALUES ($1::uuid, $2::uuid, $3)
ON CONFLICT (recipe_id, image_id) DO UPDATE SET is_featured = EXCLUDED.is_featured`,
		recipeID, photoID, photo.Featured); err != nil {
		return "", err
	}
	return photoID, nil
}

func (s *Store) replaceRecipePhotos(ctx context.Context, tx *sql.Tx, recipeID string, photos []types.Photo) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM recipes_images WHERE recipe_id = $1::uuid`, recipeID); err != nil {
		return err
	}
	for _, photo := range photos {
		if _, err := s.insertPhoto(ctx, tx, recipeID, photo); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) loadRecipePhotos(ctx context.Context, recipeID string) ([]types.Photo, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT img.id::text, img.image_base64, ri.is_featured
FROM recipes_images ri
JOIN recipe_images img ON img.id = ri.image_id
WHERE ri.recipe_id = $1::uuid
ORDER BY ri.is_featured DESC, ri.created_at, img.id`, recipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	photos := []types.Photo{}
	for rows.Next() {
		var photo types.Photo
		if err := rows.Scan(&photo.ID, &photo.ImageBase64, &photo.Featured); err != nil {
			return nil, err
		}
		photos = append(photos, photo)
	}
	return photos, rows.Err()
}

func (s *Store) loadPhotosForRecipes(ctx context.Context, recipeIDs []string) (map[string][]types.Photo, error) {
	photosByRecipeID := make(map[string][]types.Photo, len(recipeIDs))
	if len(recipeIDs) == 0 {
		return photosByRecipeID, nil
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT ri.recipe_id::text, img.id::text, img.image_base64, ri.is_featured
FROM recipes_images ri
JOIN recipe_images img ON img.id = ri.image_id
WHERE ri.recipe_id = ANY($1::uuid[])
ORDER BY ri.recipe_id, ri.is_featured DESC, ri.created_at, img.id`, pq.Array(recipeIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var recipeID string
		var photo types.Photo
		if err := rows.Scan(&recipeID, &photo.ID, &photo.ImageBase64, &photo.Featured); err != nil {
			return nil, err
		}
		photosByRecipeID[recipeID] = append(photosByRecipeID[recipeID], photo)
	}
	return photosByRecipeID, rows.Err()
}
