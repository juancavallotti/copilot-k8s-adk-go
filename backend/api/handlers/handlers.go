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

// Register mounts API routes on r (typically *gin.Engine or a router group).
func (h *Handlers) Register(r gin.IRouter) {
	r.GET("/livez", h.Liveness)
	r.GET("/readyz", h.Readiness)

	recipes := r.Group("/recipes")
	{
		recipes.GET("", h.ListRecipes)
		recipes.POST("", h.CreateRecipe)
		recipes.GET("/export", h.ExportRecipesZip)
		recipes.POST("/import", h.ImportRecipesZip)
		recipes.GET("/:id", h.GetRecipe)
		recipes.PUT("/:id", h.ReplaceRecipe)
		recipes.PATCH("/:id", h.PatchRecipe)
		recipes.DELETE("/:id", h.DeleteRecipe)
		recipes.POST("/:id/photos", h.AddRecipePhoto)
		recipes.DELETE("/:id/photos/:photo_id", h.DeleteRecipePhoto)
		recipes.PUT("/:id/photos/:photo_id/featured", h.SetFeaturedRecipePhoto)
	}

	events := r.Group("/events")
	{
		events.GET("", h.ListEvents)
		events.DELETE("", h.DeleteAllEvents)
		events.GET("/:event_id/traces", h.ListEventTraces)
		events.DELETE("/:event_id", h.DeleteEvent)
	}

	skills := r.Group("/skills")
	{
		skills.GET("", h.ListSkills)
		skills.POST("", h.CreateSkill)
		skills.GET("/:id", h.GetSkill)
		skills.PATCH("/:id", h.PatchSkill)
		skills.DELETE("/:id", h.DeleteSkill)
	}
}
