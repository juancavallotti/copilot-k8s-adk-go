package embeddings

import "context"

// Noop is a Client that always returns ErrDisabled. It lets dev environments
// without GEMINI_API_KEY build and run; write hooks log the error and move on,
// search returns it as 503.
type Noop struct{}

func (Noop) Embed(_ context.Context, _ string) ([]float32, error) {
	return nil, ErrDisabled
}
