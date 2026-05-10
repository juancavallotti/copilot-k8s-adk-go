package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	types "juancavallotti.com/recipe-types"
)

// ReplaceRecipe handles PUT /recipes/:id (full replacement of mutable fields and lines).
func (h *Handlers) ReplaceRecipe(c *gin.Context) {
	id := c.Param("id")
	var body types.Recipe
	if err := c.ShouldBindJSON(&body); err != nil {
		writeBindErr(c, err)
		return
	}
	body.ID = id

	if err := h.Repo.UpdateRecipe(c.Request.Context(), body); err != nil {
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
