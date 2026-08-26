package library

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/ferro-dev/lomax/internal/audio"
)

// Record is one track row, denormalized with its artist/album names — the
// shape All, Search, and Reconcile hand back, as opposed to audio.Track
// (what a fresh tag read produces) or the raw normalized artist/album/track
// tables (this package's own storage concern).
type Record struct {
	Path        string
	Title       string
	Artist      string
	AlbumArtist string
	Album       string
	Year        int
	TrackNum    int
	TrackTotal  int
	DiscNum     int
	DiscTotal   int
	Size        int64
	ModTime     time.Time
}

// selectRecordsSQL is the shared SELECT All and Search both filter/order.
const selectRecordsSQL = `
SELECT tracks.path, tracks.title,
       COALESCE(artists.name, ''),
       COALESCE(album_artists.name, ''),
       COALESCE(albums.title, ''),
       tracks.year, tracks.track_num, tracks.track_total,
       tracks.disc_num, tracks.disc_total, tracks.size, tracks.mod_time_unix
FROM tracks
LEFT JOIN artists ON tracks.artist_id = artists.id
LEFT JOIN albums ON tracks.album_id = albums.id
LEFT JOIN artists AS album_artists ON albums.album_artist_id = album_artists.id`

const orderBySQL = " ORDER BY artists.name, albums.title, tracks.disc_num, tracks.track_num"

// Upsert records track as a managed file: inserting it (and its artist/
// album, if new) or updating its row if track.Path is already tracked.
func (db *DB) Upsert(track audio.Track) error {
	modTime := int64(0)
	if info, err := os.Stat(track.Path); err == nil {
		modTime = info.ModTime().Unix()
	}

	tx, err := db.sql.Begin()
	if err != nil {
		return fmt.Errorf("library: begin upsert of %s: %w", track.Path, err)
	}
	defer func() { _ = tx.Rollback() }() // no-op once committed

	artistID, err := upsertArtist(tx, track.Artist)
	if err != nil {
		return err
	}
	albumArtistID, err := upsertArtist(tx, track.AlbumArtist)
	if err != nil {
		return err
	}
	albumID, err := upsertAlbum(tx, track.Album, albumArtistID, track.Year)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	_, err = tx.Exec(`
		INSERT INTO tracks (
			path, title, artist_id, album_id, track_num, track_total,
			disc_num, disc_total, year, size, mod_time_unix, added_at_unix, updated_at_unix
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			title = excluded.title,
			artist_id = excluded.artist_id,
			album_id = excluded.album_id,
			track_num = excluded.track_num,
			track_total = excluded.track_total,
			disc_num = excluded.disc_num,
			disc_total = excluded.disc_total,
			year = excluded.year,
			size = excluded.size,
			mod_time_unix = excluded.mod_time_unix,
			updated_at_unix = excluded.updated_at_unix`,
		track.Path, track.Title, nullableID(artistID), nullableID(albumID),
		track.TrackNum, track.TrackTotal, track.DiscNum, track.DiscTotal,
		track.Year, track.Size, modTime, now, now,
	)
	if err != nil {
		return fmt.Errorf("library: upsert track %s: %w", track.Path, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("library: commit upsert of %s: %w", track.Path, err)
	}
	return nil
}

// Delete removes path's track row, if any. Deleting a path that isn't
// tracked is not an error.
func (db *DB) Delete(path string) error {
	if _, err := db.sql.Exec(`DELETE FROM tracks WHERE path = ?`, path); err != nil {
		return fmt.Errorf("library: delete %s: %w", path, err)
	}
	return nil
}

// UpdatePath repoints an existing track row from oldPath to newPath,
// without touching any other column — used when Reconcile detects a file
// was moved rather than added/removed.
func (db *DB) UpdatePath(oldPath, newPath string) error {
	res, err := db.sql.Exec(`UPDATE tracks SET path = ?, updated_at_unix = ? WHERE path = ?`,
		newPath, time.Now().Unix(), oldPath)
	if err != nil {
		return fmt.Errorf("library: move %s to %s: %w", oldPath, newPath, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("library: move %s to %s: %w", oldPath, newPath, err)
	}
	if n == 0 {
		return fmt.Errorf("library: move %s to %s: no track at that path", oldPath, newPath)
	}
	return nil
}

// All returns every tracked track, ordered by artist/album/disc/track.
func (db *DB) All() ([]Record, error) {
	rows, err := db.sql.Query(selectRecordsSQL + orderBySQL)
	if err != nil {
		return nil, fmt.Errorf("library: query tracks: %w", err)
	}
	return scanRecords(rows)
}

func scanRecords(rows *sql.Rows) ([]Record, error) {
	defer func() { _ = rows.Close() }()

	var out []Record
	for rows.Next() {
		var r Record
		var modUnix int64
		if err := rows.Scan(&r.Path, &r.Title, &r.Artist, &r.AlbumArtist, &r.Album,
			&r.Year, &r.TrackNum, &r.TrackTotal, &r.DiscNum, &r.DiscTotal, &r.Size, &modUnix); err != nil {
			return nil, fmt.Errorf("library: scan track row: %w", err)
		}
		r.ModTime = time.Unix(modUnix, 0)
		out = append(out, r)
	}
	return out, rows.Err()
}

// nullableID converts the sentinel "0 = no row" used throughout this file
// into a real SQL NULL for the FK columns.
func nullableID(id int64) any {
	if id == 0 {
		return nil
	}
	return id
}

// upsertArtist returns name's artist id, inserting a row if needed. An
// empty name returns 0 (no artist) without touching the table.
func upsertArtist(tx *sql.Tx, name string) (int64, error) {
	if name == "" {
		return 0, nil
	}
	if _, err := tx.Exec(`INSERT OR IGNORE INTO artists (name) VALUES (?)`, name); err != nil {
		return 0, fmt.Errorf("library: upsert artist %q: %w", name, err)
	}
	var id int64
	if err := tx.QueryRow(`SELECT id FROM artists WHERE name = ?`, name).Scan(&id); err != nil {
		return 0, fmt.Errorf("library: fetch artist %q: %w", name, err)
	}
	return id, nil
}

// upsertAlbum returns title's album id (scoped to albumArtistID, which may
// be 0 for "no album artist"), inserting a row if needed and keeping year
// current. An empty title returns 0 (no album) without touching the table.
func upsertAlbum(tx *sql.Tx, title string, albumArtistID int64, year int) (int64, error) {
	if title == "" {
		return 0, nil
	}

	if _, err := tx.Exec(`INSERT OR IGNORE INTO albums (title, album_artist_id, year) VALUES (?, ?, ?)`,
		title, nullableID(albumArtistID), year); err != nil {
		return 0, fmt.Errorf("library: upsert album %q: %w", title, err)
	}

	var id int64
	var row *sql.Row
	if albumArtistID != 0 {
		row = tx.QueryRow(`SELECT id FROM albums WHERE title = ? AND album_artist_id = ?`, title, albumArtistID)
	} else {
		// album_artist_id is NULL for "no album artist" — "= NULL" never
		// matches in SQL, so this case needs its own query.
		row = tx.QueryRow(`SELECT id FROM albums WHERE title = ? AND album_artist_id IS NULL`, title)
	}
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("library: fetch album %q: %w", title, err)
	}

	if year != 0 {
		if _, err := tx.Exec(`UPDATE albums SET year = ? WHERE id = ?`, year, id); err != nil {
			return 0, fmt.Errorf("library: update album %q year: %w", title, err)
		}
	}
	return id, nil
}
