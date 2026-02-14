package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rhuss/readwise-mcp-server/internal/server"
	"github.com/rhuss/readwise-mcp-server/internal/types"
)

func main() {
	cfg := types.LoadConfig()

	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	if err := cfg.ValidateTLS(); err != nil {
		logger.Error("invalid TLS configuration", "error", err)
		os.Exit(1)
	}

	srv, err := server.New(cfg, logger)
	if err != nil {
		logger.Error("failed to create server", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if err := srv.ListenAndServe(ctx); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
