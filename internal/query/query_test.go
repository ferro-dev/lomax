package query

import "testing"

func TestParseSingleClause(t *testing.T) {
	q, err := Parse([]string{"artist:David Bowie"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(q.Clauses) != 1 || q.Clauses[0] != (Clause{Field: "artist", Value: "David Bowie"}) {
		t.Errorf("Parse() = %+v, unexpected", q.Clauses)
	}
}

func TestParseMultipleClausesAreANDed(t *testing.T) {
	q, err := Parse([]string{"artist:David Bowie", "year:1972"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := []Clause{{Field: "artist", Value: "David Bowie"}, {Field: "year", Value: "1972"}}
	if len(q.Clauses) != len(want) {
		t.Fatalf("Parse() = %+v, want %+v", q.Clauses, want)
	}
	for i := range want {
		if q.Clauses[i] != want[i] {
			t.Errorf("Parse()[%d] = %+v, want %+v", i, q.Clauses[i], want[i])
		}
	}
}

func TestParseIsCaseInsensitiveOnFieldName(t *testing.T) {
	q, err := Parse([]string{"Artist:Bowie"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if q.Clauses[0].Field != "artist" {
		t.Errorf("Field = %q, want lowercased %q", q.Clauses[0].Field, "artist")
	}
}

func TestParseRejectsMissingColon(t *testing.T) {
	if _, err := Parse([]string{"justabareword"}); err == nil {
		t.Error("Parse with no colon: got nil error, want an error")
	}
}

func TestParseRejectsUnknownField(t *testing.T) {
	if _, err := Parse([]string{"nonsense:value"}); err == nil {
		t.Error("Parse with an unknown field: got nil error, want an error")
	}
}

func TestParseRejectsEmptyValue(t *testing.T) {
	if _, err := Parse([]string{"artist:"}); err == nil {
		t.Error("Parse with an empty value: got nil error, want an error")
	}
}

func TestParseRejectsNoArgs(t *testing.T) {
	if _, err := Parse(nil); err == nil {
		t.Error("Parse with no arguments: got nil error, want an error")
	}
}
