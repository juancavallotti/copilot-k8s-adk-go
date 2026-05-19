// Package modelrouter holds the per-provider builders for chat models and
// image generators. In stage 1 only Google is registered; later stages add
// OpenAI and Anthropic.
package modelrouter

import (
	"context"
	"fmt"

	adkanthropic "github.com/Alcova-AI/adk-anthropic-go"
	adkopenai "github.com/byebyebruce/adk-go-openai"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"

	"juancavallotti.com/recipes-agent/internal/config"
	"juancavallotti.com/recipes-agent/internal/imagegen"
)

const (
	ProviderGoogle    = "google"
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
)

// AgentOption is a public description of one selectable chat model.
type AgentOption struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Label    string `json:"label"`
}

// ImageOption is a public description of one selectable image model.
type ImageOption struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Label    string `json:"label"`
}

// AgentBuilder constructs a chat model.
type AgentBuilder func(ctx context.Context) (model.LLM, error)

// ImageBuilder constructs an image generator.
type ImageBuilder func(ctx context.Context) (imagegen.RecipeImageGenerator, error)

// Registry holds the configured providers and their builders.
type Registry struct {
	DefaultAgent string
	DefaultImage string

	AgentOptions []AgentOption
	ImageOptions []ImageOption

	agentBuilders map[string]AgentBuilder
	imageBuilders map[string]ImageBuilder
}

// AgentBuilder returns the builder for a given agent option ID, or false.
func (r *Registry) AgentBuilder(id string) (AgentBuilder, bool) {
	b, ok := r.agentBuilders[id]
	return b, ok
}

// ImageBuilder returns the builder for a given image option ID, or false.
func (r *Registry) ImageBuilder(id string) (ImageBuilder, bool) {
	b, ok := r.imageBuilders[id]
	return b, ok
}

// BuildRegistry inspects cfg for available API keys and registers the
// matching providers. Google is always registered (its key is required at
// boot). Stage 1 only registers Google; OpenAI/Anthropic are added later.
func BuildRegistry(cfg config.Config) (*Registry, error) {
	r := &Registry{
		agentBuilders: map[string]AgentBuilder{},
		imageBuilders: map[string]ImageBuilder{},
	}

	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("modelrouter: GEMINI_API_KEY is required")
	}

	googleAgentID := id(ProviderGoogle, cfg.Model)
	googleImageID := id(ProviderGoogle, cfg.ImageModel)

	apiKey := cfg.GeminiAPIKey
	agentModel := cfg.Model
	imageModel := cfg.ImageModel

	r.agentBuilders[googleAgentID] = func(ctx context.Context) (model.LLM, error) {
		return gemini.NewModel(ctx, agentModel, &genai.ClientConfig{APIKey: apiKey})
	}
	r.imageBuilders[googleImageID] = func(ctx context.Context) (imagegen.RecipeImageGenerator, error) {
		return imagegen.NewGeminiRecipeImageGenerator(ctx, apiKey, imageModel)
	}

	r.AgentOptions = append(r.AgentOptions, AgentOption{
		ID:       googleAgentID,
		Provider: ProviderGoogle,
		Model:    agentModel,
		Label:    "Google · " + agentModel,
	})
	r.ImageOptions = append(r.ImageOptions, ImageOption{
		ID:       googleImageID,
		Provider: ProviderGoogle,
		Model:    imageModel,
		Label:    "Google · " + imageModel,
	})

	r.DefaultAgent = googleAgentID
	r.DefaultImage = googleImageID

	if cfg.AnthropicAPIKey != "" {
		anthropicKey := cfg.AnthropicAPIKey
		anthropicModel := cfg.AnthropicModel
		anthropicID := id(ProviderAnthropic, anthropicModel)
		r.agentBuilders[anthropicID] = func(ctx context.Context) (model.LLM, error) {
			return adkanthropic.NewModel(ctx, anthropicModel, &adkanthropic.Config{APIKey: anthropicKey})
		}
		r.AgentOptions = append(r.AgentOptions, AgentOption{
			ID:       anthropicID,
			Provider: ProviderAnthropic,
			Model:    anthropicModel,
			Label:    "Anthropic · " + anthropicModel,
		})
	}

	if cfg.OpenAIAPIKey != "" {
		openaiKey := cfg.OpenAIAPIKey
		openaiModel := cfg.OpenAIModel
		openaiImageModel := cfg.OpenAIImageModel

		openaiAgentID := id(ProviderOpenAI, openaiModel)
		r.agentBuilders[openaiAgentID] = func(ctx context.Context) (model.LLM, error) {
			return adkopenai.NewOpenAIModelWithAPIKey(openaiModel, openaiKey), nil
		}
		r.AgentOptions = append(r.AgentOptions, AgentOption{
			ID:       openaiAgentID,
			Provider: ProviderOpenAI,
			Model:    openaiModel,
			Label:    "OpenAI · " + openaiModel,
		})

		openaiImageID := id(ProviderOpenAI, openaiImageModel)
		r.imageBuilders[openaiImageID] = func(ctx context.Context) (imagegen.RecipeImageGenerator, error) {
			return imagegen.NewOpenAIRecipeImageGenerator(openaiKey, openaiImageModel)
		}
		r.ImageOptions = append(r.ImageOptions, ImageOption{
			ID:       openaiImageID,
			Provider: ProviderOpenAI,
			Model:    openaiImageModel,
			Label:    "OpenAI · " + openaiImageModel,
		})
	}

	return r, nil
}

func id(provider, model string) string {
	return provider + ":" + model
}
