// Package messageprocessor contains implemented slack message processors
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

type parsedMusicLink struct {
	Title string
	URL   string
	Type  musicextractors.ExtractProvider
}

// SlackMessageProcessor contains the core business logic to iterate over a thread and pull every implemented music related info from them.
type SlackMessageProcessor interface {
	SummarizeThread(channelID, threadTS string)
}

type slackMessageProcessor struct {
	client      *slack.Client
	processors  map[musicextractors.ExtractProvider]musicextractors.MusicURLExtractorFunc
	titleParser map[musicextractors.ExtractProvider]musicextractors.TitleExtractorFunc
}

var _ SlackMessageProcessor = (*slackMessageProcessor)(nil)

func (s *slackMessageProcessor) extractMusicURL(text string) (parsedMusicLink, error) {
	for _, process := range s.processors {
		url, p, err := process(text)
		if err != nil {
			if errors.Is(err, musicextractors.ErrNoURLFound) {
				continue
			}

			return parsedMusicLink{}, fmt.Errorf("url parsing: %w", err)
		}

		title, err := s.titleParser[p](url)
		if err != nil {
			return parsedMusicLink{}, fmt.Errorf("title parsing: %w", err)
		}

		return parsedMusicLink{
			Title: title,
			URL:   url,
			Type:  p,
		}, nil
	}

	return parsedMusicLink{}, musicextractors.ErrNoURLFound
}

// SummarizeThread iterates over every message on a thread with a given channelID and threadTS and creates a summarized response.
func (s *slackMessageProcessor) SummarizeThread(channelID, threadTS string) {
	logger := slog.With("channel_id", channelID, "thread_ts", threadTS)

	logger.Debug("processing thread")

	params := &slack.GetConversationRepliesParameters{
		ChannelID: channelID,
		Timestamp: threadTS,
		Limit:     1000,
	}

	pmls := []parsedMusicLink{}

	msgs, _, _, err := s.client.GetConversationReplies(params)
	if err != nil {
		logger.Error("failed to get slack replies", "error", err)
		return
	}

	for i := range msgs {
		m, eErr := s.extractMusicURL(msgs[i].Text)
		if eErr != nil {
			logger.Warn("unable to process url in reply", "text", msgs[i].Text, "username", msgs[i].Username)
			continue
		}

		pmls = append(pmls, m)
	}

	csvF, size, err := s.createCSV(pmls)
	if err != nil {
		logger.Error("failed to generate csv file", "error", err)
		return
	}

	fileName := fmt.Sprintf("%s-%s.csv", channelID, threadTS)

	_, err = s.client.UploadFileV2(slack.UploadFileV2Parameters{
		Reader:          csvF,
		Filename:        fileName,
		Title:           fileName,
		InitialComment:  fmt.Sprintf("Found %d music URLs in this thread", len(pmls)),
		Channel:         channelID,
		ThreadTimestamp: threadTS,
		FileSize:        size,
	})
	if err != nil {
		logger.Error("failed to post csv file to thread", "error", err)
		return
	}

	logger.Info("summarized thread", "count", len(pmls))
}

func (s *slackMessageProcessor) createCSV(pmls []parsedMusicLink) (io.Reader, int, error) {
	buff := bytes.NewBuffer(nil)
	w := csv.NewWriter(buff)
	w.Comma = ';'

	err := w.Write([]string{"Title", "Spotify URL", "YouTube URL", "YouTube Music URL"})
	if err != nil {
		return nil, 0, fmt.Errorf("appending csv line: %w", err)
	}

	for _, pml := range pmls {
		var lErr error

		switch pml.Type {
		case musicextractors.SpotifyProvider:
			lErr = w.Write([]string{pml.Title, pml.URL, "", ""})
		case musicextractors.YouTubeProvider:
			lErr = w.Write([]string{pml.Title, "", pml.URL, ""})
		case musicextractors.YoutTubeMusicProvider:
			lErr = w.Write([]string{pml.Title, "", "", pml.URL})
		}

		if lErr != nil {
			return nil, 0, fmt.Errorf("appending csv line: %w", err)
		}
	}

	w.Flush()

	if err = w.Error(); err != nil {
		return nil, 0, fmt.Errorf("flushing csv buffer: %w", err)
	}

	return bytes.NewReader(buff.Bytes()), buff.Len(), nil
}

// NewSlackMessageProcessor creates a new processor with the given url and title extractors.
func NewSlackMessageProcessor(
	client *slack.Client,
	urlP map[musicextractors.ExtractProvider]musicextractors.MusicURLExtractorFunc,
	tp map[musicextractors.ExtractProvider]musicextractors.TitleExtractorFunc,
) SlackMessageProcessor {
	return &slackMessageProcessor{
		client:      client,
		processors:  urlP,
		titleParser: tp,
	}
}
