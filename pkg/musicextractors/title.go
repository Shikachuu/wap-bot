package musicextractors

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// SpotifyTitleExtractor fetches and extracts the title from a Spotify URL using Open Graph meta tags.
func SpotifyTitleExtractor(musicURL string) (string, error) {
	request, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, musicURL, http.NoBody)
	if err != nil {
		return "", ErrRequestFailed
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", ErrRequestFailed
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", ErrRequestFailed
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", ErrRequestFailed
	}

	html := string(body)

	// Extract og:title for song title
	titleRegex := regexp.MustCompile(`<meta\s+property="og:title"\s+content="([^"]+)"`)
	titleMatches := titleRegex.FindStringSubmatch(html)

	// FindStringSubmatch returns the full match, then the capture groups themselves,
	// hence why we check for the 2. element
	if len(titleMatches) < 2 {
		return "", ErrNoTitleFound
	}

	songTitle := strings.TrimSpace(titleMatches[1])

	// Extract og:description for artist info
	descRegex := regexp.MustCompile(`<meta\s+property="og:description"\s+content="([^"]+)"`)
	descMatches := descRegex.FindStringSubmatch(html)

	if len(descMatches) < 2 {
		// If no description found, just return the title
		return songTitle, nil
	}

	description := strings.TrimSpace(descMatches[1])

	// Description format: "Artist(s) · Album/Song · Type · Year"
	// Split by " · " and take only the first part (artists)
	// We use SplitN here, so we only do the first split, cause we only interested in the first element
	artistParts := strings.SplitN(description, " · ", 2)

	// A short-circuit in case of a spotify html schema cahange
	if len(artistParts) < 2 {
		return description + " - " + songTitle, nil
	}

	return artistParts[0] + " - " + songTitle, nil
}

// YouTubeTitleExtractor fetches and extracts the title from a YouTube URL using oEmbed API.
func YouTubeTitleExtractor(videoURL string) (string, error) {
	// Use YouTube's oEmbed API for faster title extraction
	oembed := url.URL{
		Scheme: "https",
		Host:   "youtube.com",
		Path:   "oembed",
	}
	query := oembed.Query()
	query.Add("format", "json")
	query.Add("url", videoURL)
	oembed.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, oembed.String(), http.NoBody)
	if err != nil {
		return "", ErrRequestFailed
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", ErrRequestFailed
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", ErrRequestFailed
	}

	var result struct {
		Title string `json:"title"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", ErrNoTitleFound
	}

	if result.Title == "" {
		return "", ErrNoTitleFound
	}

	return result.Title, nil
}
