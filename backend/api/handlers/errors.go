package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	repo "juancavallotti.com/recipes-repo"
)

func writeRepoErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repo.ErrRecipeNotFound),
		errors.Is(err, repo.ErrPhotoNotFound),
		errors.Is(err, repo.ErrEventNotFound),
		errors.Is(err, repo.ErrSkillNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, repo.ErrSkillNameTaken):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, repo.ErrInvalidID),
		errors.Is(err, repo.ErrInvalidRecipe),
		errors.Is(err, repo.ErrInvalidRecipeID),
		errors.Is(err, repo.ErrParseIngredient),
		errors.Is(err, repo.ErrInvalidSkillID),
		errors.Is(err, repo.ErrInvalidSkillName),
		errors.Is(err, repo.ErrInvalidSkillDescription),
		errors.Is(err, repo.ErrInvalidSkillContent):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		slog.ErrorContext(c.Request.Context(), "api.repo_error",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"err", err,
		)
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

func writeBindErr(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
