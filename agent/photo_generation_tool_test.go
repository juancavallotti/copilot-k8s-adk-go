package main

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeRecipeImageGenerator struct {
	mu        sync.Mutex
	images    []string
	errs      []error
	calls     int
	delay     time.Duration
	active    int
	maxActive int
}

func (f *fakeRecipeImageGenerator) GenerateRecipeImage(context.Context, string) (string, error) {
	f.mu.Lock()
	call := f.calls
	f.calls++
	f.active++
	if f.active > f.maxActive {
		f.maxActive = f.active
	}
	f.mu.Unlock()
	defer func() {
		f.mu.Lock()
		f.active--
		f.mu.Unlock()
	}()
	if f.delay > 0 {
		time.Sleep(f.delay)
	}

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

	result, err := generateRecipePhotos(context.Background(), generator, generateRecipePhotosArgs{
		Name:        "Soup",
		Ingredients: []string{"tomatoes"},
		Count:       2,
	}, 1)
	if err != nil {
		t.Fatalf("generateRecipePhotos() error = %v", err)
	}

	if len(result.Photos) != 1 {
		t.Fatalf("len(photos) = %d, want 1", len(result.Photos))
	}
	if !result.Photos[0].Featured {
		t.Fatal("first successful generated photo should be featured")
	}
	if result.Photos[0].ImageBase64 != "second-image" {
		t.Fatalf("photo image = %q, want second-image", result.Photos[0].ImageBase64)
	}
	if len(result.ImageErrors) != 1 || !strings.Contains(result.ImageErrors[0], "blocked") {
		t.Fatalf("imageErrors = %#v, want blocked error", result.ImageErrors)
	}
}

func TestGenerateRecipePhotosCapsAtFour(t *testing.T) {
	generator := &fakeRecipeImageGenerator{
		images: []string{"first-image", "second-image", "third-image", "fourth-image", "fifth-image"},
	}

	result, err := generateRecipePhotos(context.Background(), generator, generateRecipePhotosArgs{
		Name:  "Soup",
		Count: 5,
	}, 4)
	if err != nil {
		t.Fatalf("generateRecipePhotos() error = %v", err)
	}
	if !result.Capped {
		t.Fatal("Capped = false, want true")
	}
	if result.PhotosRequested != 4 || result.PhotosGenerated != 4 {
		t.Fatalf("photos requested/generated = %d/%d, want 4/4", result.PhotosRequested, result.PhotosGenerated)
	}
	if generator.calls != 4 {
		t.Fatalf("generator calls = %d, want 4", generator.calls)
	}
}

func TestGenerateRecipePhotosLimitsConcurrency(t *testing.T) {
	generator := &fakeRecipeImageGenerator{
		images: []string{"first-image", "second-image", "third-image", "fourth-image"},
		delay:  10 * time.Millisecond,
	}

	result, err := generateRecipePhotos(context.Background(), generator, generateRecipePhotosArgs{
		Name:  "Soup",
		Count: 4,
	}, 2)
	if err != nil {
		t.Fatalf("generateRecipePhotos() error = %v", err)
	}
	if result.PhotosGenerated != 4 {
		t.Fatalf("PhotosGenerated = %d, want 4", result.PhotosGenerated)
	}
	if generator.maxActive > 2 {
		t.Fatalf("max active generators = %d, want <= 2", generator.maxActive)
	}
}

func TestRecipeImagePromptsAskForDistinctPresentations(t *testing.T) {
	prompts := recipeImagePrompts(generateRecipePhotosArgs{
		Name:        "Pasta",
		Description: "Bright weeknight dinner",
		Category:    "Dinner",
		Ingredients: []string{" pasta ", "", " basil "},
		UserRequest: "make it rustic",
		Count:       2,
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
		if !strings.Contains(prompt, "make it rustic") {
			t.Fatalf("prompt = %q, want user request", prompt)
		}
	}
}
