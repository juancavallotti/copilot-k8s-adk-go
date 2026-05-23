package recipes

import (
	"context"

	types "juancavallotti.com/recipe-types"
)

func (s *Service) AddRecipePhoto(ctx context.Context, recipeID string, photo types.Photo) (string, error) {
	if err := ValidateRecipeID(recipeID); err != nil {
		return "", err
	}
	if err := validatePhotos([]types.Photo{photo}); err != nil {
		return "", err
	}
	return s.store.AddRecipePhoto(ctx, recipeID, photo)
}

func (s *Service) DeleteRecipePhoto(ctx context.Context, recipeID string, photoID string) error {
	if err := ValidateRecipeID(recipeID); err != nil {
		return err
	}
	if err := ValidateRecipeID(photoID); err != nil {
		return err
	}
	return s.store.DeleteRecipePhoto(ctx, recipeID, photoID)
}

func (s *Service) SetFeaturedRecipePhoto(ctx context.Context, recipeID string, photoID string) error {
	if err := ValidateRecipeID(recipeID); err != nil {
		return err
	}
	if err := ValidateRecipeID(photoID); err != nil {
		return err
	}
	return s.store.SetFeaturedRecipePhoto(ctx, recipeID, photoID)
}
