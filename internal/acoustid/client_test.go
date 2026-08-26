package acoustid

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, apiKey string, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewClient(apiKey, WithBaseURL(server.URL))
}

func TestLookupReturnsBestScoringRecording(t *testing.T) {
	client := newTestClient(t, "test-key", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if got := r.FormValue("client"); got != "test-key" {
			t.Errorf("client param = %q, want test-key", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"results": []map[string]any{
				{
					"score": 0.93,
					"recordings": []map[string]any{
						{
							"id":            "rec-mbid-2",
							"title":         "Moon River",
							"artists":       []map[string]any{{"name": "Audrey Hepburn"}},
							"releasegroups": []map[string]any{{"title": "Breakfast at Tiffany's"}},
						},
					},
				},
			},
		})
	})

	match, err := client.Lookup(context.Background(), Fingerprint{DurationSeconds: 120, Data: "AQAA"})
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if match.RecordingMBID != "rec-mbid-2" || match.Title != "Moon River" ||
		match.Artist != "Audrey Hepburn" || match.Album != "Breakfast at Tiffany's" {
		t.Errorf("Lookup result = %+v, unexpected", match)
	}
}

func TestLookupNoAPIKey(t *testing.T) {
	client := NewClient("")
	if _, err := client.Lookup(context.Background(), Fingerprint{}); !errors.Is(err, ErrNoAPIKey) {
		t.Errorf("Lookup with no API key: err = %v, want ErrNoAPIKey", err)
	}
}

func TestLookupNoResults(t *testing.T) {
	client := newTestClient(t, "test-key", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "results": []map[string]any{}})
	})

	if _, err := client.Lookup(context.Background(), Fingerprint{}); !errors.Is(err, ErrNoMatch) {
		t.Errorf("Lookup with no results: err = %v, want ErrNoMatch", err)
	}
}

func TestLookupPropagatesHTTPErrors(t *testing.T) {
	client := newTestClient(t, "test-key", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	if _, err := client.Lookup(context.Background(), Fingerprint{}); err == nil {
		t.Error("Lookup against a 500 response: got nil error, want an error")
	}
}
