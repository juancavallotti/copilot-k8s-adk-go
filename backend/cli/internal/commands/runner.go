package commands

import (
	"context"
	"errors"
	"fmt"
	"io"

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
	ImportRecipe(ctx context.Context, recipe types.Recipe) error
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
		if len(args) != 2 {
			return r.usageError("usage: recipes-cli export <id>")
		}
		return r.cmdExport(ctx, repo, args[1])
	case "export-all":
		if len(args) != 1 {
			return r.usageError("usage: recipes-cli export-all")
		}
		return r.cmdExportAll(ctx, repo)
	case "create":
		if len(args) != 2 {
			return r.usageError("usage: recipes-cli create <path>")
		}
		return r.cmdCreate(ctx, repo, args[1])
	case "patch":
		if len(args) != 3 {
			return r.usageError("usage: recipes-cli patch <id> <path>")
		}
		return r.cmdPatch(ctx, repo, args[1], args[2])
	case "add-photo":
		if len(args) != 3 && len(args) != 4 {
			return r.usageError("usage: recipes-cli add-photo <recipe-id> <image-path> [--featured]")
		}
		if len(args) == 4 && args[3] != "--featured" {
			return r.usageError("usage: recipes-cli add-photo <recipe-id> <image-path> [--featured]")
		}
		return r.cmdAddPhoto(ctx, repo, args[1], args[2], len(args) == 4)
	case "delete-photo":
		if len(args) != 3 {
			return r.usageError("usage: recipes-cli delete-photo <recipe-id> <photo-id>")
		}
		return r.cmdDeletePhoto(ctx, repo, args[1], args[2])
	case "set-featured-photo":
		if len(args) != 3 {
			return r.usageError("usage: recipes-cli set-featured-photo <recipe-id> <photo-id>")
		}
		return r.cmdSetFeaturedPhoto(ctx, repo, args[1], args[2])
	case "import":
		if len(args) != 2 {
			return r.usageError("usage: recipes-cli import <path>")
		}
		return r.cmdImport(ctx, repo, args[1])
	default:
		r.usage()
		return ErrUsage
	}
}

func (r Runner) usage() {
	fmt.Fprintf(r.stderr, `recipes-cli — recipe backup and inspection (uses DB_* from .env like the API).

Commands:
  list                          Print recipe id and title (name), tab-separated.
  export <id>                   Print one recipe as indented JSON.
  export-all                    Print all recipes as JSON Lines (one JSON object per line).
  create <path>                 Read one recipe JSON object (use "-" for stdin); create it.
  patch <id> <path>             Read one partial recipe JSON object (use "-" for stdin); patch it.
  add-photo <id> <path> [--featured]
                                Base64-encode an image file and attach it to a recipe.
  delete-photo <id> <photo-id>  Remove a photo from a recipe by photo id.
  set-featured-photo <id> <photo-id>
                                Mark a recipe photo as featured.
  import <path>                 Read JSONL from file (use "-" for stdin); upsert each recipe.
  schema                        Print the JSON Schema for create and patch payloads.

`)
}

func (r Runner) usageError(msg string) error {
	fmt.Fprintln(r.stderr, msg)
	return ErrUsage
}
