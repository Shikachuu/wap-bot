package musicextractors

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegexURLExtractor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		name    string
		text    string
		pattern string
		want    string
	}{
		{
			name:    "single URL match",
			text:    "Check out https://example.com/path",
			pattern: `https?://example\.com/[\w/\-]+`,
			want:    "https://example.com/path",
			wantErr: nil,
		},
		{
			name:    "no match found",
			text:    "No URLs here",
			pattern: `https?://example\.com/[\w/\-]+`,
			want:    "",
			wantErr: ErrNoURLFound,
		},
		{
			name:    "multiple matches",
			text:    "Check https://example.com/one and https://example.com/two",
			pattern: `https?://example\.com/[\w/\-]+`,
			want:    "",
			wantErr: ErrMultipleResult,
		},
		{
			name:    "empty text",
			text:    "",
			pattern: `https?://example\.com/[\w/\-]+`,
			want:    "",
			wantErr: ErrNoURLFound,
		},
		{
			name:    "pattern matches substring",
			text:    "Text before https://example.com/path text after",
			pattern: `https?://example\.com/[\w/\-]+`,
			want:    "https://example.com/path",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			re, err := regexp.Compile(tt.pattern)
			require.NoError(t, err, "Test regex pattern should compile")

			got, err := regexURLExtractor(tt.text, re)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSpotifyURLExtractor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr      error
		name         string
		text         string
		want         string
		wantProvider ExtractProvider
	}{
		{
			name:         "track URL with open subdomain",
			text:         "Check out https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT",
			want:         "https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT",
			wantProvider: SpotifyProvider,
		},
		{
			name:         "track URL without open subdomain",
			text:         "Check out https://spotify.com/track/4cOdK2wGLETKBW3PvgPWqT",
			want:         "https://spotify.com/track/4cOdK2wGLETKBW3PvgPWqT",
			wantProvider: SpotifyProvider,
		},
		{
			name:         "track URL with query parameters",
			text:         "Listen to https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT?si=abc123",
			want:         "https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT?si=abc123",
			wantProvider: SpotifyProvider,
		},
		{
			name:         "http protocol",
			text:         "Check out http://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT",
			want:         "http://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT",
			wantProvider: SpotifyProvider,
		},
		{
			name:         "playlist URL should fail",
			text:         "My playlist https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M",
			wantProvider: SpotifyProvider,
			wantErr:      ErrNoURLFound,
		},
		{
			name:         "album URL should fail",
			text:         "Great album https://open.spotify.com/album/4LH4d3cOWNNsVw41Gqt2kv",
			wantProvider: SpotifyProvider,
			wantErr:      ErrNoURLFound,
		},
		{
			name:         "artist URL should fail",
			text:         "Check out https://open.spotify.com/artist/0TnOYISbd1XYRBk9myaseg",
			wantProvider: SpotifyProvider,
			wantErr:      ErrNoURLFound,
		},
		{
			name:         "no URL in text",
			text:         "This is just plain text",
			wantProvider: SpotifyProvider,
			wantErr:      ErrNoURLFound,
		},
		{
			name:         "multiple track URLs",
			text:         "Check https://open.spotify.com/track/1 and https://open.spotify.com/track/2",
			wantProvider: SpotifyProvider,
			wantErr:      ErrMultipleResult,
		},
		{
			name:         "non-spotify URL",
			text:         "Check out https://youtube.com/watch?v=abc123",
			wantProvider: SpotifyProvider,
			wantErr:      ErrNoURLFound,
		},
		{
			name:         "empty text",
			text:         "",
			wantProvider: SpotifyProvider,
			wantErr:      ErrNoURLFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, provider, err := SpotifyURLExtractor(tt.text)

			assert.Equal(t, tt.wantProvider, provider)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestYouTubeURLExtractor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr      error
		name         string
		text         string
		want         string
		wantProvider ExtractProvider
	}{
		{
			name:         "youtube.com URL with www",
			text:         "Check out https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			want:         "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			wantProvider: YouTubeProvider,
		},
		{
			name:         "youtube.com URL without www",
			text:         "Watch https://youtube.com/watch?v=dQw4w9WgXcQ",
			want:         "https://youtube.com/watch?v=dQw4w9WgXcQ",
			wantProvider: YouTubeProvider,
		},
		{
			name:         "youtu.be short URL",
			text:         "Check out https://youtu.be/dQw4w9WgXcQ",
			want:         "https://youtu.be/dQw4w9WgXcQ",
			wantProvider: YouTubeProvider,
		},
		{
			name:         "video ID with hyphen",
			text:         "Watch https://youtu.be/dQw4w9Wg-cQ",
			want:         "https://youtu.be/dQw4w9Wg-cQ",
			wantProvider: YouTubeProvider,
		},
		{
			name:         "http protocol",
			text:         "Old link http://www.youtube.com/watch?v=dQw4w9WgXcQ",
			want:         "http://www.youtube.com/watch?v=dQw4w9WgXcQ",
			wantProvider: YouTubeProvider,
		},
		{
			name:         "playlist URL should fail",
			text:         "Check out https://www.youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
			wantProvider: YouTubeProvider,
			wantErr:      ErrNoURLFound,
		},
		{
			name:         "no URL in text",
			text:         "This is just plain text",
			wantProvider: YouTubeProvider,
			wantErr:      ErrNoURLFound,
		},
		{
			name:         "multiple video URLs",
			text:         "Check https://youtube.com/watch?v=abc123 and https://youtu.be/xyz789",
			wantProvider: YouTubeProvider,
			wantErr:      ErrMultipleResult,
		},
		{
			name:         "non-youtube URL",
			text:         "Check out https://open.spotify.com/track/123",
			wantProvider: YouTubeProvider,
			wantErr:      ErrNoURLFound,
		},
		{
			name:         "youtube music URL should not match",
			text:         "Listen to https://music.youtube.com/watch?v=dQw4w9WgXcQ",
			wantProvider: YouTubeProvider,
			wantErr:      ErrNoURLFound,
		},
		{
			name:         "empty text",
			text:         "",
			wantProvider: YouTubeProvider,
			wantErr:      ErrNoURLFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, provider, err := YouTubeURLExtractor(tt.text)

			assert.Equal(t, tt.wantProvider, provider)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestYouTubeMusicURLExtractor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr      error
		name         string
		text         string
		want         string
		wantProvider ExtractProvider
	}{
		{
			name:         "watch URL",
			text:         "Listen to https://music.youtube.com/watch?v=dQw4w9WgXcQ",
			want:         "https://music.youtube.com/watch?v=dQw4w9WgXcQ",
			wantProvider: YoutTubeMusicProvider,
		},
		{
			name:         "watch URL with additional parameters",
			text:         "Check out https://music.youtube.com/watch?v=dQw4w9WgXcQ&list=RDAMVMdQw4w9WgXcQ",
			want:         "https://music.youtube.com/watch?v=dQw4w9WgXcQ&list=RDAMVMdQw4w9WgXcQ",
			wantProvider: YoutTubeMusicProvider,
		},
		{
			name:         "http protocol",
			text:         "Old link http://music.youtube.com/watch?v=dQw4w9WgXcQ",
			want:         "http://music.youtube.com/watch?v=dQw4w9WgXcQ",
			wantProvider: YoutTubeMusicProvider,
		},
		{
			name:         "playlist URL should fail",
			text:         "My playlist https://music.youtube.com/playlist?list=RDCLAK5uy_kmPRjHDECIcuVwnKsx",
			wantProvider: YoutTubeMusicProvider,
			wantErr:      ErrNoURLFound,
		},
		{
			name:         "playlist with parameters should fail",
			text:         "Playlist https://music.youtube.com/playlist?list=RDCLAK5uy_kmPRjHDECIcuVwnKsx&playnext=1",
			wantProvider: YoutTubeMusicProvider,
			wantErr:      ErrNoURLFound,
		},
		{
			name:         "no URL in text",
			text:         "This is just plain text",
			wantProvider: YoutTubeMusicProvider,
			wantErr:      ErrNoURLFound,
		},
		{
			name:         "multiple video URLs",
			text:         "Check https://music.youtube.com/watch?v=abc123 and https://music.youtube.com/watch?v=xyz789",
			wantProvider: YoutTubeMusicProvider,
			wantErr:      ErrMultipleResult,
		},
		{
			name:         "regular youtube URL should not match",
			text:         "Watch https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			wantProvider: YoutTubeMusicProvider,
			wantErr:      ErrNoURLFound,
		},
		{
			name:         "spotify URL should not match",
			text:         "Listen to https://open.spotify.com/track/123",
			wantProvider: YoutTubeMusicProvider,
			wantErr:      ErrNoURLFound,
		},
		{
			name:         "empty text",
			text:         "",
			wantProvider: YoutTubeMusicProvider,
			wantErr:      ErrNoURLFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, provider, err := YouTubeMusicURLExtractor(tt.text)

			assert.Equal(t, tt.wantProvider, provider)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
