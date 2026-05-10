package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DeleteRecipe handles DELETE /recipes/:id.
func (h *Handlers) DeleteRecipe(c *gin.Context) {
	id := c.Param("id")
	if err := h.Repo.DeleteRecipe(c.Request.Context(), id); err != nil {
		writeRepoErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
