package services

import (
	"context"
	"log/slog"
	"strings"

	"github.com/Shikachuu/wap-bot/internal/domain"
	"github.com/Shikachuu/wap-bot/internal/telemetry"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"go.opentelemetry.io/otel/attribute"
)

// SlackBot is the main communication layer of the application,
// contains and handles socket connections and sync Slack API calls.
type SlackBot struct {
	slackMessageProcessor domain.MessageProcessorDomain
	socketClient          *socketmode.Client
}

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

			ctx, t := telemetry.Tracer.Start(ctx, "slackbot.handle_events")
			t.SetAttributes(
				attribute.String("event.type", string(evt.Type)),
			)

			defer t.End()

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
			t.End()
		}
	}
}

func (bot *SlackBot) handleEventsAPI(ctx context.Context, logger *slog.Logger, evt *socketmode.Event) {
	ctx, t := telemetry.Tracer.Start(ctx, "slackbot.handle_events_api")
	defer t.End()

	eventsAPIEvent, isAPIEvent := evt.Data.(slackevents.EventsAPIEvent)
	if !isAPIEvent {
		_ = telemetry.WrapErrorWithTrace(t, "", errIgnoredInvalidAPI)
		logger.WarnContext(ctx, errIgnoredInvalidAPI.Error())
		return
	}

	telemetry.StartEvent(t, telemetry.SendACKEvent)
	bot.socketClient.Ack(*evt.Request)
	telemetry.EndEvent(t, telemetry.SendACKEvent)

	if eventsAPIEvent.Type != slackevents.CallbackEvent {
		t.AddEvent("ignored_non_callback_event")
		return
	}

	innerEvent := eventsAPIEvent.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		telemetry.StartEvent(t, telemetry.HandleMentionsEvent)
		t.SetAttributes(attribute.String("user.id", ev.User), attribute.String("slack.channel_id", ev.Channel))
		if err := bot.handleMentions(ctx, ev); err != nil {
			_ = telemetry.WrapErrorWithTrace(t, "", errHandleEvent)
			logger.ErrorContext(ctx, errHandleEvent.Error(), "error", err)
		}
		telemetry.EndEvent(t, telemetry.HandleMentionsEvent)
	default:
		_ = telemetry.WrapErrorWithTrace(t, "", errHandleEvent)
		logger.WarnContext(ctx, errNotImplementedEvent.Error(), "events_api_event_type", innerEvent.Type)
	}
}

func (bot *SlackBot) handleMentions(ctx context.Context, event *slackevents.AppMentionEvent) error {
	ctx, t := telemetry.Tracer.Start(ctx, "slackbot.handle_mentions")
	defer t.End()

	if event.ThreadTimeStamp == "" {
		telemetry.StartEvent(t, telemetry.NonThreadPostEphemeralEvent)

		_, err := bot.socketClient.PostEphemeralContext(
			ctx,
			event.Channel,
			event.User,
			slack.MsgOptionText("Bot is only usable in threads to summarize them", false),
		)

		telemetry.EndEvent(t, telemetry.NonThreadPostEphemeralEvent)

		if err != nil {
			return telemetry.WrapErrorWithTrace(t, "unable to post ephemeral notification", err)
		}

		return nil
	}

	switch {
	case strings.Contains(event.Text, string(CommandSummarize)):
		err := bot.processThread(ctx, event.Channel, event.ThreadTimeStamp)
		if err != nil {
			return telemetry.WrapErrorWithTrace(t, "processing thread", err)
		}

	default:
		return telemetry.WrapErrorWithTrace(t, "parsing command", ErrInvalidCommandType)
	}

	return nil
}

func (bot *SlackBot) processThread(ctx context.Context, channelID, threadTS string) error {
	ctx, t := telemetry.Tracer.Start(ctx, "slackbot.process_thread")
	defer t.End()

	t.SetAttributes(
		attribute.String("slack.channel_id", channelID),
		attribute.String("slack.thread_ts", threadTS),
	)

	logger := slog.With("channel_id", channelID, "thread_ts", threadTS)

	logger.DebugContext(ctx, "processing thread")

	telemetry.StartEvent(t, telemetry.GetConversationRepliesEvent)
	msgs, _, _, err := bot.socketClient.GetConversationRepliesContext(
		ctx,
		&slack.GetConversationRepliesParameters{
			ChannelID: channelID,
			Timestamp: threadTS,
			Limit:     1000,
		},
	)
	telemetry.EndEvent(t, telemetry.GetConversationRepliesEvent)

	if err != nil {
		return telemetry.WrapErrorWithTrace(t, "get slack thread replies", err)
	}

	telemetry.StartEvent(t, telemetry.SummarizeThreadEvent)
	t.SetAttributes(attribute.Int("slack.message_count", len(msgs)))
	reply, err := bot.slackMessageProcessor.SummarizeThread(msgs, channelID, threadTS)
	telemetry.EndEvent(t, telemetry.SummarizeThreadEvent)

	if err != nil {
		return telemetry.WrapErrorWithTrace(t, "summarizing thread", err)
	}

	t.SetAttributes(attribute.Int("file.size", reply.FileSize), attribute.String("file.name", reply.Filename))

	telemetry.StartEvent(t, telemetry.UploadFileV2Event)
	_, err = bot.socketClient.UploadFileV2(reply)
	telemetry.EndEvent(t, telemetry.UploadFileV2Event)

	if err != nil {
		return telemetry.WrapErrorWithTrace(t, "uploading file to reply", err)
	}

	logger.InfoContext(ctx, "summarized thread")

	return nil
}

// NewSlackBot creates a new slack bot with the given message processor and socket client.
func NewSlackBot(smp domain.MessageProcessorDomain, sc *socketmode.Client) *SlackBot {
	return &SlackBot{
		slackMessageProcessor: smp,
		socketClient:          sc,
	}
}
