// Package query implements lomax's small `field:value` query language, used
// by `lomax query` (Milestone 4) and, per docs/music-cli-plan.md section
// 11, intended for reuse as sync profiles' `filter` expressions
// (Milestone 7) — kept schema-agnostic here so it isn't tied to the
// library database's particular tables/columns.
package query

import (
	"fmt"
	"strings"
)

// Clause is one "field:value" term. Clauses within a Query are ANDed.
//
// Value is taken verbatim from its shell argument, so a value containing
// spaces is written as a single shell token, e.g. `artist:"David Bowie"` —
// the shell strips the quotes and passes one argument, "artist:David
// Bowie"; Parse never sees the quote characters. Only equality is
// supported today; comparison operators (`>=`, `>`, relative dates like
// `30d`) are anticipated by section 11's sync-profile examples but not
// needed until Milestone 7 — left as a natural extension of Value's
// grammar rather than Clause's shape.
type Clause struct {
	Field string
	Value string
}

// Query is an ordered list of ANDed clauses.
type Query struct {
	Clauses []Clause
}

// KnownFields is the set of fields Parse accepts. Exported so callers (the
// library package's Search, and future reuses) can validate against the
// same list without duplicating it, and so error messages stay consistent.
var KnownFields = map[string]bool{
	"artist":       true,
	"album_artist": true,
	"album":        true,
	"title":        true,
	"year":         true,
	"track":        true,
	"disc":         true,
}

// Parse parses args — each a "field:value" shell argument — into a Query.
// It returns an error if any argument isn't of that form, names an unknown
// field, or has an empty value.
func Parse(args []string) (*Query, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("query: at least one field:value clause is required")
	}

	q := &Query{}
	for _, arg := range args {
		field, value, found := strings.Cut(arg, ":")
		field = strings.ToLower(field)
		if !found || field == "" || value == "" {
			return nil, fmt.Errorf("query: invalid clause %q, want field:value", arg)
		}
		if !KnownFields[field] {
			return nil, fmt.Errorf("query: unknown field %q", field)
		}
		q.Clauses = append(q.Clauses, Clause{Field: field, Value: value})
	}
	return q, nil
}
