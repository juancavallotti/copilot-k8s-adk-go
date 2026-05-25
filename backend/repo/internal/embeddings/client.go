// Package embeddings produces vector embeddings used by the semantic
// search subsystem. Two providers are supported (Gemini and OpenAI);
// selection is driven by which API keys are present in the environment
// and by an optional EMBEDDING_PROVIDER override. A no-op implementation
// is returned when neither key is configured so local dev still builds
// and runs.
package embeddings

import (
	"context"
	"errors"
	"os"
	"strings"
)

// Dimensions is the vector size every Client must produce. It matches
// the vector(N) columns declared in database/db.sql. Both Gemini and
// OpenAI are asked for this dimensionality explicitly.
const Dimensions = 768

// ErrDisabled is returned by the no-op client when no API key is set.
var ErrDisabled = errors.New("embeddings: no API key configured")

// Client produces a Dimensions-sized embedding vector for a text input.
type Client interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// Provider names the active embedding provider. Used for logs and so
// callers can show which model produced the vectors.
type Provider string

const (
	ProviderNoop   Provider = "noop"
	ProviderGemini Provider = "gemini"
	ProviderOpenAI Provider = "openai"
)

// NewFromEnv picks a Client based on env vars:
//
//   - EMBEDDING_PROVIDER, when set to "gemini" or "openai", forces that
//     provider — falling back to Noop if the matching API key is missing.
//   - With no override, the first present key wins: Gemini if
//     GEMINI_API_KEY is set, else OpenAI if OPENAI_API_KEY is set, else
//     Noop. (Gemini-first because it's our default per the design doc.)
//
// The returned Provider tells callers which path was chosen.
func NewFromEnv() (Client, Provider) {
	gemKey := os.Getenv("GEMINI_API_KEY")
	oaiKey := os.Getenv("OPENAI_API_KEY")
	switch strings.ToLower(strings.TrimSpace(os.Getenv("EMBEDDING_PROVIDER"))) {
	case "gemini":
		if gemKey != "" {
			return NewGeminiClient(gemKey), ProviderGemini
		}
		return Noop{}, ProviderNoop
	case "openai":
		if oaiKey != "" {
			return NewOpenAIClient(oaiKey), ProviderOpenAI
		}
		return Noop{}, ProviderNoop
	}
	if gemKey != "" {
		return NewGeminiClient(gemKey), ProviderGemini
	}
	if oaiKey != "" {
		return NewOpenAIClient(oaiKey), ProviderOpenAI
	}
	return Noop{}, ProviderNoop
}
