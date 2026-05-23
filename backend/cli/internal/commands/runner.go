package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	types "juancavallotti.com/recipe-types"
)

// ErrUsage is returned after a usage message has already been written.
var ErrUsage = errors.New("usage")

type RecipeRepo interface {
	GetRecipes(ctx context.Context) ([]types.Recipe, error)
	GetRecipe(ctx context.Context, id string) (types.Recipe, error)
	CreateRecipe(ctx context.Context, recipe types.Recipe) (string, error)
	UpdateRecipe(ctx context.Context, recipe types.Recipe) error
	AddRecipePhoto(ctx context.Context, recipeID string, photo types.Photo) (string, error)
	DeleteRecipePhoto(ctx context.Context, recipeID string, photoID string) error
	SetFeaturedRecipePhoto(ctx context.Context, recipeID string, photoID string) error
	DeleteRecipe(ctx context.Context, id string) error
	ImportRecipe(ctx context.Context, recipe types.Recipe) error
	LogTrace(ctx context.Context, eventID string, occurredAt time.Time, data json.RawMessage) error
}

type RepoFactory func() (RecipeRepo, error)

type Runner struct {
	stdin       io.Reader
	stdout      io.Writer
	stderr      io.Writer
	repoFactory RepoFactory
}

func NewRunner(stdin io.Reader, stdout io.Writer, stderr io.Writer, repoFactory RepoFactory) Runner {
	return Runner{
		stdin:       stdin,
		stdout:      stdout,
		stderr:      stderr,
		repoFactory: repoFactory,
	}
}

func (r Runner) Run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		r.usage()
		return ErrUsage
	}

	if args[0] == "schema" {
		if len(args) != 1 {
			return r.usageError("usage: recipes-cli schema")
		}
		return r.cmdSchema()
	}

	repo, err := r.repoFactory()
	if err != nil {
		return fmt.Errorf("repo: %w", err)
	}

	switch args[0] {
	case "list":
		if len(args) != 1 {
			return r.usageError("usage: recipes-cli list")
		}
		return r.cmdList(ctx, repo)
	case "export":
		if len(args) != 2 && len(args) != 3 {
			return r.usageError("usage: recipes-cli export <id> [--image-contents]")
		}
		if len(args) == 3 && args[2] != "--image-contents" {
			return r.usageError("usage: recipes-cli export <id> [--image-contents]")
		}
		return r.cmdExport(ctx, repo, args[1], len(args) == 3)
	case "export-all":
		if len(args) != 1 && len(args) != 2 {
			return r.usageError("usage: recipes-cli export-all [--image-contents]")
		}
		if len(args) == 2 && args[1] != "--image-contents" {
			return r.usageError("usage: recipes-cli export-all [--image-contents]")
		}
		return r.cmdExportAll(ctx, repo, len(args) == 2)
	case "create":
		const usage = "usage: recipes-cli create <path> [--json]"
		if len(args) < 2 || len(args) > 3 {
			return r.usageError(usage)
		}
		returnJSON, ok := parseJSONFlag(args[2:])
		if !ok {
			return r.usageError(usage)
		}
		return r.cmdCreate(ctx, repo, args[1], returnJSON)
	case "patch":
		const usage = "usage: recipes-cli patch <id> <path> [--json]"
		if len(args) < 3 || len(args) > 4 {
			return r.usageError(usage)
		}
		returnJSON, ok := parseJSONFlag(args[3:])
		if !ok {
			return r.usageError(usage)
		}
		return r.cmdPatch(ctx, repo, args[1], args[2], returnJSON)
	case "delete":
		if len(args) != 2 {
			return r.usageError("usage: recipes-cli delete <id>")
		}
		return r.cmdDelete(ctx, repo, args[1])
	case "add-photo":
		const usage = "usage: recipes-cli add-photo <recipe-id> <image-path|-> [--featured] [--json]"
		if len(args) < 3 || len(args) > 5 {
			return r.usageError(usage)
		}
		var featured, returnJSON bool
		for _, a := range args[3:] {
			switch a {
			case "--featured":
				if featured {
					return r.usageError(usage)
				}
				featured = true
			case "--json":
				if returnJSON {
					return r.usageError(usage)
				}
				returnJSON = true
			default:
				return r.usageError(usage)
			}
		}
		return r.cmdAddPhoto(ctx, repo, args[1], args[2], featured, returnJSON)
	case "delete-photo":
		const usage = "usage: recipes-cli delete-photo <recipe-id> <photo-id> [--json]"
		if len(args) < 3 || len(args) > 4 {
			return r.usageError(usage)
		}
		returnJSON, ok := parseJSONFlag(args[3:])
		if !ok {
			return r.usageError(usage)
		}
		return r.cmdDeletePhoto(ctx, repo, args[1], args[2], returnJSON)
	case "set-featured-photo":
		const usage = "usage: recipes-cli set-featured-photo <recipe-id> <photo-id> [--json]"
		if len(args) < 3 || len(args) > 4 {
			return r.usageError(usage)
		}
		returnJSON, ok := parseJSONFlag(args[3:])
		if !ok {
			return r.usageError(usage)
		}
		return r.cmdSetFeaturedPhoto(ctx, repo, args[1], args[2], returnJSON)
	case "import":
		if len(args) != 2 {
			return r.usageError("usage: recipes-cli import <path>")
		}
		return r.cmdImport(ctx, repo, args[1])
	case "log-trace":
		return r.cmdLogTrace(ctx, repo, args[1:])
	default:
		r.usage()
		return ErrUsage
	}
}

func (r Runner) usage() {
	fmt.Fprintf(r.stderr, `recipes-cli — recipe backup and inspection (uses DB_* from .env like the API).

Update commands print a short success summary by default. Pass --json to print the
full updated recipe as indented JSON instead (useful for piping into other tools).

Commands:
  list                          Print recipe id and title (name), tab-separated.
  export <id> [--image-contents]
                                Print one recipe as indented JSON. Photo image base64 data is omitted
                                unless --image-contents is given.
  export-all [--image-contents]
                                Print all recipes as JSON Lines (one JSON object per line). Photo image
                                base64 data is omitted unless --image-contents is given.
  create <path> [--json]        Read one recipe JSON object (use "-" for stdin); create it.
  patch <id> <path> [--json]    Read one partial recipe JSON object (use "-" for stdin); patch it.
  delete <id>                   Delete one recipe by id.
  add-photo <id> <path|-> [--featured] [--json]
                                Attach a photo; pass "-" to read raw base64 image data from stdin.
  delete-photo <id> <photo-id> [--json]
                                Remove a photo from a recipe by photo id.
  set-featured-photo <id> <photo-id> [--json]
                                Mark a recipe photo as featured.
  import <path>                 Read JSONL from file (use "-" for stdin); upsert each recipe.
  log-trace [--event-id-field <name>] [--time-field <name>]
                                Read JSON-lines from stdin; insert each as a trace row.
                                event_id   <- named field (default: invocation_id).
                                occurred_at <- named field, RFC3339 (default: time).
  schema                        Print the JSON Schema for create and patch payloads.

`)
}

func (r Runner) usageError(msg string) error {
	fmt.Fprintln(r.stderr, msg)
	return ErrUsage
}

// parseJSONFlag accepts zero args or a single "--json" arg. Returns
// (returnJSON, ok). Anything else returns ok=false so the caller can emit
// a usage error.
func parseJSONFlag(extra []string) (bool, bool) {
	switch len(extra) {
	case 0:
		return false, true
	case 1:
		if extra[0] == "--json" {
			return true, true
		}
	}
	return false, false
}
