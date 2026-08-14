package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/FortiBrine/DriftMind/internal/bot"
	"github.com/FortiBrine/DriftMind/internal/config"
	"github.com/FortiBrine/DriftMind/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("error loading config", "error", err)
		os.Exit(1)
	}

	l := logger.New(cfg.Environment)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	b, err := bot.New(ctx, cfg, l)
	if err != nil {
		l.Error("error creating bot", "error", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err = b.Close(shutdownCtx); err != nil {
			l.Error("error closing bot", "error", err)
		}
	}()

	if err = b.Start(ctx); err != nil {
		l.Error("error starting bot", "error", err)
	}
}
