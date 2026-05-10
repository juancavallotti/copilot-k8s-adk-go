package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	types "juancavallotti.com/recipe-types"
)

// CreateRecipe handles POST /recipes.
func (h *Handlers) CreateRecipe(c *gin.Context) {
	var body types.Recipe
	if err := c.ShouldBindJSON(&body); err != nil {
		writeBindErr(c, err)
		return
	}
	body.ID = ""

	id, err := h.Repo.CreateRecipe(c.Request.Context(), body)
	if err != nil {
		writeRepoErr(c, err)
		return
	}
	created, err := h.Repo.GetRecipe(c.Request.Context(), id)
	if err != nil {
		writeRepoErr(c, err)
		return
	}
	c.Header("Location", "/recipes/"+id)
	c.JSON(http.StatusCreated, created)
}
