// Command lomax is a Linux-native CLI music library manager.
//
// This is the M0 bootstrap entrypoint: it wires up the root command and
// exits non-zero on error. Feature commands land in later milestones.
package main

import (
	"fmt"
	"os"

	"github.com/ferro-dev/lomax/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
