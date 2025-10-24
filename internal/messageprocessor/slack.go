package messageprocessor

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/Shikachuu/wap-bot/pkg/musicextractors"
	"github.com/slack-go/slack"
)

type ParsedMusicLink struct {
	Title string
	URL   string
	Type  musicextractors.ExtractProvider
}

type SlackMessageProcessor interface {
	SummarizeThread(channelID, threadTS string)
	ExtractMusicURL(text string) (ParsedMusicLink, error)
}

type slackMessageProcessor struct {
	client      *slack.Client
	processors  map[musicextractors.ExtractProvider]musicextractors.MusicURLExtractorFunc
	titleParser map[musicextractors.ExtractProvider]musicextractors.TitleExtractorFunc
}

var _ SlackMessageProcessor = (*slackMessageProcessor)(nil)

func (s *slackMessageProcessor) ExtractMusicURL(text string) (ParsedMusicLink, error) {
	for _, process := range s.processors {
		url, p, err := process(text)
		if err != nil {
			if errors.Is(err, musicextractors.ErrNoURLFound) {
				continue
			}

			return ParsedMusicLink{}, fmt.Errorf("url parsing: %w", err)
		}

		title, err := s.titleParser[p](url)
		if err != nil {
			return ParsedMusicLink{}, fmt.Errorf("title parsing: %w", err)
		}

		return ParsedMusicLink{
			Title: title,
			URL:   url,
			Type:  p,
		}, nil
	}

	return ParsedMusicLink{}, musicextractors.ErrNoURLFound
}

func (s *slackMessageProcessor) SummarizeThread(channelID string, threadTS string) {
	logger := slog.With("channel_id", channelID, "thread_ts", threadTS)

	logger.Debug("processing thread")

	params := &slack.GetConversationRepliesParameters{
		ChannelID: channelID,
		Timestamp: threadTS,
		Limit:     1000,
	}

	pmls := []ParsedMusicLink{}

	msgs, _, _, err := s.client.GetConversationReplies(params)
	if err != nil {
		logger.Error("failed to get slack replies", "error", err)
		return
	}

	for _, msg := range msgs {
		m, err := s.ExtractMusicURL(msg.Text)
		if err != nil {
			logger.Warn("unable to process url in reply", "text", msg.Text, "username", msg.Username)
			continue
		}

		pmls = append(pmls, m)
	}

	csv, size, err := s.createCSV(pmls)
	if err != nil {
		logger.Error("failed to generate csv file", "error", err)
		return
	}

	_, err = s.client.UploadFileV2(slack.UploadFileV2Parameters{
		Reader:          csv,
		Filename:        channelID + "-" + threadTS + ".csv",
		Title:           "Music URLs from Thread",
		InitialComment:  fmt.Sprintf("Found %d music URLs in this thread", len(pmls)),
		Channel:         channelID,
		ThreadTimestamp: threadTS,
		FileSize:        size,
	})
	if err != nil {
		logger.Error("failed to post csv file to thread", "error", err, "api", "v2")
		return
	}

	logger.Info("summarized thread", "count", len(pmls))
}

func (s *slackMessageProcessor) createCSV(pmls []ParsedMusicLink) (io.Reader, int, error) {
	buff := bytes.NewBuffer(nil)
	w := csv.NewWriter(buff)
	w.Comma = ';'

	w.Write([]string{"Title", "Spotify URL", "YouTube URL", "YouTube Music URL"})

	for _, pml := range pmls {
		switch pml.Type {
		case musicextractors.SpotifyProvider:
			w.Write([]string{pml.Title, pml.URL, "", ""})
		case musicextractors.YouTubeProvider:
			w.Write([]string{pml.Title, "", pml.URL, ""})
		case musicextractors.YoutTubeMusicProvider:
			w.Write([]string{pml.Title, "", "", pml.URL})
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return nil, 0, err
	}

	return bytes.NewReader(buff.Bytes()), buff.Cap(), nil
}

func NewSlackMessageProcessor(
	client *slack.Client,
	urlP map[musicextractors.ExtractProvider]musicextractors.MusicURLExtractorFunc,
	tp map[musicextractors.ExtractProvider]musicextractors.TitleExtractorFunc,
) *slackMessageProcessor {
	return &slackMessageProcessor{
		client:      client,
		processors:  urlP,
		titleParser: tp,
	}
}
