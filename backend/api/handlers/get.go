package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListRecipes handles GET /recipes.
func (h *Handlers) ListRecipes(c *gin.Context) {
	recipes, err := h.Repo.GetRecipes(c.Request.Context())
	if err != nil {
		writeRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, recipes)
}

// GetRecipe handles GET /recipes/:id.
func (h *Handlers) GetRecipe(c *gin.Context) {
	id := c.Param("id")
	recipe, err := h.Repo.GetRecipe(c.Request.Context(), id)
	if err != nil {
		writeRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, recipe)
}
