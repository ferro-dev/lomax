// Package musicbrainz is a thin hand-rolled client for the MusicBrainz WS/2
// JSON search API. No mature Go client exists for this API (see
// docs/music-cli-plan.md section 6), so this wraps only what lomax's
// metadata resolver needs: searching for a recording by artist/album/title.
package musicbrainz

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
	"sync"
	"time"
)

// DefaultBaseURL is the production MusicBrainz WS/2 endpoint.
const DefaultBaseURL = "https://musicbrainz.org/ws/2"

// defaultMinInterval enforces MusicBrainz's documented rate limit for
// unauthenticated clients: no more than one request per second.
const defaultMinInterval = time.Second

// ErrNoMatch is returned when a search completes successfully but no result
// meets the client's minimum acceptance score.
var ErrNoMatch = errors.New("musicbrainz: no matching recording found")

// Recording is the subset of a MusicBrainz recording search result lomax's
// resolver needs. Track/disc numbers are deliberately not included: the
// search endpoint's embedded release summaries don't carry media/track
// listings, and fetching them requires a separate per-candidate lookup that
// M2 intentionally defers (see docs/music-cli-plan.md Milestone 2 notes).
type Recording struct {
	MBID   string
	Title  string
	Score  int
	Artist string
	Album  string
	Year   int
}

// Client is a MusicBrainz WS/2 search client. The zero value is not usable;
// construct with NewClient.
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string

	mu          sync.Mutex
	lastRequest time.Time
	minInterval time.Duration
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

// WithMinInterval overrides the minimum spacing between requests. Tests set
// this to 0 to avoid paying the real rate limit.
func WithMinInterval(d time.Duration) Option {
	return func(c *Client) { c.minInterval = d }
}

// NewClient builds a MusicBrainz client. userAgent must identify the
// application and a contact method per MusicBrainz's API usage policy
// (requests with a generic or missing User-Agent may be rejected).
func NewClient(userAgent string, opts ...Option) *Client {
	c := &Client{
		httpClient:  &http.Client{Timeout: 15 * time.Second},
		baseURL:     DefaultBaseURL,
		userAgent:   userAgent,
		minInterval: defaultMinInterval,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// searchResponse mirrors the subset of the /recording search JSON response
// this client reads.
type searchResponse struct {
	Recordings []struct {
		ID     string      `json:"id"`
		Score  json.Number `json:"score"`
		Title  string      `json:"title"`
		Artist []struct {
			Name string `json:"name"`
		} `json:"artist-credit"`
		Releases []struct {
			Title string `json:"title"`
			Date  string `json:"date"`
		} `json:"releases"`
	} `json:"recordings"`
}

// MinScore is the minimum MusicBrainz relevance score (0-100) SearchRecording
// will accept as a match. Below this, a result is treated as no match at
// all — a wrong tag is worse than no proposal.
const MinScore = 80

// SearchRecording looks up a recording by artist, album, and title (album
// may be empty). It returns ErrNoMatch if the search found nothing scoring
// at least MinScore.
func (c *Client) SearchRecording(ctx context.Context, artist, album, title string) (*Recording, error) {
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("musicbrainz: title is required to search")
	}

	var clauses []string
	clauses = append(clauses, "recording:"+luceneQuote(title))
	if strings.TrimSpace(artist) != "" {
		clauses = append(clauses, "artist:"+luceneQuote(artist))
	}
	if strings.TrimSpace(album) != "" {
		clauses = append(clauses, "release:"+luceneQuote(album))
	}
	query := strings.Join(clauses, " AND ")

	u := fmt.Sprintf("%s/recording/?query=%s&fmt=json&limit=5", c.baseURL, url.QueryEscape(query))
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}

	var resp searchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("musicbrainz: parse search response: %w", err)
	}
	if len(resp.Recordings) == 0 {
		return nil, ErrNoMatch
	}

	best := resp.Recordings[0]
	score, _ := best.Score.Int64()
	if int(score) < MinScore {
		return nil, ErrNoMatch
	}

	rec := &Recording{
		MBID:  best.ID,
		Title: best.Title,
		Score: int(score),
	}
	if len(best.Artist) > 0 {
		names := make([]string, len(best.Artist))
		for i, a := range best.Artist {
			names[i] = a.Name
		}
		rec.Artist = strings.Join(names, ", ")
	}
	if len(best.Releases) > 0 {
		rec.Album = best.Releases[0].Title
		if len(best.Releases[0].Date) >= 4 {
			if year, err := strconv.Atoi(best.Releases[0].Date[:4]); err == nil {
				rec.Year = year
			}
		}
	}
	return rec, nil
}

// get performs a rate-limited GET request and returns the response body.
func (c *Client) get(ctx context.Context, u string) ([]byte, error) {
	c.wait()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("musicbrainz: build request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("musicbrainz: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("musicbrainz: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("musicbrainz: unexpected status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// wait blocks, if necessary, so requests stay at least minInterval apart.
func (c *Client) wait() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.lastRequest.IsZero() {
		if elapsed := time.Since(c.lastRequest); elapsed < c.minInterval {
			time.Sleep(c.minInterval - elapsed)
		}
	}
	c.lastRequest = time.Now()
}

// luceneQuote wraps s as a Lucene phrase query, escaping backslashes and
// double quotes. This handles ordinary artist/album/title values (spaces,
// punctuation) but does not escape the full set of Lucene special
// characters (+ - && || ! ( ) { } [ ] ^ ~ * ? :) — a title containing those
// as literal text could produce an unintended query. Acceptable for M2;
// revisit if this surfaces in practice.
func luceneQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}
