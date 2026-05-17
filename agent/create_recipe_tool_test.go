package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type fakeRecipeImageGenerator struct {
	images []string
	errs   []error
	calls  int
}

func (f *fakeRecipeImageGenerator) GenerateRecipeImage(context.Context, string) (string, error) {
	call := f.calls
	f.calls++
	if call < len(f.errs) && f.errs[call] != nil {
		return "", f.errs[call]
	}
	if call < len(f.images) {
		return f.images[call], nil
	}
	return "image", nil
}

func TestGenerateRecipePhotosContinuesAfterImageFailure(t *testing.T) {
	generator := &fakeRecipeImageGenerator{
		images: []string{"", "second-image"},
		errs:   []error{errors.New("blocked"), nil},
	}

	photos, imageErrors := generateRecipePhotos(context.Background(), generator, createRecipeWithGeneratedPhotosArgs{
		Name:        "Soup",
		Ingredients: []string{"tomatoes"},
	})

	if len(photos) != 1 {
		t.Fatalf("len(photos) = %d, want 1", len(photos))
	}
	if !photos[0].Featured {
		t.Fatal("first successful generated photo should be featured")
	}
	if photos[0].ImageBase64 != "second-image" {
		t.Fatalf("photo image = %q, want second-image", photos[0].ImageBase64)
	}
	if len(imageErrors) != 1 || !strings.Contains(imageErrors[0], "blocked") {
		t.Fatalf("imageErrors = %#v, want blocked error", imageErrors)
	}
}

func TestCreateRecipeWithGeneratedPhotosSendsPhotosToCLI(t *testing.T) {
	generator := &fakeRecipeImageGenerator{
		images: []string{"first-image", "second-image"},
	}
	var gotPayload createRecipePayload
	runCLI := func(_ context.Context, input callRecipesCLIArgs) (callRecipesCLIResult, error) {
		if got, want := strings.Join(input.Args, " "), "create -"; got != want {
			t.Fatalf("args = %q, want %q", got, want)
		}
		if err := json.Unmarshal([]byte(input.Stdin), &gotPayload); err != nil {
			t.Fatalf("unmarshal stdin: %v", err)
		}
		return callRecipesCLIResult{
			Command:    "recipes-cli create -",
			ExitCode:   0,
			Stdout:     `{"id":"recipe-1","name":"Soup","photos":[{"id":"photo-1","featured":true},{"id":"photo-2","featured":false}]}`,
			Successful: true,
		}, nil
	}

	result, err := createRecipeWithGeneratedPhotos(context.Background(), generator, runCLI, createRecipeWithGeneratedPhotosArgs{
		Name:         "Soup",
		Ingredients:  []string{"tomatoes"},
		Instructions: []string{"Simmer."},
	})
	if err != nil {
		t.Fatalf("createRecipeWithGeneratedPhotos() error = %v", err)
	}
	if !result.Successful {
		t.Fatal("result.Successful = false, want true")
	}
	if result.RecipeID != "recipe-1" {
		t.Fatalf("RecipeID = %q, want recipe-1", result.RecipeID)
	}
	if result.PhotosGenerated != 2 {
		t.Fatalf("PhotosGenerated = %d, want 2", result.PhotosGenerated)
	}
	if len(gotPayload.Photos) != 2 {
		t.Fatalf("len(payload.Photos) = %d, want 2", len(gotPayload.Photos))
	}
	if !gotPayload.Photos[0].Featured || gotPayload.Photos[1].Featured {
		t.Fatalf("payload photo featured flags = %#v, want only first featured", gotPayload.Photos)
	}
	if gotPayload.Photos[0].ImageBase64 != "first-image" || gotPayload.Photos[1].ImageBase64 != "second-image" {
		t.Fatalf("payload photos = %#v, want generated images", gotPayload.Photos)
	}
	if strings.Contains(result.CLI.Stdout, "image_base64") {
		t.Fatalf("CLI stdout should be sanitized, got %q", result.CLI.Stdout)
	}
}

func TestRecipeImagePromptsAskForDistinctPresentations(t *testing.T) {
	prompts := recipeImagePrompts(createRecipeWithGeneratedPhotosArgs{
		Name:        "Pasta",
		Description: "Bright weeknight dinner",
		Category:    "Dinner",
		Ingredients: []string{" pasta ", "", " basil "},
	})

	if len(prompts) != 2 {
		t.Fatalf("len(prompts) = %d, want 2", len(prompts))
	}
	if !strings.Contains(prompts[0], "three-quarter angle") {
		t.Fatalf("first prompt = %q, want angle guidance", prompts[0])
	}
	if !strings.Contains(prompts[1], "overhead presentation") {
		t.Fatalf("second prompt = %q, want presentation guidance", prompts[1])
	}
	for _, prompt := range prompts {
		if !strings.Contains(prompt, "Pasta") || !strings.Contains(prompt, "pasta, basil") {
			t.Fatalf("prompt = %q, want recipe details", prompt)
		}
	}
}
