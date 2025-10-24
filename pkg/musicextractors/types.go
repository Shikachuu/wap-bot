// Package musicextractors contains the reusable logic for extracting different music URLs from long texts
package musicextractors

// ExtractProvider stands for the implemented URL and Title extract providers.
type ExtractProvider string

const (
	// SpotifyProvider that implements both URL and music title extractor funcs.
	SpotifyProvider ExtractProvider = "spotify"
	// YouTubeProvider that implements both URL and music title extractor funcs.
	YouTubeProvider ExtractProvider = "youtube"
	// YoutTubeMusicProvider that implements both URL and music title extractor funcs.
	YoutTubeMusicProvider ExtractProvider = "youtube-music"
)

// MusicURLExtractorFunc is extracting music links from text messages
//
// text is the input text that possibly contains a link for an implemented provider
//
// returns the extracted url, the provider it used to extract it and an error if any.
type MusicURLExtractorFunc func(text string) (string, ExtractProvider, error)

// TitleExtractorFunc is extracting title and artist information from music urls
//
// url is the input url that we have to fetch some title information for
//
// returns the extracted title and an error if any.
type TitleExtractorFunc func(url string) (string, error)
