package acoustid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL is the production AcoustID web service endpoint.
const DefaultBaseURL = "https://api.acoustid.org/v2"

// ErrNoAPIKey is returned by Lookup when the client has no API key
// configured. AcoustID requires free registration for a client key; lomax
// treats a missing key as "this source is disabled" rather than an error
// the user must silence (matches the per-source-optional design in
// docs/music-cli-plan.md section 6).
var ErrNoAPIKey = errors.New("acoustid: no API key configured")

// ErrNoMatch is returned when the lookup completes but AcoustID has no
// recording matching the fingerprint.
var ErrNoMatch = errors.New("acoustid: no matching recording found")

// Match is the subset of an AcoustID lookup result lomax's resolver needs.
type Match struct {
	RecordingMBID string
	Title         string
	Artist        string
	Album         string
	Score         float64
}

// Client is an AcoustID lookup API client. The zero value is not usable;
// construct with NewClient.
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the API base URL. Used by tests to point at an
// httptest.Server instead of the production API.
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

// WithHTTPClient overrides the underlying *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// NewClient builds an AcoustID client using apiKey. An empty apiKey is
// valid to construct with, but Lookup will return ErrNoAPIKey.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		baseURL:    DefaultBaseURL,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// lookupResponse mirrors the subset of the /lookup JSON response this
// client reads (meta=recordings+releasegroups).
type lookupResponse struct {
	Status  string `json:"status"`
	Results []struct {
		Score      float64 `json:"score"`
		Recordings []struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
			ReleaseGroups []struct {
				Title string `json:"title"`
			} `json:"releasegroups"`
		} `json:"recordings"`
	} `json:"results"`
}

// Lookup submits fp to the AcoustID lookup API and returns the best-scoring
// recording match. It returns ErrNoAPIKey if the client has no API key, and
// ErrNoMatch if the API returned no recordings for the fingerprint.
func (c *Client) Lookup(ctx context.Context, fp Fingerprint) (*Match, error) {
	if c.apiKey == "" {
		return nil, ErrNoAPIKey
	}

	// POST rather than GET: fingerprints are long enough to risk exceeding
	// URL length limits, which is why AcoustID's own docs recommend POST.
	form := url.Values{
		"client":      {c.apiKey},
		"meta":        {"recordings+releasegroups"},
		"duration":    {strconv.Itoa(fp.DurationSeconds)},
		"fingerprint": {fp.Data},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/lookup", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("acoustid: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("acoustid: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("acoustid: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("acoustid: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var out lookupResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("acoustid: parse lookup response: %w", err)
	}
	if out.Status != "ok" {
		return nil, fmt.Errorf("acoustid: lookup returned status %q", out.Status)
	}

	for _, result := range out.Results {
		if len(result.Recordings) == 0 {
			continue
		}
		rec := result.Recordings[0]
		match := &Match{RecordingMBID: rec.ID, Title: rec.Title, Score: result.Score}
		if len(rec.Artists) > 0 {
			names := make([]string, len(rec.Artists))
			for i, a := range rec.Artists {
				names[i] = a.Name
			}
			match.Artist = strings.Join(names, ", ")
		}
		if len(rec.ReleaseGroups) > 0 {
			match.Album = rec.ReleaseGroups[0].Title
		}
		return match, nil
	}
	return nil, ErrNoMatch
}
