package library

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ferro-dev/lomax/internal/audio"
	"github.com/ferro-dev/lomax/internal/testsupport"
)

func TestReconcileDetectsMovedMissingAndUntracked(t *testing.T) {
	dir := t.TempDir()
	db := openTestDB(t)

	movedFrom := testsupport.WriteID3v2Fixture(t, "moved-original.mp3", map[string]string{
		"TIT2": "Will Be Moved", "TPE1": "Artist A",
	})
	deletedPath := testsupport.WriteID3v2Fixture(t, "will-be-deleted.mp3", map[string]string{
		"TIT2": "Will Be Deleted", "TPE1": "Artist B",
	})

	for _, p := range []string{movedFrom, deletedPath} {
		track, err := audio.ReadTrack(p)
		if err != nil {
			t.Fatalf("ReadTrack(%s): %v", p, err)
		}
		if err := db.Upsert(track); err != nil {
			t.Fatalf("Upsert(%s): %v", p, err)
		}
	}

	// Simulate real filesystem changes: rename one tracked file (a move),
	// delete the other (a real deletion), and drop in one brand-new file
	// the database has never seen (untracked).
	movedTo := filepath.Join(dir, "moved-renamed.mp3")
	if err := os.Rename(movedFrom, movedTo); err != nil {
		t.Fatalf("simulate move: %v", err)
	}
	if err := os.Remove(deletedPath); err != nil {
		t.Fatalf("simulate delete: %v", err)
	}
	newFilePath := testsupport.WriteID3v2Fixture(t, "brand-new.mp3", map[string]string{
		"TIT2": "Never Tracked", "TPE1": "Artist C",
	})
	newFileFinal := filepath.Join(dir, "brand-new.mp3")
	if err := os.Rename(newFilePath, newFileFinal); err != nil {
		t.Fatalf("place untracked fixture in scan root: %v", err)
	}

	report, err := db.Reconcile(dir)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(report.Moved) != 1 || report.Moved[0].From != movedFrom || report.Moved[0].To != movedTo {
		t.Errorf("Moved = %+v, want exactly one move from %s to %s", report.Moved, movedFrom, movedTo)
	}
	if len(report.Missing) != 1 || report.Missing[0].Path != deletedPath {
		t.Errorf("Missing = %+v, want exactly the deleted path %s", report.Missing, deletedPath)
	}
	if len(report.Untracked) != 1 || report.Untracked[0] != newFileFinal {
		t.Errorf("Untracked = %+v, want exactly %s", report.Untracked, newFileFinal)
	}
}

func TestReconcileWithNothingChangedReportsNothing(t *testing.T) {
	dir := t.TempDir()
	db := openTestDB(t)

	path := filepath.Join(dir, "track.mp3")
	if err := os.WriteFile(path, testsupport.BuildID3v2Tag(map[string]string{"TIT2": "T", "TPE1": "A"}, 128), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	track, err := audio.ReadTrack(path)
	if err != nil {
		t.Fatalf("ReadTrack: %v", err)
	}
	if err := db.Upsert(track); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	report, err := db.Reconcile(dir)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if len(report.Moved) != 0 || len(report.Missing) != 0 || len(report.Untracked) != 0 {
		t.Errorf("Reconcile with nothing changed = %+v, want an empty report", report)
	}
}
