package imagegen

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type OpenAIRecipeImageGenerator struct {
	client *openai.Client
	model  string
}

func NewOpenAIRecipeImageGenerator(apiKey string, model string) (*OpenAIRecipeImageGenerator, error) {
	if apiKey == "" {
		return nil, errors.New("openai image generator: api key is required")
	}
	if model == "" {
		return nil, errors.New("openai image generator: model is required")
	}
	return &OpenAIRecipeImageGenerator{
		client: openai.NewClient(apiKey),
		model:  model,
	}, nil
}

func (g *OpenAIRecipeImageGenerator) GenerateRecipeImage(ctx context.Context, prompt string) ([]byte, error) {
	if g == nil || g.client == nil {
		return nil, errors.New("openai image generator is not configured")
	}
	resp, err := g.client.CreateImage(ctx, openai.ImageRequest{
		Prompt:         prompt,
		Model:          g.model,
		N:              1,
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
	})
	if err != nil {
		return nil, fmt.Errorf("openai create image: %w", err)
	}
	if len(resp.Data) == 0 || resp.Data[0].B64JSON == "" {
		return nil, errors.New("openai image response did not include b64_json data")
	}
	data, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)
	if err != nil {
		return nil, fmt.Errorf("decode openai image b64: %w", err)
	}
	return data, nil
}
