package library

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ferro-dev/lomax/internal/query"
)

// queryColumns maps a query.Clause field to the SQL column Search filters
// on. Defined here, not in package query, because the column/join layout
// is this package's schema concern — query stays schema-agnostic so it can
// be reused for Milestone 7's sync-profile filters against different data.
var queryColumns = map[string]string{
	"artist":       "artists.name",
	"album_artist": "album_artists.name",
	"album":        "albums.title",
	"title":        "tracks.title",
	"year":         "tracks.year",
	"track":        "tracks.track_num",
	"disc":         "tracks.disc_num",
}

// numericQueryFields are query.Clause fields compared as integers.
var numericQueryFields = map[string]bool{"year": true, "track": true, "disc": true}

// Search returns every track matching every clause in q (ANDed), via a
// parameterized query — clause values are never interpolated into SQL text.
func (db *DB) Search(q *query.Query) ([]Record, error) {
	var conditions []string
	var args []any
	for _, c := range q.Clauses {
		column, ok := queryColumns[c.Field]
		if !ok {
			return nil, fmt.Errorf("library: no column mapped for query field %q", c.Field)
		}
		if numericQueryFields[c.Field] {
			n, err := strconv.Atoi(c.Value)
			if err != nil {
				return nil, fmt.Errorf("library: field %q expects a number, got %q", c.Field, c.Value)
			}
			args = append(args, n)
		} else {
			args = append(args, c.Value)
		}
		conditions = append(conditions, column+" = ?")
	}

	sqlStr := selectRecordsSQL
	if len(conditions) > 0 {
		sqlStr += " WHERE " + strings.Join(conditions, " AND ")
	}
	sqlStr += orderBySQL

	rows, err := db.sql.Query(sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("library: search: %w", err)
	}
	return scanRecords(rows)
}
