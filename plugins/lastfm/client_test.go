package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, apiKey string, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c := NewClient(apiKey)
	c.baseURL = server.URL
	return c
}

func TestSearchReturnsTitleAndArtist(t *testing.T) {
	client := newTestClient(t, "test-key", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("api_key"); got != "test-key" {
			t.Errorf("api_key param = %q, want test-key", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":{"trackmatches":{"track":[{"name":"Sea of Love","artist":"Cat Power"}]}}}`))
	})

	match, err := client.Search(context.Background(), "Cat Power", "", "Sea of Love")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if match.Title != "Sea of Love" || match.Artist != "Cat Power" {
		t.Errorf("Search() = %+v, unexpected", match)
	}
}

// TestSearchToleratesEmptyStringTrackField covers a known Last.fm API
// quirk: a zero-result track.search response sometimes represents the
// "track" field as an empty string instead of an empty array.
func TestSearchToleratesEmptyStringTrackField(t *testing.T) {
	client := newTestClient(t, "test-key", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":{"trackmatches":{"track":""}}}`))
	})

	match, err := client.Search(context.Background(), "", "", "Nonexistent Song")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if match != nil {
		t.Errorf("Search() = %+v, want nil", match)
	}
}

func TestSearchWithNoAPIKeySkipsRequest(t *testing.T) {
	requested := false
	client := newTestClient(t, "", func(w http.ResponseWriter, r *http.Request) {
		requested = true
	})

	match, err := client.Search(context.Background(), "Artist", "", "Title")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if match != nil || requested {
		t.Errorf("Search() with no API key = %+v (requested=%v), want nil match and no request", match, requested)
	}
}

func TestSearchWithNoTitleSkipsRequest(t *testing.T) {
	requested := false
	client := newTestClient(t, "test-key", func(w http.ResponseWriter, r *http.Request) {
		requested = true
	})

	if _, err := client.Search(context.Background(), "Artist", "", ""); err != nil {
		t.Fatalf("Search: %v", err)
	}
	if requested {
		t.Error("Search with no title made a request; track.search requires one")
	}
}

func TestSearchPropagatesHTTPErrors(t *testing.T) {
	client := newTestClient(t, "test-key", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	if _, err := client.Search(context.Background(), "", "", "Title"); err == nil {
		t.Error("Search against a 503 response: got nil error, want an error")
	}
}
