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

const geminiEndpointFmt = "https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent"

// GeminiModel is the embedding model used for Gemini requests.
// gemini-embedding-001 supports MRL truncation via outputDimensionality,
// which we set to 768 to match the pgvector column.
const GeminiModel = "gemini-embedding-001"

// GeminiClient calls Gemini's :embedContent REST endpoint directly. The
// REST shape is small enough that pulling in the full google.golang.org/genai
// SDK (and its ~30 indirect deps) would be more weight than this module needs.
type GeminiClient struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
}

// NewGeminiClient returns a client that signs requests with apiKey.
func NewGeminiClient(apiKey string) *GeminiClient {
	return &GeminiClient{
		apiKey:     apiKey,
		endpoint:   fmt.Sprintf(geminiEndpointFmt, GeminiModel),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type embedRequest struct {
	Content              embedContent `json:"content"`
	OutputDimensionality int          `json:"outputDimensionality,omitempty"`
}

type embedContent struct {
	Parts []embedPart `json:"parts"`
}

type embedPart struct {
	Text string `json:"text"`
}

type embedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Embed returns the Gemini embedding for text. Empty input is rejected
// because the model errors out on it and the caller almost certainly
// has a bug.
func (c *GeminiClient) Embed(ctx context.Context, text string) ([]float32, error) {
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("embeddings: empty input")
	}
	payload, err := json.Marshal(embedRequest{
		Content:              embedContent{Parts: []embedPart{{Text: text}}},
		OutputDimensionality: Dimensions,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embed request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read embed response: %w", err)
	}
	var out embedResponse
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
	if len(out.Embedding.Values) != Dimensions {
		return nil, fmt.Errorf("embed returned %d values, want %d", len(out.Embedding.Values), Dimensions)
	}
	return out.Embedding.Values, nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}
