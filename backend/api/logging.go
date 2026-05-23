package main

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func initSlog(out io.Writer, level string) {
	lvl := parseSlogLevel(level)
	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{Level: lvl})
	slog.SetDefault(slog.New(handler))
	slog.SetLogLoggerLevel(lvl)
}

func parseSlogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func slogRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		attrs := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"route", c.FullPath(),
			"status", status,
			"bytes", c.Writer.Size(),
			"remote", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"duration_ms", time.Since(start).Milliseconds(),
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}

		ctx := c.Request.Context()
		switch {
		case status >= http.StatusInternalServerError:
			slog.ErrorContext(ctx, "http.request", attrs...)
		case status >= http.StatusBadRequest:
			slog.WarnContext(ctx, "http.request", attrs...)
		default:
			slog.InfoContext(ctx, "http.request", attrs...)
		}
	}
}

func slogRecovery() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, recovered any) {
		slog.ErrorContext(c.Request.Context(), "http.panic",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"remote", c.ClientIP(),
			"panic", recovered,
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	})
}
