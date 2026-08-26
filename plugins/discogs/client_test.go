package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, token string, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c := NewClient(token)
	c.baseURL = server.URL
	return c
}

func TestSearchReturnsArtistAlbumYear(t *testing.T) {
	client := newTestClient(t, "test-token", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("token"); got != "test-token" {
			t.Errorf("token param = %q, want test-token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{
				{"title": "Cat Power - Jukebox", "year": "2008"},
			},
		})
	})

	match, err := client.Search(context.Background(), "Cat Power", "Jukebox", "")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if match.Artist != "Cat Power" || match.Album != "Jukebox" || match.Year != 2008 {
		t.Errorf("Search() = %+v, unexpected", match)
	}
	if match.Title != "" {
		t.Errorf("Title = %q, want empty (Discogs release search has no track-level title)", match.Title)
	}
}

func TestSearchNoResultsReturnsNilMatch(t *testing.T) {
	client := newTestClient(t, "test-token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"results": []map[string]any{}})
	})

	match, err := client.Search(context.Background(), "Nobody", "", "")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if match != nil {
		t.Errorf("Search() = %+v, want nil", match)
	}
}

func TestSearchWithNoTokenSkipsRequest(t *testing.T) {
	requested := false
	client := newTestClient(t, "", func(w http.ResponseWriter, r *http.Request) {
		requested = true
	})

	match, err := client.Search(context.Background(), "Artist", "Album", "")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if match != nil || requested {
		t.Errorf("Search() with no token = %+v (requested=%v), want nil match and no request", match, requested)
	}
}

func TestSearchWithNoArtistSkipsRequest(t *testing.T) {
	requested := false
	client := newTestClient(t, "test-token", func(w http.ResponseWriter, r *http.Request) {
		requested = true
	})

	if _, err := client.Search(context.Background(), "", "Album", ""); err != nil {
		t.Fatalf("Search: %v", err)
	}
	if requested {
		t.Error("Search with no artist made a request; Discogs matching needs an artist")
	}
}

func TestSearchPropagatesHTTPErrors(t *testing.T) {
	client := newTestClient(t, "test-token", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	if _, err := client.Search(context.Background(), "Artist", "", ""); err == nil {
		t.Error("Search against a 403 response: got nil error, want an error")
	}
}
