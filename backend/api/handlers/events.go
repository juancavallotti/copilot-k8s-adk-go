package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListEvents handles GET /events?limit=&offset=
func (h *Handlers) ListEvents(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	events, err := h.Repo.ListEvents(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

// ListEventTraces handles GET /events/:event_id/traces?limit=&offset=
func (h *Handlers) ListEventTraces(c *gin.Context) {
	eventID := c.Param("event_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	traces, err := h.Repo.ListTracesByEvent(c.Request.Context(), eventID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, traces)
}
