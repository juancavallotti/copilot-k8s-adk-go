package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/joho/godotenv"
	types "juancavallotti.com/recipe-types"
	"juancavallotti.com/recipes-repo"
)

func loadDotenv() {
	for _, path := range []string{".env", "backend/.env"} {
		if err := godotenv.Load(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			log.Printf("dotenv: load %q: %v", path, err)
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `recipes-cli — recipe backup and inspection (uses DB_* from .env like the API).

Commands:
  list                          Print recipe id and title (name), tab-separated.
  export <id>                   Print one recipe as indented JSON.
  export-all                    Print all recipes as JSON Lines (one JSON object per line).
  import <path>                 Read JSONL from file (use "-" for stdin); upsert each recipe.

`)
}

func main() {
	loadDotenv()
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	ctx := context.Background()
	r, err := repo.NewRepo()
	if err != nil {
		log.Fatalf("repo: %v", err)
	}

	switch os.Args[1] {
	case "list":
		if err := cmdList(ctx, r); err != nil {
			log.Fatal(err)
		}
	case "export":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "usage: recipes-cli export <id>")
			os.Exit(2)
		}
		if err := cmdExport(ctx, r, os.Args[2]); err != nil {
			log.Fatal(err)
		}
	case "export-all":
		if len(os.Args) != 2 {
			fmt.Fprintln(os.Stderr, "usage: recipes-cli export-all")
			os.Exit(2)
		}
		if err := cmdExportAll(ctx, r); err != nil {
			log.Fatal(err)
		}
	case "import":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "usage: recipes-cli import <path>")
			os.Exit(2)
		}
		if err := cmdImport(ctx, r, os.Args[2]); err != nil {
			log.Fatal(err)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func cmdList(ctx context.Context, r *repo.Repo) error {
	recipes, err := r.GetRecipes(ctx)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE")
	for _, rec := range recipes {
		fmt.Fprintf(w, "%s\t%s\n", rec.ID, rec.Name)
	}
	return w.Flush()
}

func cmdExport(ctx context.Context, r *repo.Repo, id string) error {
	rec, err := r.GetRecipe(ctx, strings.TrimSpace(id))
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(rec)
}

func cmdExportAll(ctx context.Context, r *repo.Repo) error {
	summaries, err := r.GetRecipes(ctx)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	for _, s := range summaries {
		rec, err := r.GetRecipe(ctx, s.ID)
		if err != nil {
			return fmt.Errorf("recipe %s: %w", s.ID, err)
		}
		if err := enc.Encode(rec); err != nil {
			return err
		}
	}
	return nil
}

func cmdImport(ctx context.Context, r *repo.Repo, path string) error {
	var in io.Reader
	if path == "-" {
		in = os.Stdin
	} else {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}

	sc := bufio.NewScanner(in)
	// Default buffer may be too small for long JSON lines.
	const max = 16 * 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, max)

	lineNum := 0
	for sc.Scan() {
		lineNum++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var rec types.Recipe
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
		}
		if err := r.ImportRecipe(ctx, rec); err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return nil
}
