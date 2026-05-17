package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

const (
	defaultGeneratedRecipePhotoCount = 1
	maxGeneratedRecipePhotoCount     = 4
	imageGenerationTimeout           = 45 * time.Second
)

type generateRecipePhotosArgs struct {
	Name        string   `json:"name" jsonschema:"Recipe or dish name. Required. Must not be blank."`
	Description string   `json:"description,omitempty" jsonschema:"Optional recipe description to guide photo generation."`
	Ingredients []string `json:"ingredients,omitempty" jsonschema:"Optional recipe ingredient lines to guide photo generation."`
	Category    string   `json:"category,omitempty" jsonschema:"Optional recipe category to guide photo generation."`
	UserRequest string   `json:"userRequest,omitempty" jsonschema:"Optional user guidance for the photo, such as angle, plating, mood, or featured image intent."`
	Count       int      `json:"count,omitempty" jsonschema:"Number of photos to generate. Defaults to 1. Maximum is 4."`
}

type generatedRecipePhoto struct {
	ImageBase64 string `json:"image_base64"`
	Featured    bool   `json:"featured"`
}

type generateRecipePhotosResult struct {
	PhotosRequested int                    `json:"photosRequested"`
	PhotosGenerated int                    `json:"photosGenerated"`
	Photos          []generatedRecipePhoto `json:"photos,omitempty"`
	ImageErrors     []string               `json:"imageErrors,omitempty"`
	Capped          bool                   `json:"capped,omitempty"`
}

func newGenerateRecipePhotosTool(generator recipeImageGenerator, concurrency int) (tool.Tool, error) {
	concurrency = normalizedImageGenerationConcurrency(concurrency)
	generate := func(ctx tool.Context, input generateRecipePhotosArgs) (generateRecipePhotosResult, error) {
		return generateRecipePhotos(ctx, generator, input, concurrency)
	}
	return functiontool.New(functiontool.Config{
		Name:          "generate_recipe_photos",
		Description:   "Generates up to four Gemini dish photos and returns raw base64 image data that can be passed to recipes-cli create or add-photo. This can take up to 45 seconds per photo.",
		IsLongRunning: true,
	}, generate)
}

func generateRecipePhotos(ctx context.Context, generator recipeImageGenerator, input generateRecipePhotosArgs, concurrency int) (generateRecipePhotosResult, error) {
	result := generateRecipePhotosResult{
		PhotosRequested: normalizedRecipePhotoCount(input.Count),
		Capped:          input.Count > maxGeneratedRecipePhotoCount,
	}
	if strings.TrimSpace(input.Name) == "" {
		return result, fmt.Errorf("name is required")
	}
	if generator == nil {
		result.ImageErrors = []string{"image generator is not configured"}
		return result, nil
	}

	prompts := recipeImagePrompts(input)
	generated := generateRecipePhotosParallel(ctx, generator, prompts, normalizedImageGenerationConcurrency(concurrency))
	photos := make([]generatedRecipePhoto, 0, len(prompts))
	var imageErrors []string
	for i, item := range generated {
		if item.err != nil {
			imageErrors = append(imageErrors, fmt.Sprintf("photo %d: %v", i+1, item.err))
			continue
		}
		photos = append(photos, generatedRecipePhoto{
			ImageBase64: item.imageBase64,
			Featured:    len(photos) == 0,
		})
	}
	result.Photos = photos
	result.PhotosGenerated = len(photos)
	result.ImageErrors = imageErrors
	return result, nil
}

type generatedRecipePhotoAttempt struct {
	imageBase64 string
	err         error
}

func generateRecipePhotosParallel(ctx context.Context, generator recipeImageGenerator, prompts []string, concurrency int) []generatedRecipePhotoAttempt {
	concurrency = normalizedImageGenerationConcurrency(concurrency)
	results := make([]generatedRecipePhotoAttempt, len(prompts))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for i, prompt := range prompts {
		i := i
		prompt := prompt
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			imageCtx, cancel := context.WithTimeout(ctx, imageGenerationTimeout)
			defer cancel()
			imageBase64, err := generator.GenerateRecipeImage(imageCtx, prompt)
			results[i] = generatedRecipePhotoAttempt{
				imageBase64: imageBase64,
				err:         err,
			}
		}()
	}
	wg.Wait()
	return results
}

func normalizedImageGenerationConcurrency(concurrency int) int {
	if concurrency < 1 {
		return defaultImageGenerationConcurrency
	}
	if concurrency > maxGeneratedRecipePhotoCount {
		return maxGeneratedRecipePhotoCount
	}
	return concurrency
}

func normalizedRecipePhotoCount(count int) int {
	if count <= 0 {
		return defaultGeneratedRecipePhotoCount
	}
	if count > maxGeneratedRecipePhotoCount {
		return maxGeneratedRecipePhotoCount
	}
	return count
}

func recipeImagePrompts(input generateRecipePhotosArgs) []string {
	base := fmt.Sprintf(
		"Create a photorealistic food photograph of the finished dish. Dish name: %s. Description: %s. Category: %s. Key ingredients: %s. Do not include text, captions, labels, watermarks, people, or hands.",
		strings.TrimSpace(input.Name),
		strings.TrimSpace(input.Description),
		strings.TrimSpace(input.Category),
		strings.Join(trimmedNonEmpty(input.Ingredients), ", "),
	)
	if request := strings.TrimSpace(input.UserRequest); request != "" {
		base += " User requested photo guidance: " + request + "."
	}
	variations := []string{
		base + " Show a close three-quarter angle with natural window light, shallow depth of field, and appetizing texture.",
		base + " Show an overhead presentation idea on a styled table setting with complementary garnishes and serving dishes.",
		base + " Show a clean hero shot with restaurant-style plating and a softly blurred background.",
		base + " Show a cozy serving scene with realistic tableware, garnishes, and warm natural color.",
	}
	return variations[:normalizedRecipePhotoCount(input.Count)]
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
