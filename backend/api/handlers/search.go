package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	repo "juancavallotti.com/recipes-repo"
)

// defaultSearchLimit / maxSearchLimit cap result counts so a stray
// ?limit=10000 doesn't burn API quota.
const (
	defaultSearchLimit = 10
	maxSearchLimit     = 50
)

// SearchRecipes handles GET /search/recipes?q=<text>&limit=N.
func (h *Handlers) SearchRecipes(c *gin.Context) {
	q, limit, ok := parseSearchParams(c)
	if !ok {
		return
	}
	matches, err := h.Repo.SearchRecipes(c.Request.Context(), q, limit)
	if err != nil {
		writeSearchErr(c, err)
		return
	}
	c.JSON(http.StatusOK, matches)
}

// SearchEvents handles GET /search/events?q=<text>&limit=N.
func (h *Handlers) SearchEvents(c *gin.Context) {
	q, limit, ok := parseSearchParams(c)
	if !ok {
		return
	}
	matches, err := h.Repo.SearchEvents(c.Request.Context(), q, limit)
	if err != nil {
		writeSearchErr(c, err)
		return
	}
	c.JSON(http.StatusOK, matches)
}

func parseSearchParams(c *gin.Context) (string, int, bool) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing q"})
		return "", 0, false
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultSearchLimit)))
	if err != nil || limit <= 0 {
		limit = defaultSearchLimit
	}
	if limit > maxSearchLimit {
		limit = maxSearchLimit
	}
	return q, limit, true
}

// writeSearchErr maps an embedding/search error to an HTTP response.
// ErrSearchDisabled gets a 503 so callers can tell "you need to set
// an API key" from "the backend is broken."
func writeSearchErr(c *gin.Context, err error) {
	if errors.Is(err, repo.ErrSearchDisabled) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "search disabled: no embedding API key configured"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
