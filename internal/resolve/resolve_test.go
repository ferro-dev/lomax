package resolve

import (
	"context"
	"errors"
	"testing"

	"github.com/ferro-dev/lomax/internal/acoustid"
	"github.com/ferro-dev/lomax/internal/audio"
	"github.com/ferro-dev/lomax/internal/musicbrainz"
	"github.com/ferro-dev/lomax/internal/pluginapi"
)

func TestResolveReturnsMusicBrainzMatchWithoutTryingAcoustID(t *testing.T) {
	acoustIDCalled := false
	r := &Resolver{
		SearchRecording: func(ctx context.Context, artist, album, title string) (*musicbrainz.Recording, error) {
			return &musicbrainz.Recording{MBID: "mb-1", Title: title, Artist: artist, Album: "Album", Year: 1999, Score: 100}, nil
		},
		Fingerprint: func(path string) (acoustid.Fingerprint, error) {
			acoustIDCalled = true
			return acoustid.Fingerprint{}, nil
		},
		LookupFingerprint: func(ctx context.Context, fp acoustid.Fingerprint) (*acoustid.Match, error) {
			acoustIDCalled = true
			return nil, acoustid.ErrNoMatch
		},
	}

	track := audio.Track{Artist: "Artist", Title: "Title"}
	proposal, warnings, err := r.Resolve(context.Background(), track)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if proposal == nil || proposal.Source != "MusicBrainz recording mb-1 (score 100)" {
		t.Errorf("proposal = %+v, want a MusicBrainz-sourced proposal", proposal)
	}
	if acoustIDCalled {
		t.Error("AcoustID was consulted despite a confident MusicBrainz match")
	}
}

func TestResolveFallsBackToAcoustIDWhenMusicBrainzHasNoMatch(t *testing.T) {
	r := &Resolver{
		SearchRecording: func(ctx context.Context, artist, album, title string) (*musicbrainz.Recording, error) {
			return nil, musicbrainz.ErrNoMatch
		},
		Fingerprint: func(path string) (acoustid.Fingerprint, error) {
			return acoustid.Fingerprint{DurationSeconds: 200, Data: "AQAA"}, nil
		},
		LookupFingerprint: func(ctx context.Context, fp acoustid.Fingerprint) (*acoustid.Match, error) {
			return &acoustid.Match{RecordingMBID: "ac-1", Title: "Found Title", Artist: "Found Artist"}, nil
		},
	}

	track := audio.Track{Artist: "Artist", Title: "Title", Path: "/tmp/track.mp3"}
	proposal, _, err := r.Resolve(context.Background(), track)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if proposal == nil || proposal.Title != "Found Title" || proposal.Source != "AcoustID fingerprint match (recording ac-1)" {
		t.Errorf("proposal = %+v, want an AcoustID-sourced proposal", proposal)
	}
}

func TestResolveNoMatchFromAnySourceReturnsNilProposal(t *testing.T) {
	r := &Resolver{
		SearchRecording: func(ctx context.Context, artist, album, title string) (*musicbrainz.Recording, error) {
			return nil, musicbrainz.ErrNoMatch
		},
		Fingerprint: func(path string) (acoustid.Fingerprint, error) {
			return acoustid.Fingerprint{}, nil
		},
		LookupFingerprint: func(ctx context.Context, fp acoustid.Fingerprint) (*acoustid.Match, error) {
			return nil, acoustid.ErrNoMatch
		},
	}

	proposal, _, err := r.Resolve(context.Background(), audio.Track{Artist: "Artist", Title: "Title"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if proposal != nil {
		t.Errorf("proposal = %+v, want nil", proposal)
	}
}

func TestResolveSkipsMusicBrainzWhenTagsAreMissing(t *testing.T) {
	called := false
	r := &Resolver{
		SearchRecording: func(ctx context.Context, artist, album, title string) (*musicbrainz.Recording, error) {
			called = true
			return nil, musicbrainz.ErrNoMatch
		},
	}

	// No Artist or Title on the track — nothing to search MusicBrainz with.
	if _, _, err := r.Resolve(context.Background(), audio.Track{}); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if called {
		t.Error("SearchRecording was called despite the track having no artist/title tags")
	}
}

func TestResolveCollectsWarningsWithoutFailing(t *testing.T) {
	r := &Resolver{
		SearchRecording: func(ctx context.Context, artist, album, title string) (*musicbrainz.Recording, error) {
			return nil, errors.New("network exploded")
		},
	}

	proposal, warnings, err := r.Resolve(context.Background(), audio.Track{Artist: "A", Title: "T"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if proposal != nil {
		t.Errorf("proposal = %+v, want nil", proposal)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings = %v, want exactly one", warnings)
	}
}

func TestResolveFallsBackToPluginSourcesInOrder(t *testing.T) {
	firstCalled := false
	r := &Resolver{
		PluginSources: []PluginSource{
			{
				Name: "discogs",
				ResolveMetadata: func(ctx context.Context, artist, album, title string) (*pluginapi.Match, error) {
					firstCalled = true
					return nil, nil // no match — falls through to the next plugin
				},
			},
			{
				Name: "lastfm",
				ResolveMetadata: func(ctx context.Context, artist, album, title string) (*pluginapi.Match, error) {
					return &pluginapi.Match{Title: "Found via Last.fm", Artist: artist}, nil
				},
			},
		},
	}

	proposal, _, err := r.Resolve(context.Background(), audio.Track{Artist: "Artist", Title: "Title"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !firstCalled {
		t.Error("first plugin source was never tried")
	}
	if proposal == nil || proposal.Title != "Found via Last.fm" || proposal.Source != "lastfm plugin match" {
		t.Errorf("proposal = %+v, want the second plugin source's match", proposal)
	}
}

func TestResolvePluginSourceErrorIsAWarningNotAFailure(t *testing.T) {
	r := &Resolver{
		PluginSources: []PluginSource{
			{
				Name: "flaky",
				ResolveMetadata: func(ctx context.Context, artist, album, title string) (*pluginapi.Match, error) {
					return nil, errors.New("plugin subprocess crashed")
				},
			},
		},
	}

	proposal, warnings, err := r.Resolve(context.Background(), audio.Track{Artist: "A", Title: "T"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if proposal != nil {
		t.Errorf("proposal = %+v, want nil", proposal)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings = %v, want exactly one", warnings)
	}
}

func TestResolveWithNoSourcesConfiguredReturnsNilProposal(t *testing.T) {
	r := &Resolver{}
	proposal, warnings, err := r.Resolve(context.Background(), audio.Track{Artist: "A", Title: "T"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if proposal != nil || len(warnings) != 0 {
		t.Errorf("Resolve with no sources configured = %+v, %v, want nil, no warnings", proposal, warnings)
	}
}
