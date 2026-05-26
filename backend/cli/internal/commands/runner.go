package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	types "juancavallotti.com/recipe-types"
	repo "juancavallotti.com/recipes-repo"
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
}

type TraceRepo interface {
	LogTrace(ctx context.Context, eventID string, occurredAt time.Time, data json.RawMessage) error
	ListEvents(ctx context.Context, limit, offset int) ([]types.Event, error)
	ListTracesByEvent(ctx context.Context, eventID string, limit, offset int) ([]types.Trace, error)
}

type SkillRepo interface {
	ListSkills(ctx context.Context) ([]types.Skill, error)
	GetSkillByName(ctx context.Context, name string) (types.Skill, error)
}

type EmbedRepo interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	ReindexRecipes(ctx context.Context, opts repo.ReindexOptions) error
	ReindexEvents(ctx context.Context, opts repo.ReindexEventsOptions) error
	SearchRecipes(ctx context.Context, query string, limit int) ([]types.RecipeMatch, error)
	SearchRecipeChunks(ctx context.Context, query string, limit int) ([]types.RecipeHit, error)
	SearchEvents(ctx context.Context, query string, limit int) ([]types.EventMatch, error)
}

// openRepo is the package hook for opening a real repo. Production calls
// repo.NewRepo; tests can swap it to assert opening behavior. There is no
// runner-level union interface: cmd methods take narrow interfaces and
// tests call them directly with fakes.
var openRepo = repo.NewRepo

type Runner struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
	logger *slog.Logger
}

func NewRunner(stdin io.Reader, stdout io.Writer, stderr io.Writer) Runner {
	return NewRunnerWithLogger(stdin, stdout, stderr, slog.New(slog.NewJSONHandler(stderr, nil)))
}

func NewRunnerWithLogger(stdin io.Reader, stdout io.Writer, stderr io.Writer, logger *slog.Logger) Runner {
	if logger == nil {
		logger = slog.Default()
	}
	return Runner{
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		logger: logger,
	}
}

func (r Runner) log() *slog.Logger {
	if r.logger != nil {
		return r.logger
	}
	return slog.Default()
}

func (r Runner) Run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		r.printHelp()
		return nil
	}

	switch args[0] {
	case "-h", "--help", "help":
		r.printHelp()
		return nil
	}

	if args[0] == "schema" {
		if len(args) != 1 {
			return r.usageError("usage: recipes-cli schema")
		}
		return r.cmdSchema()
	}

	rp, err := openRepo()
	if err != nil {
		return fmt.Errorf("repo: %w", err)
	}
	defer func() {
		if cerr := rp.Close(); cerr != nil {
			r.log().Warn("cli.repo_close_failed", "err", cerr)
		}
	}()

	switch args[0] {
	case "list":
		if len(args) != 1 {
			return r.usageError("usage: recipes-cli list")
		}
		return r.cmdList(ctx, rp)
	case "export":
		if len(args) != 2 && len(args) != 3 {
			return r.usageError("usage: recipes-cli export <id> [--image-contents]")
		}
		if len(args) == 3 && args[2] != "--image-contents" {
			return r.usageError("usage: recipes-cli export <id> [--image-contents]")
		}
		return r.cmdExport(ctx, rp, args[1], len(args) == 3)
	case "export-all":
		if len(args) != 1 && len(args) != 2 {
			return r.usageError("usage: recipes-cli export-all [--image-contents]")
		}
		if len(args) == 2 && args[1] != "--image-contents" {
			return r.usageError("usage: recipes-cli export-all [--image-contents]")
		}
		return r.cmdExportAll(ctx, rp, len(args) == 2)
	case "create":
		const usage = "usage: recipes-cli create <path> [--json]"
		if len(args) < 2 || len(args) > 3 {
			return r.usageError(usage)
		}
		returnJSON, ok := parseJSONFlag(args[2:])
		if !ok {
			return r.usageError(usage)
		}
		return r.cmdCreate(ctx, rp, args[1], returnJSON)
	case "patch":
		const usage = "usage: recipes-cli patch <id> <path> [--json]"
		if len(args) < 3 || len(args) > 4 {
			return r.usageError(usage)
		}
		returnJSON, ok := parseJSONFlag(args[3:])
		if !ok {
			return r.usageError(usage)
		}
		return r.cmdPatch(ctx, rp, args[1], args[2], returnJSON)
	case "delete":
		if len(args) != 2 {
			return r.usageError("usage: recipes-cli delete <id>")
		}
		return r.cmdDelete(ctx, rp, args[1])
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
		return r.cmdAddPhoto(ctx, rp, args[1], args[2], featured, returnJSON)
	case "delete-photo":
		const usage = "usage: recipes-cli delete-photo <recipe-id> <photo-id> [--json]"
		if len(args) < 3 || len(args) > 4 {
			return r.usageError(usage)
		}
		returnJSON, ok := parseJSONFlag(args[3:])
		if !ok {
			return r.usageError(usage)
		}
		return r.cmdDeletePhoto(ctx, rp, args[1], args[2], returnJSON)
	case "set-featured-photo":
		const usage = "usage: recipes-cli set-featured-photo <recipe-id> <photo-id> [--json]"
		if len(args) < 3 || len(args) > 4 {
			return r.usageError(usage)
		}
		returnJSON, ok := parseJSONFlag(args[3:])
		if !ok {
			return r.usageError(usage)
		}
		return r.cmdSetFeaturedPhoto(ctx, rp, args[1], args[2], returnJSON)
	case "import":
		if len(args) != 2 {
			return r.usageError("usage: recipes-cli import <path>")
		}
		return r.cmdImport(ctx, rp, args[1])
	case "log-trace":
		return r.cmdLogTrace(ctx, rp, args[1:])
	case "list-events":
		return r.cmdListEvents(ctx, rp, args[1:])
	case "list-traces":
		return r.cmdListTraces(ctx, rp, args[1:])
	case "embed-test":
		if len(args) != 2 {
			return r.usageError("usage: recipes-cli embed-test <text>")
		}
		return r.cmdEmbedTest(ctx, rp, args[1])
	case "reindex":
		return r.cmdReindex(ctx, rp, args[1:])
	case "search-recipes":
		return r.cmdSearch(ctx, rp, "recipes", args[1:])
	case "search-events":
		return r.cmdSearch(ctx, rp, "events", args[1:])
	case "list-skills":
		if len(args) != 1 {
			return r.usageError("usage: recipes-cli list-skills")
		}
		return r.cmdListSkills(ctx, rp)
	case "load-skill":
		if len(args) != 2 {
			return r.usageError("usage: recipes-cli load-skill <name>")
		}
		return r.cmdLoadSkill(ctx, rp, args[1])
	default:
		r.usage()
		return ErrUsage
	}
}

func (r Runner) printHelp() {
	fmt.Fprint(r.stdout, helpText)
}

func (r Runner) usage() {
	fmt.Fprint(r.stderr, helpText)
}

const helpText = `recipes-cli — recipe backup and inspection (uses DB_* from .env like the API).

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
  list-events [--limit N] [--offset N]
                                Print events as JSON Lines (one JSON object per line),
                                newest first. limit defaults to 50, max 200.
  list-traces <event-id> [--limit N] [--offset N]
                                Print traces for an event as JSON Lines, oldest first.
                                limit defaults to 50, max 200.
  list-skills                   Print available skills as JSON Lines: {name, description}.
                                Use load-skill to fetch the full instructions for one.
  load-skill <name>             Print the markdown content of one skill to stdout.
                                Exits non-zero if no skill has that name.
  schema                        Print the JSON Schema for create and patch payloads.
  embed-test <text>             Smoke-test the embeddings client. Prints vector
                                dimensions and a short preview.
  reindex --target {recipes|events|all} [--force] [--limit N] [--json]
                                Rebuild the semantic-search index. --force re-embeds
                                rows that already have embeddings; without it only
                                rows missing an embedding are processed. --json
                                streams one report object per line, agent-readable.
                                Exit 0 = all ok, 1 = at least one row failed,
                                2 = bad arguments.
  search-recipes <query> [--limit N] [--json]
                                Semantic search over recipes. Default output is
                                SCORE\tID\tTITLE\tCHUNK per line (chunk truncated).
                                --json emits one slim hit per line:
                                {id, name, chunk, score}. Use "export <id>" to
                                fetch the full recipe when you need one.
  search-events <query> [--limit N] [--json]
                                Semantic search over events by user_prompt. Default
                                output is SCORE\tEVENT_ID\tPROMPT per line; --json
                                emits one EventMatch object per line.

`

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
