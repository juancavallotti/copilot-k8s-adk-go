package commands

import (
	"context"
	"encoding/base64"
	"os"
	"strings"

	types "juancavallotti.com/recipe-types"
)

func (r Runner) cmdAddPhoto(ctx context.Context, repo RecipeRepo, recipeID string, path string, featured bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if _, err := repo.AddRecipePhoto(ctx, strings.TrimSpace(recipeID), types.Photo{
		ImageBase64: base64.StdEncoding.EncodeToString(data),
		Featured:    featured,
	}); err != nil {
		return err
	}
	updated, err := repo.GetRecipe(ctx, strings.TrimSpace(recipeID))
	if err != nil {
		return err
	}
	return r.writeIndentedJSON(updated)
}

func (r Runner) cmdDeletePhoto(ctx context.Context, repo RecipeRepo, recipeID string, photoID string) error {
	recipeID = strings.TrimSpace(recipeID)
	if err := repo.DeleteRecipePhoto(ctx, recipeID, strings.TrimSpace(photoID)); err != nil {
		return err
	}
	updated, err := repo.GetRecipe(ctx, recipeID)
	if err != nil {
		return err
	}
	return r.writeIndentedJSON(updated)
}

func (r Runner) cmdSetFeaturedPhoto(ctx context.Context, repo RecipeRepo, recipeID string, photoID string) error {
	recipeID = strings.TrimSpace(recipeID)
	if err := repo.SetFeaturedRecipePhoto(ctx, recipeID, strings.TrimSpace(photoID)); err != nil {
		return err
	}
	updated, err := repo.GetRecipe(ctx, recipeID)
	if err != nil {
		return err
	}
	return r.writeIndentedJSON(updated)
}
