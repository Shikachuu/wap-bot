// Package services contains the available communication layers of the application
package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Shikachuu/wap-bot/internal/messageprocessor"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// SlackBot is the main communication layer of the application, contains and handles socket connections and sync Slack API calls.
type SlackBot struct {
	slackMessageProcessor messageprocessor.SlackMessageProcessor
	socketClient          *socketmode.Client
}

type commandType string

// CommandSummarize is the command that tells handleMentions to run slackMessageProcessor's message handler.
const CommandSummarize commandType = "summarize"

// ErrInvalidCommandType returned by handleMentions in case of an unimplemented CommandType occures.
var ErrInvalidCommandType = errors.New("invalid command type")

// HandleEvents is the main event loop that listens to Slack Socket Events and handles them based on the event's Type field.
func (bot *SlackBot) HandleEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-bot.socketClient.Events:
			if !ok {
				slog.InfoContext(ctx, "events channel closed")
				return
			}

			logger := slog.With("event_type", evt.Type)
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				logger.DebugContext(ctx, "connection to slack socket")
			case socketmode.EventTypeConnectionError:
				logger.WarnContext(ctx, "socket connection failed")
			case socketmode.EventTypeConnected:
				logger.InfoContext(ctx, "connected to slack socket")
			case socketmode.EventTypeHello:
				logger.DebugContext(ctx, "greeting message received from slack connection")
			case socketmode.EventTypeEventsAPI:
				bot.handleEventsAPI(ctx, logger, &evt)
			default:
				logger.WarnContext(ctx, "not implemented event received")
			}
		}
	}
}

func (bot *SlackBot) handleEventsAPI(ctx context.Context, logger *slog.Logger, evt *socketmode.Event) {
	eventsAPIEvent, isAPIEvent := evt.Data.(slackevents.EventsAPIEvent)
	if !isAPIEvent {
		logger.WarnContext(ctx, "ignored invalid events api data")
		return
	}

	bot.socketClient.Ack(*evt.Request)

	if eventsAPIEvent.Type != slackevents.CallbackEvent {
		return
	}

	innerEvent := eventsAPIEvent.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		if err := bot.handleMentions(ev); err != nil {
			logger.ErrorContext(ctx, "failed to handle event", "error", err)
		}
	default:
		logger.WarnContext(ctx, "not implemented events api event received", "events_api_event_type", innerEvent.Type)
	}
}

func (bot *SlackBot) handleMentions(event *slackevents.AppMentionEvent) error {
	if event.ThreadTimeStamp == "" {
		_, err := bot.socketClient.PostEphemeral(
			event.Channel,
			event.User,
			slack.MsgOptionText("Bot is only usable in threads to summarize them", false),
		)
		if err != nil {
			return fmt.Errorf("unable to post ephemeral notification text to user: %w", err)
		}

		return nil
	}

	switch {
	case strings.Contains(event.Text, string(CommandSummarize)):
		bot.slackMessageProcessor.SummarizeThread(event.Channel, event.ThreadTimeStamp)
	default:
		return ErrInvalidCommandType
	}

	return nil
}

// NewSlackBot creates a new slack bot with the given message processor and socket client.
func NewSlackBot(smp messageprocessor.SlackMessageProcessor, sc *socketmode.Client) *SlackBot {
	return &SlackBot{
		slackMessageProcessor: smp,
		socketClient:          sc,
	}
}
