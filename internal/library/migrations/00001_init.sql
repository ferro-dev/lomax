-- +goose Up
CREATE TABLE artists (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE albums (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    title           TEXT NOT NULL,
    album_artist_id INTEGER REFERENCES artists (id),
    year            INTEGER NOT NULL DEFAULT 0,
    UNIQUE (title, album_artist_id)
);

CREATE INDEX idx_albums_album_artist_id ON albums (album_artist_id);

CREATE TABLE tracks (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    path            TEXT NOT NULL UNIQUE,
    title           TEXT NOT NULL DEFAULT '',
    artist_id       INTEGER REFERENCES artists (id),
    album_id        INTEGER REFERENCES albums (id),
    track_num       INTEGER NOT NULL DEFAULT 0,
    track_total     INTEGER NOT NULL DEFAULT 0,
    disc_num        INTEGER NOT NULL DEFAULT 0,
    disc_total      INTEGER NOT NULL DEFAULT 0,
    year            INTEGER NOT NULL DEFAULT 0,
    size            INTEGER NOT NULL DEFAULT 0,
    mod_time_unix   INTEGER NOT NULL DEFAULT 0,
    added_at_unix   INTEGER NOT NULL DEFAULT 0,
    updated_at_unix INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_tracks_artist_id ON tracks (artist_id);
CREATE INDEX idx_tracks_album_id ON tracks (album_id);

-- +goose Down
DROP TABLE tracks;
DROP TABLE albums;
DROP TABLE artists;
