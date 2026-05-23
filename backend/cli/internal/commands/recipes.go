package commands

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	types "juancavallotti.com/recipe-types"
)

type recipeInput struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Ingredients  []string      `json:"ingredients"`
	Instructions []string      `json:"instructions"`
	Category     string        `json:"category"`
	Image        string        `json:"image"`
	Photos       []types.Photo `json:"photos"`
}

func (in recipeInput) recipe() types.Recipe {
	return types.Recipe{
		Name:         in.Name,
		Description:  in.Description,
		Ingredients:  in.Ingredients,
		Instructions: in.Instructions,
		Category:     in.Category,
		Image:        in.Image,
		Photos:       in.Photos,
	}
}

func (r Runner) cmdList(ctx context.Context, repo RecipeRepo) error {
	recipes, err := repo.GetRecipes(ctx)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(r.stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE")
	for _, rec := range recipes {
		fmt.Fprintf(w, "%s\t%s\n", rec.ID, rec.Name)
	}
	return w.Flush()
}

func (r Runner) cmdExport(ctx context.Context, repo RecipeRepo, id string, includeImageContents bool) error {
	rec, err := repo.GetRecipe(ctx, strings.TrimSpace(id))
	if err != nil {
		return err
	}
	if !includeImageContents {
		stripPhotoContents(&rec)
	}
	return r.writeIndentedJSON(rec)
}

func (r Runner) cmdExportAll(ctx context.Context, repo RecipeRepo, includeImageContents bool) error {
	summaries, err := repo.GetRecipes(ctx)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(r.stdout)
	for _, s := range summaries {
		rec, err := repo.GetRecipe(ctx, s.ID)
		if err != nil {
			return fmt.Errorf("recipe %s: %w", s.ID, err)
		}
		if !includeImageContents {
			stripPhotoContents(&rec)
		}
		if err := enc.Encode(rec); err != nil {
			return err
		}
	}
	return nil
}

// stripPhotoContents clears the base64 image data from each photo while
// preserving the photo IDs and metadata. Used by export commands when the
// caller did not request --image-contents.
func stripPhotoContents(rec *types.Recipe) {
	for i := range rec.Photos {
		rec.Photos[i].ImageBase64 = ""
	}
}

func (r Runner) cmdCreate(ctx context.Context, repo RecipeRepo, path string, returnJSON bool) error {
	var in recipeInput
	if err := r.readJSONObject(path, &in); err != nil {
		return err
	}
	id, err := repo.CreateRecipe(ctx, in.recipe())
	if err != nil {
		return err
	}
	if !returnJSON {
		fmt.Fprintf(r.stdout, "Successfully created recipe %s\n", id)
		return nil
	}
	created, err := repo.GetRecipe(ctx, id)
	if err != nil {
		return err
	}
	stripPhotoContents(&created)
	return r.writeIndentedJSON(created)
}

func (r Runner) cmdPatch(ctx context.Context, repo RecipeRepo, id string, path string, returnJSON bool) error {
	id = strings.TrimSpace(id)
	var patch recipePatch
	if err := r.readJSONObject(path, &patch); err != nil {
		return err
	}
	if !patch.anySet() {
		return errors.New("no fields to update")
	}
	cur, err := repo.GetRecipe(ctx, id)
	if err != nil {
		return err
	}
	merged := mergeRecipePatch(cur, patch)
	merged.ID = id
	if err := repo.UpdateRecipe(ctx, merged); err != nil {
		return err
	}
	if !returnJSON {
		fmt.Fprintf(r.stdout, "Successfully updated recipe %s (fields: %s)\n", id, strings.Join(patch.setFields(), ", "))
		return nil
	}
	updated, err := repo.GetRecipe(ctx, id)
	if err != nil {
		return err
	}
	stripPhotoContents(&updated)
	return r.writeIndentedJSON(updated)
}

func (r Runner) cmdDelete(ctx context.Context, repo RecipeRepo, id string) error {
	id = strings.TrimSpace(id)
	if err := repo.DeleteRecipe(ctx, id); err != nil {
		return err
	}
	fmt.Fprintf(r.stdout, "Successfully deleted recipe %s\n", id)
	return nil
}

func (r Runner) cmdImport(ctx context.Context, repo RecipeRepo, path string) error {
	in, closeInput, err := r.openInput(path)
	if err != nil {
		return err
	}
	defer closeInput()

	sc := bufio.NewScanner(in)
	// Default buffer may be too small for long JSON lines.
	const max = 64 * 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, max)

	lineNum := 0
	imported := 0
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
		if err := repo.ImportRecipe(ctx, rec); err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
		}
		imported++
	}
	if err := sc.Err(); err != nil {
		return err
	}
	fmt.Fprintf(r.stdout, "Successfully imported %d recipes\n", imported)
	return nil
}
