package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	types "juancavallotti.com/recipe-types"
)

const importMaxBytes = 256 << 20 // 256 MiB

type backupManifest struct {
	Version     int       `json:"version"`
	ExportedAt  time.Time `json:"exported_at"`
	RecipeCount int       `json:"recipe_count"`
}

type importResult struct {
	Imported int               `json:"imported"`
	Failed   []importFailEntry `json:"failed"`
}

type importFailEntry struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

// ExportRecipesZip handles GET /recipes/export. Streams a zip whose layout is:
//
//	manifest.json
//	recipes/<recipe-id>/recipe.json         (Photos[].ImageBase64 cleared)
//	recipes/<recipe-id>/photos/<photo-id>.<ext>
func (h *Handlers) ExportRecipesZip(c *gin.Context) {
	ctx := c.Request.Context()

	summaries, err := h.Repo.GetRecipes(ctx)
	if err != nil {
		writeRepoErr(c, err)
		return
	}

	filename := fmt.Sprintf("recipes-backup-%s.zip", time.Now().UTC().Format("2006-01-02"))
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	zw := zip.NewWriter(c.Writer)
	defer zw.Close()

	for _, s := range summaries {
		rec, err := h.Repo.GetRecipe(ctx, s.ID)
		if err != nil {
			_ = c.Error(err)
			return
		}
		if err := writeRecipeToZip(zw, rec); err != nil {
			_ = c.Error(err)
			return
		}
	}

	manifest := backupManifest{
		Version:     1,
		ExportedAt:  time.Now().UTC(),
		RecipeCount: len(summaries),
	}
	if err := writeJSONEntry(zw, "manifest.json", manifest); err != nil {
		_ = c.Error(err)
		return
	}
}

func writeRecipeToZip(zw *zip.Writer, rec types.Recipe) error {
	base := path.Join("recipes", rec.ID)

	sanitized := rec
	sanitized.Photos = make([]types.Photo, len(rec.Photos))
	for i, p := range rec.Photos {
		sanitized.Photos[i] = types.Photo{ID: p.ID, Featured: p.Featured}
	}
	if err := writeJSONEntry(zw, path.Join(base, "recipe.json"), sanitized); err != nil {
		return err
	}

	for _, p := range rec.Photos {
		if p.ImageBase64 == "" {
			continue
		}
		raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(p.ImageBase64))
		if err != nil {
			return fmt.Errorf("recipe %s photo %s: decode base64: %w", rec.ID, p.ID, err)
		}
		ext := extForBytes(raw)
		photoID := p.ID
		if photoID == "" {
			photoID = fmt.Sprintf("photo-%d", time.Now().UnixNano())
		}
		w, err := zw.Create(path.Join(base, "photos", photoID+ext))
		if err != nil {
			return err
		}
		if _, err := w.Write(raw); err != nil {
			return err
		}
	}
	return nil
}

func writeJSONEntry(zw *zip.Writer, name string, v any) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func extForBytes(b []byte) string {
	ct := http.DetectContentType(b)
	switch {
	case strings.HasPrefix(ct, "image/png"):
		return ".png"
	case strings.HasPrefix(ct, "image/jpeg"):
		return ".jpg"
	case strings.HasPrefix(ct, "image/gif"):
		return ".gif"
	case strings.HasPrefix(ct, "image/webp"):
		return ".webp"
	default:
		return ".bin"
	}
}

// ImportRecipesZip handles POST /recipes/import. Accepts a multipart upload
// with a single "file" field containing a zip in the export layout. Each
// recipe.json is decoded, its photos[] are re-hydrated from sibling
// photos/<photo-id>.<ext> files, and the recipe is upserted via ImportRecipe.
func (h *Handlers) ImportRecipesZip(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, importMaxBytes)

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("missing or invalid 'file' field: %v", err)})
		return
	}
	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	zr, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid zip: %v", err)})
		return
	}

	// photoFiles is keyed by "<recipe-id>/<photo-id>" (no extension).
	photoFiles := make(map[string]*zip.File)
	var recipeFiles []*zip.File
	for _, zf := range zr.File {
		name := zf.Name
		if zf.FileInfo().IsDir() {
			continue
		}
		if strings.HasPrefix(name, "recipes/") && strings.HasSuffix(name, "/recipe.json") {
			recipeFiles = append(recipeFiles, zf)
			continue
		}
		// Match recipes/<id>/photos/<photo-id>.<ext>
		if strings.HasPrefix(name, "recipes/") && strings.Contains(name, "/photos/") {
			rest := strings.TrimPrefix(name, "recipes/")
			parts := strings.SplitN(rest, "/photos/", 2)
			if len(parts) != 2 {
				continue
			}
			recipeID := parts[0]
			photoFile := parts[1]
			photoID := strings.TrimSuffix(photoFile, path.Ext(photoFile))
			photoFiles[recipeID+"/"+photoID] = zf
		}
	}

	result := importResult{Failed: []importFailEntry{}}
	ctx := c.Request.Context()
	for _, rf := range recipeFiles {
		recipeID := recipeIDFromPath(rf.Name)
		rec, err := readRecipeFromZip(rf, recipeID, photoFiles)
		if err != nil {
			result.Failed = append(result.Failed, importFailEntry{ID: recipeID, Error: err.Error()})
			continue
		}
		if err := h.Repo.ImportRecipe(ctx, rec); err != nil {
			result.Failed = append(result.Failed, importFailEntry{ID: recipeID, Error: err.Error()})
			continue
		}
		result.Imported++
	}

	c.JSON(http.StatusOK, result)
}

func recipeIDFromPath(name string) string {
	rest := strings.TrimPrefix(name, "recipes/")
	return strings.TrimSuffix(rest, "/recipe.json")
}

func readRecipeFromZip(rf *zip.File, recipeID string, photoFiles map[string]*zip.File) (types.Recipe, error) {
	rc, err := rf.Open()
	if err != nil {
		return types.Recipe{}, err
	}
	defer rc.Close()

	var rec types.Recipe
	if err := json.NewDecoder(rc).Decode(&rec); err != nil {
		return types.Recipe{}, fmt.Errorf("decode recipe.json: %w", err)
	}

	for i, p := range rec.Photos {
		if p.ID == "" {
			continue
		}
		pf, ok := photoFiles[recipeID+"/"+p.ID]
		if !ok {
			continue
		}
		bytesData, err := readZipFile(pf)
		if err != nil {
			return types.Recipe{}, fmt.Errorf("read photo %s: %w", p.ID, err)
		}
		rec.Photos[i].ImageBase64 = base64.StdEncoding.EncodeToString(bytesData)
	}
	return rec, nil
}

func readZipFile(zf *zip.File) ([]byte, error) {
	rc, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}
