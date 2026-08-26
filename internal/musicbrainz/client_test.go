package musicbrainz

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewClient("lomax-test/0.0 (test@example.invalid)",
		WithBaseURL(server.URL),
		WithMinInterval(0),
	)
}

func TestSearchRecordingReturnsBestMatch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got == "" {
			t.Error("request missing User-Agent header")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"recordings": []map[string]any{
				{
					"id":    "rec-mbid-1",
					"score": "100",
					"title": "Sea of Love",
					"artist-credit": []map[string]any{
						{"name": "Cat Power"},
					},
					"releases": []map[string]any{
						{"title": "Jukebox", "date": "2008-01-22"},
					},
				},
			},
		})
	})

	rec, err := client.SearchRecording(context.Background(), "Cat Power", "Jukebox", "Sea of Love")
	if err != nil {
		t.Fatalf("SearchRecording: %v", err)
	}
	if rec.MBID != "rec-mbid-1" || rec.Title != "Sea of Love" || rec.Artist != "Cat Power" ||
		rec.Album != "Jukebox" || rec.Year != 2008 || rec.Score != 100 {
		t.Errorf("SearchRecording result = %+v, unexpected", rec)
	}
}

func TestSearchRecordingRejectsLowScore(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"recordings": []map[string]any{
				{"id": "weak-match", "score": "40", "title": "Something Else"},
			},
		})
	})

	_, err := client.SearchRecording(context.Background(), "Cat Power", "", "Sea of Love")
	if !errors.Is(err, ErrNoMatch) {
		t.Errorf("SearchRecording with a below-threshold score: err = %v, want ErrNoMatch", err)
	}
}

func TestSearchRecordingNoResults(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"recordings": []map[string]any{}})
	})

	_, err := client.SearchRecording(context.Background(), "Nobody", "", "Nothing")
	if !errors.Is(err, ErrNoMatch) {
		t.Errorf("SearchRecording with no results: err = %v, want ErrNoMatch", err)
	}
}

func TestSearchRecordingRequiresTitle(t *testing.T) {
	client := NewClient("lomax-test/0.0")
	if _, err := client.SearchRecording(context.Background(), "Artist", "", "  "); err == nil {
		t.Error("SearchRecording with a blank title: got nil error, want an error")
	}
}

func TestSearchRecordingPropagatesHTTPErrors(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	if _, err := client.SearchRecording(context.Background(), "Artist", "", "Title"); err == nil {
		t.Error("SearchRecording against a 503 response: got nil error, want an error")
	}
}

func TestClientEnforcesMinInterval(t *testing.T) {
	var requestTimes []time.Time
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestTimes = append(requestTimes, time.Now())
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"recordings": []map[string]any{}})
	}))
	t.Cleanup(server.Close)

	client := NewClient("lomax-test/0.0", WithBaseURL(server.URL), WithMinInterval(50*time.Millisecond))

	_, _ = client.SearchRecording(context.Background(), "A", "", "One")
	_, _ = client.SearchRecording(context.Background(), "A", "", "Two")

	if len(requestTimes) != 2 {
		t.Fatalf("got %d requests, want 2", len(requestTimes))
	}
	if gap := requestTimes[1].Sub(requestTimes[0]); gap < 50*time.Millisecond {
		t.Errorf("requests were %v apart, want at least 50ms", gap)
	}
}
