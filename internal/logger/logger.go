package logger

import (
	"log/slog"
	"os"

	"github.com/FortiBrine/DriftMind/internal/config"
)

func New(env config.Environment) *slog.Logger {
	level := slog.LevelInfo
	if env.IsDev() {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, new(slog.HandlerOptions{
		Level: level,
	}))

	return slog.New(handler)
}
