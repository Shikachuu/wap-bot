package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"

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
	isDebug := inDebugMode()

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
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	botToken, appToken, err := getConfig()
	if err != nil {
		slog.Error("failed to get configuration", "error", err)
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

		if err := client.RunContext(ctx); err != nil {
			slog.Error("slack client error", "error", err)
		}
	}()

	// Wait for shutdown signal
	<-sigCh
	slog.Info("shutdown signal received, gracefully shutting down...")
	cancel()

	slog.Info("shutdown complete")
}

func inDebugMode() bool {
	debugEnabledOptions := []string{"1", "true", "enable"}

	if slices.Contains(debugEnabledOptions, strings.ToLower(os.Getenv("DEBUG"))) {
		return true
	}

	return false
}

func getConfig() (string, string, error) {
	var (
		botToken = os.Getenv("SLACK_BOT_TOKEN")
		appToken = os.Getenv("SLACK_APP_TOKEN")
	)

	if botToken == "" {
		return "", "", errors.New("SLACK_BOT_TOKEN variable is required")
	}

	if !strings.HasPrefix(botToken, "xoxb-") {
		return "", "", errors.New(`SLACK_BOT_TOKEN variable must have a prefix "xoxb-"`)
	}

	if !strings.HasPrefix(appToken, "xapp-") {
		return "", "", errors.New(`SLACK_APP_TOKEN variable must have a prefix "xapp-"`)
	}

	if appToken == "" {
		return "", "", errors.New("SLACK_APP_TOKEN variable is required")
	}

	return botToken, appToken, nil
}
