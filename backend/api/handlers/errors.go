package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	repo "juancavallotti.com/recipes-repo"
)

func writeRepoErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repo.ErrRecipeNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, repo.ErrPhotoNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, repo.ErrInvalidID),
		errors.Is(err, repo.ErrInvalidRecipe),
		errors.Is(err, repo.ErrInvalidRecipeID),
		errors.Is(err, repo.ErrParseIngredient):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

func writeBindErr(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
