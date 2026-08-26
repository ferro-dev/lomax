package cli

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"

	"github.com/ferro-dev/lomax/internal/audio"
)

// newInspectCmd builds `lomax inspect <path>`: read tags (and, where
// ffprobe is available, stream properties) for one file or every audio file
// under a directory, and render them as a table.
func newInspectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <path>",
		Short: "Show tags and stream info for one file or a directory of audio files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInspect(cmd, args[0])
		},
	}
}

func runInspect(cmd *cobra.Command, path string) error {
	files, err := scanPathForAudio("inspect", path)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "no audio files found under %s\n", path)
		return err
	}

	t := table.New().
		Headers("FILE", "FORMAT", "ARTIST", "ALBUM", "TITLE", "TRACK", "YEAR", "DURATION", "BITRATE")

	for _, file := range files {
		track, err := audio.ReadTrack(file)
		if err != nil {
			// One unreadable file shouldn't abort the whole inspection —
			// report it inline and keep going.
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", err)
			continue
		}

		duration, bitrate := "n/a", "n/a"
		if stream, err := audio.Probe(file); err == nil {
			if stream.Duration > 0 {
				duration = stream.Duration.Round(time.Second).String()
			}
			if stream.BitrateKbps > 0 {
				bitrate = fmt.Sprintf("%d kbps", stream.BitrateKbps)
			}
		}

		trackNum := ""
		if track.TrackNum > 0 {
			trackNum = fmt.Sprintf("%d", track.TrackNum)
			if track.TrackTotal > 0 {
				trackNum = fmt.Sprintf("%d/%d", track.TrackNum, track.TrackTotal)
			}
		}

		year := ""
		if track.Year > 0 {
			year = fmt.Sprintf("%d", track.Year)
		}

		t.Row(
			track.Path,
			string(track.FileType),
			track.Artist,
			track.Album,
			track.Title,
			trackNum,
			year,
			duration,
			bitrate,
		)
	}

	_, err = fmt.Fprintln(cmd.OutOrStdout(), t.Render())
	return err
}
