package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	repo "juancavallotti.com/recipes-repo"
)

// cmdSearch implements `recipes-cli search-recipes <query>` and
// `recipes-cli search-events <query>`. Default output is one row per
// match, tab-separated: score, id, title-or-prompt. --json emits one
// JSON object per line so the agent can parse it.
func (r Runner) cmdSearch(ctx context.Context, cmdRepo EmbedRepo, target string, args []string) error {
	usage := fmt.Sprintf("usage: recipes-cli search-%s <query> [--limit N] [--json]", target)
	if len(args) < 1 {
		return r.usageError(usage)
	}

	var queryParts []string
	limit := 10
	returnJSON := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--limit":
			if i+1 >= len(args) {
				return r.usageError(usage)
			}
			n, err := strconv.Atoi(args[i+1])
			if err != nil || n <= 0 {
				return r.usageError(usage)
			}
			limit = n
			i++
		case "--json":
			returnJSON = true
		default:
			queryParts = append(queryParts, args[i])
		}
	}
	query := strings.TrimSpace(strings.Join(queryParts, " "))
	if query == "" {
		return r.usageError(usage)
	}

	switch target {
	case "recipes":
		return r.searchRecipes(ctx, cmdRepo, query, limit, returnJSON)
	case "events":
		return r.searchEvents(ctx, cmdRepo, query, limit, returnJSON)
	default:
		return r.usageError(usage)
	}
}

func (r Runner) searchRecipes(ctx context.Context, cmdRepo EmbedRepo, query string, limit int, returnJSON bool) error {
	hits, err := cmdRepo.SearchRecipeChunks(ctx, query, limit)
	if err != nil {
		if errors.Is(err, repo.ErrSearchDisabled) {
			fmt.Fprintln(r.stderr, "search disabled: set GEMINI_API_KEY or OPENAI_API_KEY")
			return err
		}
		return fmt.Errorf("search-recipes: %w", err)
	}
	if returnJSON {
		enc := json.NewEncoder(r.stdout)
		for _, h := range hits {
			_ = enc.Encode(h)
		}
		return nil
	}
	if len(hits) == 0 {
		fmt.Fprintln(r.stdout, "no matches")
		return nil
	}
	fmt.Fprintln(r.stdout, "SCORE\tID\tTITLE\tCHUNK")
	for _, h := range hits {
		fmt.Fprintf(r.stdout, "%.4f\t%s\t%s\t%s\n", h.Score, h.ID, h.Name, truncate(h.Chunk, 80))
	}
	return nil
}

func (r Runner) searchEvents(ctx context.Context, cmdRepo EmbedRepo, query string, limit int, returnJSON bool) error {
	matches, err := cmdRepo.SearchEvents(ctx, query, limit)
	if err != nil {
		if errors.Is(err, repo.ErrSearchDisabled) {
			fmt.Fprintln(r.stderr, "search disabled: set GEMINI_API_KEY or OPENAI_API_KEY")
			return err
		}
		return fmt.Errorf("search-events: %w", err)
	}
	if returnJSON {
		enc := json.NewEncoder(r.stdout)
		for _, m := range matches {
			_ = enc.Encode(m)
		}
		return nil
	}
	if len(matches) == 0 {
		fmt.Fprintln(r.stdout, "no matches")
		return nil
	}
	fmt.Fprintln(r.stdout, "SCORE\tEVENT_ID\tPROMPT")
	for _, m := range matches {
		fmt.Fprintf(r.stdout, "%.4f\t%s\t%s\n", m.Score, m.EventID, truncate(m.UserPrompt, 80))
	}
	return nil
}

// truncate caps a single-line string so the tab-separated default
// output stays scannable. JSON output keeps the full text.
func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
