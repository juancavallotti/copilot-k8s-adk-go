package imagegen

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"

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

// describeOpenAIError formats the most informative parts of an OpenAI SDK
// error. The SDK's default Error() only includes Message; we also surface
// Type, Param, and Code (where present), plus the raw body for non-API errors.
func describeOpenAIError(err error) string {
	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		return fmt.Sprintf(
			"status=%d type=%q code=%v param=%q message=%q",
			apiErr.HTTPStatusCode,
			apiErr.Type,
			apiErr.Code,
			derefString(apiErr.Param),
			apiErr.Message,
		)
	}
	var reqErr *openai.RequestError
	if errors.As(err, &reqErr) {
		return fmt.Sprintf(
			"status=%d body=%q wrapped=%v",
			reqErr.HTTPStatusCode,
			string(reqErr.Body),
			reqErr.Err,
		)
	}
	return err.Error()
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (g *OpenAIRecipeImageGenerator) GenerateRecipeImage(ctx context.Context, prompt string) ([]byte, error) {
	if g == nil || g.client == nil {
		return nil, errors.New("openai image generator is not configured")
	}
	// gpt-image-1 family always returns base64-encoded images and rejects
	// the response_format parameter, so we don't set it. dall-e-* would
	// need it explicitly, but we don't target those models.
	resp, err := g.client.CreateImage(ctx, openai.ImageRequest{
		Prompt: prompt,
		Model:  g.model,
		N:      1,
		Size:   openai.CreateImageSize1024x1024,
	})
	if err != nil {
		log.Printf("openai create image: model=%q error=%s", g.model, describeOpenAIError(err))
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
