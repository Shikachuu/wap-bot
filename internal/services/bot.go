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

type SlackBot struct {
	slackMessageProcessor messageprocessor.SlackMessageProcessor
	socketClient          *socketmode.Client
}

type CommandType string

const CommandSummarize CommandType = "summarize"

var ErrInvalidCommandType = errors.New("invalid command type")

func (bot *SlackBot) HandleEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-bot.socketClient.Events:
			if !ok {
				slog.Info("events channel closed")
				return
			}

			switch evt.Type {
			case socketmode.EventTypeConnecting:
				slog.Debug("connection to slack socket")
			case socketmode.EventTypeConnectionError:
				slog.Warn("socket connection failed")
			case socketmode.EventTypeConnected:
				slog.Info("connected to slack socket")
			case socketmode.EventTypeHello:
				slog.Debug("greeting message received from slack connection")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					slog.Warn("ignored invalid events api data", "event_type", evt.Type)
					continue
				}

				bot.socketClient.Ack(*evt.Request)

				if eventsAPIEvent.Type != slackevents.CallbackEvent {
					continue
				}

				innerEvent := eventsAPIEvent.InnerEvent
				switch ev := innerEvent.Data.(type) {
				case *slackevents.AppMentionEvent:
					if err := bot.HandleMentions(ev); err != nil {
						slog.Error("failed to handle event", "event_type", evt.Type, "error", err)
					}
				default:
					slog.Warn("not implemented events api event received", "event_type", evt.Type, "events_api_event_type", innerEvent.Type)
				}
			default:
				slog.Warn("not implemented event received", "event_type", evt.Type)
			}
		}
	}
}

func (bot *SlackBot) HandleMentions(event *slackevents.AppMentionEvent) error {
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

func NewSlackBot(smp messageprocessor.SlackMessageProcessor, sc *socketmode.Client) *SlackBot {
	return &SlackBot{
		slackMessageProcessor: smp,
		socketClient:          sc,
	}
}
