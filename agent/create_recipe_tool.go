package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

const (
	generatedRecipePhotoCount = 2
	imageGenerationTimeout    = 45 * time.Second
	maxCreateCLIOutputBytes   = 20 * 1024 * 1024
)

type createRecipeWithGeneratedPhotosArgs struct {
	Name         string   `json:"name" jsonschema:"Recipe name. Required. Must not be blank."`
	Description  string   `json:"description,omitempty" jsonschema:"Optional recipe description."`
	Ingredients  []string `json:"ingredients" jsonschema:"Recipe ingredient lines. Required. Include at least one non-empty ingredient."`
	Instructions []string `json:"instructions" jsonschema:"Recipe instruction lines. Required. Include at least one non-empty instruction."`
	Category     string   `json:"category,omitempty" jsonschema:"Optional recipe category."`
	Image        string   `json:"image,omitempty" jsonschema:"Optional legacy image URL. Usually leave empty when generated photos are created."`
}

type createRecipePayload struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Ingredients  []string               `json:"ingredients"`
	Instructions []string               `json:"instructions"`
	Category     string                 `json:"category,omitempty"`
	Image        string                 `json:"image,omitempty"`
	Photos       []generatedRecipePhoto `json:"photos,omitempty"`
}

type generatedRecipePhoto struct {
	ImageBase64 string `json:"image_base64"`
	Featured    bool   `json:"featured"`
}

type createRecipeWithGeneratedPhotosResult struct {
	Successful      bool                 `json:"successful"`
	RecipeID        string               `json:"recipeId,omitempty"`
	RecipeName      string               `json:"recipeName,omitempty"`
	PhotosRequested int                  `json:"photosRequested"`
	PhotosGenerated int                  `json:"photosGenerated"`
	ImageErrors     []string             `json:"imageErrors,omitempty"`
	CreatedRecipe   *createdRecipeResult `json:"createdRecipe,omitempty"`
	CLI             callRecipesCLIResult `json:"cli"`
}

type createdRecipeResult struct {
	ID     string               `json:"id"`
	Name   string               `json:"name"`
	Photos []createdPhotoResult `json:"photos"`
}

type createdPhotoResult struct {
	ID       string `json:"id,omitempty"`
	Featured bool   `json:"featured"`
}

type recipeCreateCLIRunner func(context.Context, callRecipesCLIArgs) (callRecipesCLIResult, error)

func newCreateRecipeWithGeneratedPhotosTool(generator recipeImageGenerator) (tool.Tool, error) {
	return newCreateRecipeWithGeneratedPhotosToolWithRunner(generator, func(ctx context.Context, input callRecipesCLIArgs) (callRecipesCLIResult, error) {
		return runRecipesCLIWithOutputLimit(ctx, input, maxCreateCLIOutputBytes)
	})
}

func newCreateRecipeWithGeneratedPhotosToolWithRunner(generator recipeImageGenerator, runCLI recipeCreateCLIRunner) (tool.Tool, error) {
	create := func(ctx tool.Context, input createRecipeWithGeneratedPhotosArgs) (createRecipeWithGeneratedPhotosResult, error) {
		return createRecipeWithGeneratedPhotos(ctx, generator, runCLI, input)
	}
	return functiontool.New(functiontool.Config{
		Name:        "create_recipe_with_generated_photos",
		Description: "Creates a recipe with recipes-cli after attempting to generate two Gemini dish photos stored in the recipe photos array.",
	}, create)
}

func createRecipeWithGeneratedPhotos(ctx context.Context, generator recipeImageGenerator, runCLI recipeCreateCLIRunner, input createRecipeWithGeneratedPhotosArgs) (createRecipeWithGeneratedPhotosResult, error) {
	result := createRecipeWithGeneratedPhotosResult{
		PhotosRequested: generatedRecipePhotoCount,
	}

	photos, imageErrors := generateRecipePhotos(ctx, generator, input)
	result.PhotosGenerated = len(photos)
	result.ImageErrors = imageErrors

	payload := createRecipePayload{
		Name:         input.Name,
		Description:  input.Description,
		Ingredients:  input.Ingredients,
		Instructions: input.Instructions,
		Category:     input.Category,
		Image:        input.Image,
		Photos:       photos,
	}
	stdin, err := json.Marshal(payload)
	if err != nil {
		return result, fmt.Errorf("marshal recipe create payload: %w", err)
	}

	cliResult, err := runCLI(ctx, callRecipesCLIArgs{
		Args:           []string{"create", "-"},
		Stdin:          string(stdin),
		TimeoutSeconds: int(maxCLITimeout / time.Second),
	})
	if err != nil {
		return result, err
	}

	result.CLI = sanitizeCreateCLIResult(cliResult)
	result.Successful = cliResult.Successful
	if !cliResult.Successful {
		return result, nil
	}

	created, err := parseCreatedRecipe(cliResult.Stdout)
	if err != nil {
		result.ImageErrors = append(result.ImageErrors, fmt.Sprintf("created recipe but could not parse recipes-cli output: %v", err))
		return result, nil
	}
	result.RecipeID = created.ID
	result.RecipeName = created.Name
	result.CreatedRecipe = created
	result.CLI.Stdout = fmt.Sprintf("Created recipe %q with %d generated photo(s).", created.Name, len(created.Photos))
	return result, nil
}

func generateRecipePhotos(ctx context.Context, generator recipeImageGenerator, input createRecipeWithGeneratedPhotosArgs) ([]generatedRecipePhoto, []string) {
	if generator == nil {
		return nil, []string{"image generator is not configured"}
	}

	prompts := recipeImagePrompts(input)
	photos := make([]generatedRecipePhoto, 0, len(prompts))
	var imageErrors []string
	for i, prompt := range prompts {
		imageCtx, cancel := context.WithTimeout(ctx, imageGenerationTimeout)
		imageBase64, err := generator.GenerateRecipeImage(imageCtx, prompt)
		cancel()
		if err != nil {
			imageErrors = append(imageErrors, fmt.Sprintf("photo %d: %v", i+1, err))
			continue
		}
		photos = append(photos, generatedRecipePhoto{
			ImageBase64: imageBase64,
			Featured:    len(photos) == 0,
		})
	}
	return photos, imageErrors
}

func recipeImagePrompts(input createRecipeWithGeneratedPhotosArgs) []string {
	base := fmt.Sprintf(
		"Create a photorealistic food photograph of the finished dish. Dish name: %s. Description: %s. Category: %s. Key ingredients: %s. Do not include text, captions, labels, watermarks, people, or hands.",
		strings.TrimSpace(input.Name),
		strings.TrimSpace(input.Description),
		strings.TrimSpace(input.Category),
		strings.Join(trimmedNonEmpty(input.Ingredients), ", "),
	)
	return []string{
		base + " Show a close three-quarter angle with natural window light, shallow depth of field, and appetizing texture.",
		base + " Show an overhead presentation idea on a styled table setting with complementary garnishes and serving dishes.",
	}
}

func trimmedNonEmpty(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func parseCreatedRecipe(stdout string) (*createdRecipeResult, error) {
	var created createdRecipeResult
	if err := json.Unmarshal([]byte(stdout), &created); err != nil {
		return nil, err
	}
	if strings.TrimSpace(created.ID) == "" {
		return nil, fmt.Errorf("missing recipe id")
	}
	return &created, nil
}

func sanitizeCreateCLIResult(result callRecipesCLIResult) callRecipesCLIResult {
	if result.Successful {
		result.Stdout = "recipes-cli create succeeded; full stdout omitted because it contains generated image data"
		return result
	}
	if len(result.Stdout) > maxCLIOutputBytes {
		result.Stdout = result.Stdout[:maxCLIOutputBytes] + truncatedOutputNote
	}
	return result
}
