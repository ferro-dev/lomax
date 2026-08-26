package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ferro-dev/lomax/internal/audio"
	"github.com/ferro-dev/lomax/internal/library"
	"github.com/ferro-dev/lomax/internal/testsupport"
)

func TestVerifyReportsConsistentLibrary(t *testing.T) {
	dbPath := seedLibrary(t)
	got := runRoot(t, "verify", t.TempDir(), "--library-db", dbPath)
	if !strings.Contains(got, "consistent") {
		t.Errorf("verify output = %q, want it to report consistency", got)
	}
}

func TestVerifyDetectsAndAppliesAMove(t *testing.T) {
	path := testsupport.WriteID3v2Fixture(t, "original.mp3", map[string]string{"TIT2": "T", "TPE1": "A"})
	dir := filepath.Dir(path)
	track, err := audio.ReadTrack(path)
	if err != nil {
		t.Fatalf("ReadTrack: %v", err)
	}
	dbPath := seedLibrary(t, track)

	renamed := filepath.Join(dir, "renamed.mp3")
	if err := os.Rename(path, renamed); err != nil {
		t.Fatalf("rename: %v", err)
	}

	got := runRoot(t, "verify", dir, "--library-db", dbPath)
	if !strings.Contains(got, "moved: "+path+" -> "+renamed) {
		t.Errorf("verify output = %q, want it to report the move", got)
	}

	db, err := library.Open(dbPath)
	if err != nil {
		t.Fatalf("reopen library: %v", err)
	}
	defer func() { _ = db.Close() }()
	all, err := db.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 1 || all[0].Path != renamed {
		t.Errorf("database after verify = %+v, want the row repointed to %s", all, renamed)
	}
}

func TestVerifyReportsMissingAndPrunesOnlyWithFlag(t *testing.T) {
	path := testsupport.WriteID3v2Fixture(t, "gone.mp3", map[string]string{"TIT2": "T", "TPE1": "A"})
	track, err := audio.ReadTrack(path)
	if err != nil {
		t.Fatalf("ReadTrack: %v", err)
	}
	dbPath := seedLibrary(t, track)
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove fixture: %v", err)
	}

	got := runRoot(t, "verify", filepath.Dir(path), "--library-db", dbPath)
	if !strings.Contains(got, "missing: "+path) {
		t.Errorf("verify output = %q, want it to report the missing file", got)
	}

	db, err := library.Open(dbPath)
	if err != nil {
		t.Fatalf("reopen library: %v", err)
	}
	all, err := db.All()
	_ = db.Close()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("row was removed without --prune: All() = %+v", all)
	}

	got = runRoot(t, "verify", filepath.Dir(path), "--library-db", dbPath, "--prune")
	if !strings.Contains(got, "pruned: "+path) {
		t.Errorf("verify --prune output = %q, want it to report the prune", got)
	}

	db, err = library.Open(dbPath)
	if err != nil {
		t.Fatalf("reopen library: %v", err)
	}
	defer func() { _ = db.Close() }()
	all, err = db.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("row still present after --prune: All() = %+v", all)
	}
}

func TestVerifyReportsUntrackedFiles(t *testing.T) {
	dbPath := seedLibrary(t)
	path := testsupport.WriteID3v2Fixture(t, "untracked.mp3", map[string]string{"TIT2": "T"})

	got := runRoot(t, "verify", filepath.Dir(path), "--library-db", dbPath)
	if !strings.Contains(got, "untracked: "+path) {
		t.Errorf("verify output = %q, want it to report the untracked file", got)
	}
}
