package musicextractors

import "errors"

var (
	// ErrNoURLFound returned by MusicURLExtractorFunc if no URL was found in text.
	ErrNoURLFound = errors.New("no URL found in text")
	// ErrMultipleResult returned by MusicURLExtractorFunc if multiple URLs was in a single text.
	ErrMultipleResult = errors.New("multiple results found in string")

	// ErrNoTitleFound returned by TitleExtractorFunc if it was unable to find any title info.
	ErrNoTitleFound = errors.New("no title found in page")
	// ErrRequestFailed returned by TitleExtractorFunc if it was unable to make the necessary API calls to determine the title.
	ErrRequestFailed = errors.New("failed to fetch URL")
)
