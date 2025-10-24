// Package main
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shikachuu/wap-bot/internal/config"
	"github.com/Shikachuu/wap-bot/internal/messageprocessor"
	"github.com/Shikachuu/wap-bot/internal/services"
	"github.com/Shikachuu/wap-bot/pkg/musicextractors"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

var urlProcessors = map[musicextractors.ExtractProvider]musicextractors.MusicURLExtractorFunc{
	musicextractors.SpotifyProvider:       musicextractors.SpotifyURLExtractor,
	musicextractors.YouTubeProvider:       musicextractors.YouTubeURLExtractor,
	musicextractors.YoutTubeMusicProvider: musicextractors.YouTubeMusicURLExtractor,
}

var titleExtractors = map[musicextractors.ExtractProvider]musicextractors.TitleExtractorFunc{
	musicextractors.SpotifyProvider:       musicextractors.SpotifyTitleExtractor,
	musicextractors.YouTubeProvider:       musicextractors.YouTubeTitleExtractor,
	musicextractors.YoutTubeMusicProvider: musicextractors.YouTubeTitleExtractor,
}

func main() {
	isDebug := config.InDebugMode()

	level := slog.LevelInfo
	if isDebug {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}))

	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	botToken, appToken, err := config.GetConfig()
	if err != nil {
		slog.Error("failed to get configuration", "error", err)
		cancel()
		os.Exit(1)
	}

	api := slack.New(
		botToken,
		slack.OptionAppLevelToken(appToken),
		slack.OptionDebug(isDebug),
	)

	client := socketmode.New(api)

	smp := messageprocessor.NewSlackMessageProcessor(api, urlProcessors, titleExtractors)

	sb := services.NewSlackBot(smp, client)

	slog.Info("starting event handler...")

	go sb.HandleEvents(ctx)

	go func() {
		slog.Info("starting slack socket connection...")

		if rErr := client.RunContext(ctx); rErr != nil {
			slog.Error("slack client error", "error", err)
		}
	}()

	// Wait for shutdown signal
	<-sigCh
	slog.Info("shutdown signal received, gracefully shutting down...")
	cancel()

	slog.Info("shutdown complete")
}
