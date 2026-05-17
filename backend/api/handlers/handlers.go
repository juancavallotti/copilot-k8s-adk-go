package handlers

import (
	"github.com/gin-gonic/gin"
	repo "juancavallotti.com/recipes-repo"
)

// Handlers exposes recipe HTTP endpoints backed by Repo.
type Handlers struct {
	Repo *repo.Repo
}

// New constructs HTTP handlers for the given repository.
func New(r *repo.Repo) *Handlers {
	return &Handlers{Repo: r}
}

// Register mounts recipe CRUD routes on r (typically *gin.Engine or a group).
func (h *Handlers) Register(r gin.IRoutes) {
	r.GET("/livez", h.Liveness)
	r.GET("/readyz", h.Readiness)
	r.GET("/recipes", h.ListRecipes)
	r.GET("/recipes/:id", h.GetRecipe)
	r.POST("/recipes", h.CreateRecipe)
	r.PUT("/recipes/:id", h.ReplaceRecipe)
	r.PATCH("/recipes/:id", h.PatchRecipe)
	r.DELETE("/recipes/:id", h.DeleteRecipe)
}
