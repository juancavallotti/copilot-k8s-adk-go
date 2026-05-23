package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type skillCreatePayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

type skillUpdatePayload struct {
	Description *string `json:"description"`
	Content     *string `json:"content"`
}

// ListSkills handles GET /skills.
func (h *Handlers) ListSkills(c *gin.Context) {
	skills, err := h.Repo.ListSkills(c.Request.Context())
	if err != nil {
		writeRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, skills)
}

// GetSkill handles GET /skills/:id.
func (h *Handlers) GetSkill(c *gin.Context) {
	skill, err := h.Repo.GetSkill(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, skill)
}

// CreateSkill handles POST /skills.
func (h *Handlers) CreateSkill(c *gin.Context) {
	var body skillCreatePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		writeBindErr(c, err)
		return
	}
	id, err := h.Repo.CreateSkill(c.Request.Context(), body.Name, body.Description, body.Content)
	if err != nil {
		writeRepoErr(c, err)
		return
	}
	created, err := h.Repo.GetSkill(c.Request.Context(), id)
	if err != nil {
		writeRepoErr(c, err)
		return
	}
	c.Header("Location", "/skills/"+id)
	c.JSON(http.StatusCreated, created)
}

// PatchSkill handles PATCH /skills/:id. Updates description and/or content.
// The name is intentionally immutable to keep agent references stable.
func (h *Handlers) PatchSkill(c *gin.Context) {
	id := c.Param("id")
	var body skillUpdatePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		writeBindErr(c, err)
		return
	}
	if body.Description == nil && body.Content == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}
	cur, err := h.Repo.GetSkill(c.Request.Context(), id)
	if err != nil {
		writeRepoErr(c, err)
		return
	}
	description := cur.Description
	if body.Description != nil {
		description = *body.Description
	}
	content := cur.Content
	if body.Content != nil {
		content = *body.Content
	}
	if err := h.Repo.UpdateSkill(c.Request.Context(), id, description, content); err != nil {
		writeRepoErr(c, err)
		return
	}
	out, err := h.Repo.GetSkill(c.Request.Context(), id)
	if err != nil {
		writeRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// DeleteSkill handles DELETE /skills/:id.
func (h *Handlers) DeleteSkill(c *gin.Context) {
	if err := h.Repo.DeleteSkill(c.Request.Context(), c.Param("id")); err != nil {
		writeRepoErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
