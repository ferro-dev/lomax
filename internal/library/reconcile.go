package library

import (
	"fmt"
	"os"

	"github.com/ferro-dev/lomax/internal/audio"
)

// Move is a track Reconcile matched from a missing database row to a new,
// untracked file on disk.
type Move struct {
	From string
	To   string
}

// Report is Reconcile's result.
type Report struct {
	// Missing are tracked rows whose file no longer exists on disk and
	// couldn't be matched to a moved file.
	Missing []Record
	// Moved are tracked rows matched to a new path.
	Moved []Move
	// Untracked are audio files found under the scanned root that aren't
	// in the database at all.
	Untracked []string
}

// Reconcile compares the database against the actual files under root: it
// finds tracked rows whose file has vanished, tries to match each to an
// untracked file that looks like the same track (by size, title, and
// artist — a moved rather than deleted file), and reports whatever's left
// as genuinely missing or genuinely untracked. It does not modify the
// database or the filesystem — see the `lomax verify` command for applying
// the result.
func (db *DB) Reconcile(root string) (Report, error) {
	all, err := db.All()
	if err != nil {
		return Report{}, err
	}

	tracked := make(map[string]bool, len(all))
	var missing []Record
	for _, r := range all {
		tracked[r.Path] = true
		if _, err := os.Stat(r.Path); err != nil {
			missing = append(missing, r)
		}
	}

	diskFiles, err := audio.Scan(root)
	if err != nil {
		return Report{}, fmt.Errorf("library: scan %s: %w", root, err)
	}

	var untracked []string
	for _, f := range diskFiles {
		if !tracked[f] {
			untracked = append(untracked, f)
		}
	}

	claimed := make(map[string]bool, len(untracked))
	var moved []Move
	var stillMissing []Record
	for _, m := range missing {
		match := findMove(m, untracked, claimed)
		if match == "" {
			stillMissing = append(stillMissing, m)
			continue
		}
		claimed[match] = true
		moved = append(moved, Move{From: m.Path, To: match})
	}

	var remainingUntracked []string
	for _, f := range untracked {
		if !claimed[f] {
			remainingUntracked = append(remainingUntracked, f)
		}
	}

	return Report{Missing: stillMissing, Moved: moved, Untracked: remainingUntracked}, nil
}

// findMove looks for an unclaimed candidate that plausibly is missing,
// relocated: matching file size first (cheap, from a stat call) before
// paying for a tag read to confirm title and artist still match. Returns
// "" if nothing matches.
func findMove(missing Record, candidates []string, claimed map[string]bool) string {
	for _, c := range candidates {
		if claimed[c] {
			continue
		}
		info, err := os.Stat(c)
		if err != nil || info.Size() != missing.Size {
			continue
		}
		track, err := audio.ReadTrack(c)
		if err != nil {
			continue
		}
		if track.Title == missing.Title && track.Artist == missing.Artist {
			return c
		}
	}
	return ""
}
