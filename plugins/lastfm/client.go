package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ferro-dev/lomax/internal/pluginapi"
)

// DefaultBaseURL is the production Last.fm API endpoint.
const DefaultBaseURL = "https://ws.audioscrobbler.com"

// Client is a minimal Last.fm API client. Last.fm's role in lomax's
// resolution chain is supplemental correction (see docs/music-cli-plan.md
// section 6: "genre enrichment, supplemental metadata"), not primary
// tagging, so this client does a single track.search call and proposes
// corrected title/artist only — no album or year, which would need a
// second track.getInfo request this plugin doesn't make.
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewClient builds a Last.fm client using apiKey
// (https://www.last.fm/api/account/create). An empty apiKey is valid to
// construct with; Search returns (nil, nil) without making a request.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		baseURL:    DefaultBaseURL,
	}
}

// searchResponse mirrors the subset of track.search's JSON response this
// client reads. TrackMatches.Track is left as raw JSON because Last.fm's
// API represents "no results" inconsistently (sometimes an empty array,
// sometimes an empty string) depending on query shape — see Search.
type searchResponse struct {
	Results struct {
		TrackMatches struct {
			Track json.RawMessage `json:"track"`
		} `json:"trackmatches"`
	} `json:"results"`
}

type trackMatch struct {
	Name   string `json:"name"`
	Artist string `json:"artist"`
}

// Search looks up title (required — track.search has no album-only mode)
// by track name and, if given, artist. It returns (nil, nil) if
// unconfigured or if nothing matched.
func (c *Client) Search(ctx context.Context, artist, _, title string) (*pluginapi.Match, error) {
	if c.apiKey == "" || strings.TrimSpace(title) == "" {
		return nil, nil
	}

	q := url.Values{
		"method":  {"track.search"},
		"track":   {title},
		"api_key": {c.apiKey},
		"format":  {"json"},
		"limit":   {"1"},
	}
	if artist != "" {
		q.Set("artist", artist)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/2.0/?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("lastfm: build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lastfm: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lastfm: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lastfm: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var out searchResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("lastfm: parse search response: %w", err)
	}

	var tracks []trackMatch
	// A malformed/empty-string "track" field means no results — treat that
	// the same as a clean zero-length array rather than propagating the
	// unmarshal error.
	_ = json.Unmarshal(out.Results.TrackMatches.Track, &tracks)
	if len(tracks) == 0 {
		return nil, nil
	}
	return &pluginapi.Match{Title: tracks[0].Name, Artist: tracks[0].Artist}, nil
}
