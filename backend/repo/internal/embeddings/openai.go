package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const openAIEndpoint = "https://api.openai.com/v1/embeddings"

// OpenAIModel is the embedding model used for OpenAI requests. The
// text-embedding-3-small family accepts a `dimensions` parameter so we
// can request 768-dim vectors that fit the existing pgvector column —
// no schema change needed when swapping providers.
const OpenAIModel = "text-embedding-3-small"

// OpenAIClient calls OpenAI's /v1/embeddings endpoint directly. Kept lean
// for the same reason as GeminiClient: we only need one POST and don't
// want to pull the SDK and its deps into the repo module.
type OpenAIClient struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
}

// NewOpenAIClient returns a client that signs requests with apiKey.
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		apiKey:     apiKey,
		endpoint:   openAIEndpoint,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type openAIRequest struct {
	Input      string `json:"input"`
	Model      string `json:"model"`
	Dimensions int    `json:"dimensions,omitempty"`
}

type openAIResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Embed returns the OpenAI embedding for text. Asks the API for exactly
// Dimensions values so the response fits the pgvector(768) column.
func (c *OpenAIClient) Embed(ctx context.Context, text string) ([]float32, error) {
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("embeddings: empty input")
	}
	payload, err := json.Marshal(openAIRequest{
		Input:      text,
		Model:      OpenAIModel,
		Dimensions: Dimensions,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embed request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read embed response: %w", err)
	}
	var out openAIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode embed response: %w (body=%s)", err, truncate(body, 200))
	}
	if resp.StatusCode != http.StatusOK {
		msg := strings.TrimSpace(string(body))
		if out.Error != nil && out.Error.Message != "" {
			msg = out.Error.Message
		}
		return nil, fmt.Errorf("embed status %d: %s", resp.StatusCode, truncate([]byte(msg), 200))
	}
	if len(out.Data) == 0 {
		return nil, errors.New("embed response had no data")
	}
	values := out.Data[0].Embedding
	if len(values) != Dimensions {
		return nil, fmt.Errorf("embed returned %d values, want %d", len(values), Dimensions)
	}
	return values, nil
}
