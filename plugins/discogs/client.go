package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ferro-dev/lomax/internal/pluginapi"
)

// DefaultBaseURL is the production Discogs API endpoint.
const DefaultBaseURL = "https://api.discogs.com"

// Client is a minimal Discogs API client: just enough to search for a
// release and propose corrected artist/album/year. Discogs' search index
// is release-oriented (see docs/music-cli-plan.md section 3: "good for
// vinyl rips"), not track-oriented — getting a per-track title would need
// a second request to fetch the matched release's full tracklist, which
// this plugin doesn't do. Title is deliberately left unset in every Match
// this client returns.
type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

// NewClient builds a Discogs client using token, a Discogs personal access
// token (https://www.discogs.com/settings/developers). An empty token is
// valid to construct with; Search returns (nil, nil) without making a
// request, matching "this source isn't configured" rather than erroring.
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		baseURL:    DefaultBaseURL,
	}
}

// searchResponse mirrors the subset of /database/search's JSON response
// this client reads.
type searchResponse struct {
	Results []struct {
		Title string `json:"title"` // Discogs formats this as "Artist - Release Title"
		Year  string `json:"year"`
	} `json:"results"`
}

// Search looks up a release by artist and album (title is accepted for
// symmetry with the other sources' Search signatures, but Discogs' release
// search doesn't have a per-track field to use it against). It returns
// (nil, nil) if unconfigured or if nothing matched.
func (c *Client) Search(ctx context.Context, artist, album, _ string) (*pluginapi.Match, error) {
	if c.token == "" || strings.TrimSpace(artist) == "" {
		return nil, nil
	}

	q := url.Values{
		"token":    {c.token},
		"type":     {"release"},
		"artist":   {artist},
		"per_page": {"1"},
	}
	if album != "" {
		q.Set("release_title", album)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/database/search?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("discogs: build request: %w", err)
	}
	// Discogs requires a descriptive User-Agent, same policy as
	// MusicBrainz (see internal/musicbrainz).
	req.Header.Set("User-Agent", "lomax-plugin-discogs/0.1 (+https://github.com/ferro-dev/lomax)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("discogs: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("discogs: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discogs: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var out searchResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("discogs: parse search response: %w", err)
	}
	if len(out.Results) == 0 {
		return nil, nil
	}

	result := out.Results[0]
	match := &pluginapi.Match{}
	if resultArtist, resultAlbum, ok := strings.Cut(result.Title, " - "); ok {
		match.Artist, match.Album = resultArtist, resultAlbum
	} else {
		match.Album = result.Title
	}
	if year, err := strconv.Atoi(result.Year); err == nil {
		match.Year = year
	}
	return match, nil
}
