// Package domain contains implemented slack message processors, contains the main business logic
package domain

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"github.com/Shikachuu/wap-bot/pkg/musicextractors"
	"github.com/slack-go/slack"
)

type parsedMusicLink struct {
	Title string
	URL   string
	Type  musicextractors.ExtractProvider
}

// MessageProcessorDomain contains the core business logic to iterate over a thread and pull every implemented music related info from them.
type MessageProcessorDomain interface {
	SummarizeThread(msgs []slack.Message, channelID, threadTS string) (slack.UploadFileV2Parameters, error)
}

type messageProcessorDomain struct {
	processors  map[musicextractors.ExtractProvider]musicextractors.MusicURLExtractorFunc
	titleParser map[musicextractors.ExtractProvider]musicextractors.TitleExtractorFunc
}

var _ MessageProcessorDomain = (*messageProcessorDomain)(nil)

func (s *messageProcessorDomain) extractMusicURL(text string) (parsedMusicLink, error) {
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

// SummarizeThread iterates over every message and creates a summarized response.
//
// Returns the response file or an error if any.
func (s *messageProcessorDomain) SummarizeThread(msgs []slack.Message, channelID, threadTS string) (slack.UploadFileV2Parameters, error) {
	pmls := []parsedMusicLink{}

	for i := range msgs {
		m, eErr := s.extractMusicURL(msgs[i].Text)
		if eErr != nil {
			continue
		}

		pmls = append(pmls, m)
	}

	csvF, size, err := s.createCSV(pmls)
	if err != nil {
		return slack.UploadFileV2Parameters{}, fmt.Errorf("create csv: %w", err)
	}

	fileName := fmt.Sprintf("%s-%s.csv", channelID, threadTS)

	return slack.UploadFileV2Parameters{
		Reader:          csvF,
		Filename:        fileName,
		Title:           fileName,
		InitialComment:  fmt.Sprintf("Found %d music URLs in this thread", len(pmls)),
		Channel:         channelID,
		ThreadTimestamp: threadTS,
		FileSize:        size,
	}, nil
}

func (s *messageProcessorDomain) createCSV(pmls []parsedMusicLink) (io.Reader, int, error) {
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
	urlP map[musicextractors.ExtractProvider]musicextractors.MusicURLExtractorFunc,
	tp map[musicextractors.ExtractProvider]musicextractors.TitleExtractorFunc,
) MessageProcessorDomain {
	return &messageProcessorDomain{
		processors:  urlP,
		titleParser: tp,
	}
}
