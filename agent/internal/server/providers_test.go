package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"juancavallotti.com/recipes-agent/internal/modelrouter"
)

func TestProvidersHandlerReturnsRegistryContents(t *testing.T) {
	registry := &modelrouter.Registry{
		DefaultAgent: "google:gemini-3.1-flash-lite",
		DefaultImage: "google:gemini-3.1-flash-image-preview",
		AgentOptions: []modelrouter.AgentOption{{
			ID: "google:gemini-3.1-flash-lite", Provider: "google",
			Model: "gemini-3.1-flash-lite", Label: "Google · gemini-3.1-flash-lite",
		}},
		ImageOptions: []modelrouter.ImageOption{{
			ID: "google:gemini-3.1-flash-image-preview", Provider: "google",
			Model: "gemini-3.1-flash-image-preview", Label: "Google · gemini-3.1-flash-image-preview",
		}},
	}

	req := httptest.NewRequest(http.MethodGet, "/agent/providers", nil)
	rec := httptest.NewRecorder()

	providers(registry)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body providersResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.DefaultAgentModel != registry.DefaultAgent {
		t.Errorf("DefaultAgentModel = %q, want %q", body.DefaultAgentModel, registry.DefaultAgent)
	}
	if body.DefaultImageModel != registry.DefaultImage {
		t.Errorf("DefaultImageModel = %q, want %q", body.DefaultImageModel, registry.DefaultImage)
	}
	if len(body.AgentOptions) != 1 || body.AgentOptions[0].ID != registry.AgentOptions[0].ID {
		t.Errorf("AgentOptions = %+v, want %+v", body.AgentOptions, registry.AgentOptions)
	}
	if len(body.ImageOptions) != 1 || body.ImageOptions[0].ID != registry.ImageOptions[0].ID {
		t.Errorf("ImageOptions = %+v, want %+v", body.ImageOptions, registry.ImageOptions)
	}
}

func TestProvidersHandlerRejectsNonGET(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/agent/providers", nil)
	rec := httptest.NewRecorder()

	providers(&modelrouter.Registry{})(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}
