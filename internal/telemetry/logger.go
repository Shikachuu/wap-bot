package telemetry

import (
	"log/slog"
	"os"
)

// SetupLogger creates a new structured slog logger and sets in on the global slog context
//
// inDebug defines the log level, if true the level is debug, otherwise it's info.
func SetupLogger(inDebug bool) {
	level := slog.LevelInfo
	if inDebug {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}))

	slog.SetDefault(logger)
}
