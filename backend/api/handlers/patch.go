package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	types "juancavallotti.com/recipe-types"
)

// recipePatch is a partial update payload. Omitted JSON fields are nil and left unchanged.
type recipePatch struct {
	Name         *string   `json:"name"`
	Description  *string   `json:"description"`
	Category     *string   `json:"category"`
	Image        *string   `json:"image"`
	Ingredients  *[]string `json:"ingredients"`
	Instructions *[]string `json:"instructions"`
}

func (p recipePatch) anySet() bool {
	return p.Name != nil || p.Description != nil || p.Category != nil || p.Image != nil ||
		p.Ingredients != nil || p.Instructions != nil
}

func mergeRecipePatch(cur types.Recipe, p recipePatch) types.Recipe {
	out := cur
	if p.Name != nil {
		out.Name = *p.Name
	}
	if p.Description != nil {
		out.Description = *p.Description
	}
	if p.Category != nil {
		out.Category = *p.Category
	}
	if p.Image != nil {
		out.Image = *p.Image
	}
	if p.Ingredients != nil {
		out.Ingredients = *p.Ingredients
	}
	if p.Instructions != nil {
		out.Instructions = *p.Instructions
	}
	return out
}

// PatchRecipe handles PATCH /recipes/:id.
func (h *Handlers) PatchRecipe(c *gin.Context) {
	id := c.Param("id")
	var patch recipePatch
	if err := c.ShouldBindJSON(&patch); err != nil {
		writeBindErr(c, err)
		return
	}
	if !patch.anySet() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	cur, err := h.Repo.GetRecipe(c.Request.Context(), id)
	if err != nil {
		writeRepoErr(c, err)
		return
	}
	merged := mergeRecipePatch(cur, patch)
	merged.ID = id

	if err := h.Repo.UpdateRecipe(c.Request.Context(), merged); err != nil {
		writeRepoErr(c, err)
		return
	}
	out, err := h.Repo.GetRecipe(c.Request.Context(), id)
	if err != nil {
		writeRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}
