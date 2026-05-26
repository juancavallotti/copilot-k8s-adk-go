package commands

import (
	"context"
	"fmt"
)

// cmdEmbedTest is a smoke-test for the embedding client. Prints the
// vector dimensions and the first few values so the operator can confirm
// the GEMINI_API_KEY is wired up correctly.
func (r Runner) cmdEmbedTest(ctx context.Context, repo EmbedRepo, text string) error {
	vec, err := repo.Embed(ctx, text)
	if err != nil {
		return fmt.Errorf("embed: %w", err)
	}
	preview := vec
	if len(preview) > 5 {
		preview = preview[:5]
	}
	fmt.Fprintf(r.stdout, "dimensions=%d preview=%v\n", len(vec), preview)
	return nil
}
