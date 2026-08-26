package cli

import (
	"testing"

	"github.com/ferro-dev/lomax/internal/resolve"
)

func TestShouldApply(t *testing.T) {
	proposal := &resolve.Proposal{Title: "Something"}
	cases := []struct {
		name     string
		proposal *resolve.Proposal
		dryRun   bool
		want     bool
	}{
		{"proposal and not dry-run", proposal, false, true},
		{"proposal but dry-run", proposal, true, false},
		{"no proposal, not dry-run", nil, false, false},
		{"no proposal and dry-run", nil, true, false},
	}
	for _, c := range cases {
		if got := shouldApply(c.proposal, c.dryRun); got != c.want {
			t.Errorf("%s: shouldApply() = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestWritableTagsFromProposal(t *testing.T) {
	if got := writableTagsFromProposal(nil); !got.IsZero() {
		t.Errorf("writableTagsFromProposal(nil) = %+v, want zero value", got)
	}

	proposal := &resolve.Proposal{Title: "T", Artist: "A", Album: "Al", AlbumArtist: "AA", Year: 2000}
	got := writableTagsFromProposal(proposal)
	if got.Title != "T" || got.Artist != "A" || got.Album != "Al" || got.AlbumArtist != "AA" || got.Year != 2000 {
		t.Errorf("writableTagsFromProposal(%+v) = %+v, unexpected", proposal, got)
	}
}

func TestDiffRowNoProposal(t *testing.T) {
	row := diffRow("Title", "Current", "")
	if row[2] != "(no change)" {
		t.Errorf("diffRow with empty proposed = %q, want %q", row[2], "(no change)")
	}
}

func TestYearString(t *testing.T) {
	if got := yearString(0); got != "" {
		t.Errorf("yearString(0) = %q, want empty", got)
	}
	if got := yearString(1999); got != "1999" {
		t.Errorf("yearString(1999) = %q, want %q", got, "1999")
	}
}
