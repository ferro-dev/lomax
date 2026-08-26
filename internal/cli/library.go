package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ferro-dev/lomax/internal/library"
)

// libraryDBFlag is the flag name shared by every command that opens the
// library database, so help text and behavior stay consistent.
const libraryDBFlag = "library-db"

// addLibraryDBFlag registers --library-db on cmd, binding it to path.
func addLibraryDBFlag(cmd *cobra.Command, path *string) {
	cmd.Flags().StringVar(path, libraryDBFlag, "", "path to the library database (default: $LOMAX_STATE_DIR/library.db, or the XDG state dir)")
}

// openLibrary opens the library database at path, or the default location
// if path is empty (see library.DefaultPath).
func openLibrary(path string) (*library.DB, error) {
	if path == "" {
		var err error
		path, err = library.DefaultPath()
		if err != nil {
			return nil, err
		}
	}
	db, err := library.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open library database: %w", err)
	}
	return db, nil
}
