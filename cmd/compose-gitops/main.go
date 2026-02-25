package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/michael/dev/compose-gitops/internal/compose"
	"github.com/michael/dev/compose-gitops/internal/config"
	"github.com/michael/dev/compose-gitops/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runner, err := compose.NewComposeRunner()
	if err != nil {
		slog.Error("compose runner", "error", err)
		os.Exit(1)
	}
	srv := server.New(cfg, runner)
	if err := srv.Run(ctx); err != nil && err != context.Canceled {
		slog.Error("server exited", "error", err)
		os.Exit(1)
	}
}
