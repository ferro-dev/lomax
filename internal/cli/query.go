package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"

	"github.com/ferro-dev/lomax/internal/query"
)

// newQueryCmd builds `lomax query <field:value>...`: search the library
// database. Each argument is one "field:value" clause; clauses are ANDed.
// A value with spaces is one shell token, e.g. `artist:"David Bowie"` — see
// internal/query's doc comment for why that's the right way to quote it.
func newQueryCmd() *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:     "query <field:value>...",
		Short:   `Search the library database, e.g. lomax query artist:"David Bowie" year:1972`,
		Example: `  lomax query artist:"David Bowie" year:1972`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuery(cmd, args, dbPath)
		},
	}
	addLibraryDBFlag(cmd, &dbPath)
	return cmd
}

func runQuery(cmd *cobra.Command, args []string, dbPath string) error {
	q, err := query.Parse(args)
	if err != nil {
		return err
	}

	db, err := openLibrary(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	records, err := db.Search(q)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	if len(records) == 0 {
		_, err := fmt.Fprintln(out, "no tracks matched")
		return err
	}

	t := table.New().Headers("ARTIST", "ALBUM", "#", "TITLE", "YEAR", "PATH")
	for _, r := range records {
		t.Row(r.Artist, r.Album, trackPosition(r.TrackNum, r.TrackTotal), r.Title, yearString(r.Year), r.Path)
	}
	_, err = fmt.Fprintln(out, t.Render())
	return err
}

// trackPosition formats a track number as "N" or "N/Total", matching the
// TRACK column style already used by `lomax inspect`.
func trackPosition(num, total int) string {
	if num == 0 {
		return ""
	}
	if total == 0 {
		return fmt.Sprintf("%d", num)
	}
	return fmt.Sprintf("%d/%d", num, total)
}
