package recipes

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	types "juancavallotti.com/recipe-types"
)

var (
	// ErrInvalidRecipeID is returned when a recipe id is missing or whitespace-only.
	ErrInvalidRecipeID = errors.New("invalid recipe id")
	// ErrInvalidRecipe is returned when recipe fields fail validation.
	ErrInvalidRecipe = errors.New("invalid recipe")
)

// ValidateRecipeID checks that id is non-empty after trimming spaces.
func ValidateRecipeID(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w", ErrInvalidRecipeID)
	}
	return nil
}

// ValidateRecipeForCreate checks fields required when inserting a recipe.
// ID must be empty; name, at least one ingredient, and at least one instruction are required.
func ValidateRecipeForCreate(recipe types.Recipe) error {
	if strings.TrimSpace(recipe.ID) != "" {
		return fmt.Errorf("%w: id must be empty for create", ErrInvalidRecipe)
	}
	return validateRecipeContent(recipe)
}

// ValidateRecipeForUpdate checks fields required when updating a recipe.
func ValidateRecipeForUpdate(recipe types.Recipe) error {
	if err := ValidateRecipeID(recipe.ID); err != nil {
		return err
	}
	return validateRecipeContent(recipe)
}

func validateRecipeContent(recipe types.Recipe) error {
	if strings.TrimSpace(recipe.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidRecipe)
	}
	if !hasNonEmptyEntry(recipe.Ingredients) {
		return fmt.Errorf("%w: at least one non-empty ingredient is required", ErrInvalidRecipe)
	}
	if !hasNonEmptyEntry(recipe.Instructions) {
		return fmt.Errorf("%w: at least one non-empty instruction is required", ErrInvalidRecipe)
	}
	if err := validatePhotos(recipe.Photos); err != nil {
		return err
	}
	return nil
}

func validatePhotos(photos []types.Photo) error {
	featured := 0
	for i, photo := range photos {
		if strings.TrimSpace(photo.ImageBase64) == "" {
			return fmt.Errorf("%w: photo %d image_base64 is required", ErrInvalidRecipe, i+1)
		}
		if _, err := base64.StdEncoding.DecodeString(photo.ImageBase64); err != nil {
			return fmt.Errorf("%w: photo %d image_base64 is not valid base64", ErrInvalidRecipe, i+1)
		}
		if photo.Featured {
			featured++
		}
	}
	if featured > 1 {
		return fmt.Errorf("%w: at most one featured photo is allowed", ErrInvalidRecipe)
	}
	return nil
}

func hasNonEmptyEntry(items []string) bool {
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			return true
		}
	}
	return false
}
