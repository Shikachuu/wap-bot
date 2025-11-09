// Package main
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shikachuu/wap-bot/internal/config"
	"github.com/Shikachuu/wap-bot/internal/domain"
	"github.com/Shikachuu/wap-bot/internal/services"
	"github.com/Shikachuu/wap-bot/internal/telemetry"
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
	ctx, cancel := context.WithCancel(context.Background())
	if err := run(ctx, cancel); err != nil {
		slog.Error("failed to run server", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()

	inDebug := config.InDebugMode()

	telemetry.SetupLogger(inDebug)

	tShutdown, err := telemetry.SetupOTel(ctx)
	if err != nil {
		return fmt.Errorf("setting up otel: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	botToken, appToken, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	api := slack.New(
		botToken,
		slack.OptionAppLevelToken(appToken),
		slack.OptionDebug(inDebug),
	)

	client := socketmode.New(api)

	smp := domain.NewSlackMessageProcessor(urlProcessors, titleExtractors)

	sb := services.NewSlackBot(smp, client)

	slog.InfoContext(ctx, "starting event handler...")

	go sb.HandleEvents(ctx)

	go func() {
		slog.Info("starting slack socket connection...")

		if rErr := client.RunContext(ctx); rErr != nil {
			slog.Error("slack client error", "error", rErr)
		}
	}()

	// Wait for shutdown signal
	<-sigCh
	slog.InfoContext(ctx, "shutdown signal received, gracefully shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer shutdownCancel()

	//nolint:contextcheck // we cannot inherit the context here, it canceled above
	if sErr := tShutdown(shutdownCtx); sErr != nil {
		return fmt.Errorf("shutdown otel: %w", sErr)
	}

	slog.InfoContext(ctx, "shutdown complete")

	return nil
}
