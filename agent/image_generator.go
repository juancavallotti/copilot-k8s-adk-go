package main

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/genai"
)

type recipeImageGenerator interface {
	GenerateRecipeImage(ctx context.Context, prompt string) ([]byte, error)
}

type geminiRecipeImageGenerator struct {
	client *genai.Client
	model  string
}

func newGeminiRecipeImageGenerator(ctx context.Context, cfg config) (*geminiRecipeImageGenerator, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.GeminiAPIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}
	return &geminiRecipeImageGenerator{
		client: client,
		model:  cfg.ImageModel,
	}, nil
}

func (g *geminiRecipeImageGenerator) GenerateRecipeImage(ctx context.Context, prompt string) ([]byte, error) {
	if g == nil || g.client == nil {
		return nil, errors.New("image generator is not configured")
	}
	response, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(prompt), &genai.GenerateContentConfig{
		ResponseModalities: []string{"IMAGE", "TEXT"},
	})
	if err != nil {
		return nil, err
	}
	data, err := firstInlineImageData(response)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func firstInlineImageData(response *genai.GenerateContentResponse) ([]byte, error) {
	if response == nil {
		return nil, errors.New("image response was empty")
	}
	for _, candidate := range response.Candidates {
		if candidate == nil || candidate.Content == nil {
			continue
		}
		for _, part := range candidate.Content.Parts {
			if part == nil || part.InlineData == nil || len(part.InlineData.Data) == 0 {
				continue
			}
			return part.InlineData.Data, nil
		}
	}
	return nil, errors.New("image response did not include inline image data")
}
