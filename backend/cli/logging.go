package main

import (
	"io"
	"log/slog"
	"strings"
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
