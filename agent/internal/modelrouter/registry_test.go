package modelrouter

import (
	"testing"

	"juancavallotti.com/recipes-agent/internal/config"
)

func TestBuildRegistryRegistersGoogleDefaults(t *testing.T) {
	cfg := config.Config{
		GeminiAPIKey: "fake-key",
		Model:        "gemini-3.1-flash-lite",
		ImageModel:   "gemini-3.1-flash-image-preview",
	}

	r, err := BuildRegistry(cfg)
	if err != nil {
		t.Fatalf("BuildRegistry() error = %v", err)
	}

	wantAgent := "google:gemini-3.1-flash-lite"
	wantImage := "google:gemini-3.1-flash-image-preview"

	if r.DefaultAgent != wantAgent {
		t.Errorf("DefaultAgent = %q, want %q", r.DefaultAgent, wantAgent)
	}
	if r.DefaultImage != wantImage {
		t.Errorf("DefaultImage = %q, want %q", r.DefaultImage, wantImage)
	}
	if len(r.AgentOptions) != 1 || r.AgentOptions[0].ID != wantAgent {
		t.Errorf("AgentOptions = %+v, want one entry with ID %q", r.AgentOptions, wantAgent)
	}
	if len(r.ImageOptions) != 1 || r.ImageOptions[0].ID != wantImage {
		t.Errorf("ImageOptions = %+v, want one entry with ID %q", r.ImageOptions, wantImage)
	}
}

func TestBuildRegistryRequiresGeminiAPIKey(t *testing.T) {
	if _, err := BuildRegistry(config.Config{}); err == nil {
		t.Fatal("BuildRegistry() with empty GeminiAPIKey: got nil error, want error")
	}
}

func TestRegistryBuilderLookup(t *testing.T) {
	r, err := BuildRegistry(config.Config{
		GeminiAPIKey: "fake-key",
		Model:        "gemini-3.1-flash-lite",
		ImageModel:   "gemini-3.1-flash-image-preview",
	})
	if err != nil {
		t.Fatalf("BuildRegistry() error = %v", err)
	}

	if _, ok := r.AgentBuilder(r.DefaultAgent); !ok {
		t.Errorf("AgentBuilder(%q) not found", r.DefaultAgent)
	}
	if _, ok := r.ImageBuilder(r.DefaultImage); !ok {
		t.Errorf("ImageBuilder(%q) not found", r.DefaultImage)
	}
	if _, ok := r.AgentBuilder("openai:something"); ok {
		t.Error("AgentBuilder(unknown) returned ok=true, want false")
	}
	if _, ok := r.ImageBuilder("openai:something"); ok {
		t.Error("ImageBuilder(unknown) returned ok=true, want false")
	}
}
