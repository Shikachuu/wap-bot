package musicextractors

import (
	"regexp"
)

// regexURLExtractor extracts the given URL regex from a text message.
func regexURLExtractor(text string, re *regexp.Regexp) (string, error) {
	matches := re.FindAllString(text, -1)

	if matches == nil {
		return "", ErrNoURLFound
	}

	if len(matches) != 1 {
		return "", ErrMultipleResult
	}

	return matches[0], nil
}

// SpotifyURLExtractor finds spotify track links in a given text
//
// returns the found url, the type of ExtractProvider and an error if any.
func SpotifyURLExtractor(text string) (string, ExtractProvider, error) {
	spotifyRegex := regexp.MustCompile(`https?://(?:open\.)?spotify\.com/track/[\w\-?=&]+`)

	url, err := regexURLExtractor(text, spotifyRegex)

	return url, SpotifyProvider, err
}

// YouTubeURLExtractor finds youtube watch links in a given text
//
// returns the found url, the type of ExtractProvider and an error if any.
func YouTubeURLExtractor(text string) (string, ExtractProvider, error) {
	youtubeRegex := regexp.MustCompile(`https?://(?:www\.)?(?:youtube\.com/watch\?v=|youtu\.be/)[\w\-]+`)

	url, err := regexURLExtractor(text, youtubeRegex)

	return url, YouTubeProvider, err
}

// YouTubeMusicURLExtractor finds youtube music watch links in a given text
//
// returns the found url, the type of ExtractProvider and an error if any.
func YouTubeMusicURLExtractor(text string) (string, ExtractProvider, error) {
	youtubeMusicRegex := regexp.MustCompile(`https?://music\.youtube\.com/watch\?v=[\w\-]+(?:&[\w=&\-]+)?`)

	url, err := regexURLExtractor(text, youtubeMusicRegex)

	return url, YoutTubeMusicProvider, err
}
