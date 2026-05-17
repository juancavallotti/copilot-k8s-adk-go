package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	types "juancavallotti.com/recipe-types"
)

type photoPayload struct {
	ImageBase64 string `json:"image_base64"`
	Featured    bool   `json:"featured"`
}

// AddRecipePhoto handles POST /recipes/:id/photos.
func (h *Handlers) AddRecipePhoto(c *gin.Context) {
	var body photoPayload
	if err := c.ShouldBindJSON(&body); err != nil {
		writeBindErr(c, err)
		return
	}

	photo := types.Photo{
		ImageBase64: body.ImageBase64,
		Featured:    body.Featured,
	}
	id, err := h.Repo.AddRecipePhoto(c.Request.Context(), c.Param("id"), photo)
	if err != nil {
		writeRepoErr(c, err)
		return
	}

	created, err := h.Repo.GetRecipe(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeRepoErr(c, err)
		return
	}
	c.Header("Location", "/recipes/"+c.Param("id")+"/photos/"+id)
	c.JSON(http.StatusCreated, created)
}

// DeleteRecipePhoto handles DELETE /recipes/:id/photos/:photo_id.
func (h *Handlers) DeleteRecipePhoto(c *gin.Context) {
	if err := h.Repo.DeleteRecipePhoto(c.Request.Context(), c.Param("id"), c.Param("photo_id")); err != nil {
		writeRepoErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// SetFeaturedRecipePhoto handles PUT /recipes/:id/photos/:photo_id/featured.
func (h *Handlers) SetFeaturedRecipePhoto(c *gin.Context) {
	if err := h.Repo.SetFeaturedRecipePhoto(c.Request.Context(), c.Param("id"), c.Param("photo_id")); err != nil {
		writeRepoErr(c, err)
		return
	}
	out, err := h.Repo.GetRecipe(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}
