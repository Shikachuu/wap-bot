package musicextractors

import (
	"errors"
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

func SpotifyURLExtractor(text string) (string, ExtractProvider, error) {
	spotifyRegex, err := regexp.Compile(`https?://(?:open\.)?spotify\.com/track/[\w\-?=&]+`)
	if err != nil {
		return "", SpotifyProvider, errors.Join(ErrInvalidRegexCompiled, err)
	}

	url, err := regexURLExtractor(text, spotifyRegex)

	return url, SpotifyProvider, err
}

func YouTubeURLExtractor(text string) (string, ExtractProvider, error) {
	youtubeRegex, err := regexp.Compile(`https?://(?:www\.)?(?:youtube\.com/watch\?v=|youtu\.be/)[\w\-]+`)
	if err != nil {
		return "", YouTubeProvider, errors.Join(ErrInvalidRegexCompiled, err)
	}

	url, err := regexURLExtractor(text, youtubeRegex)

	return url, YouTubeProvider, err
}

func YouTubeMusicURLExtractor(text string) (string, ExtractProvider, error) {
	youtubeMusicRegex, err := regexp.Compile(`https?://music\.youtube\.com/watch\?v=[\w\-]+(?:&[\w=&\-]+)?`)
	if err != nil {
		return "", YoutTubeMusicProvider, errors.Join(ErrInvalidRegexCompiled, err)
	}

	url, err := regexURLExtractor(text, youtubeMusicRegex)

	return url, YoutTubeMusicProvider, err
}
