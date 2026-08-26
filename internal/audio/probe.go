package audio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

// ErrProbeUnavailable is returned by Probe when ffprobe is not on $PATH.
// Callers should treat this as "duration/bitrate unknown", not a fatal
// error — ffprobe is a runtime dependency the tool checks for and hints
// about installing (see docs/music-cli-plan.md section 7), not something
// lomax bundles.
var ErrProbeUnavailable = errors.New("ffprobe not found on PATH (install the ffmpeg package)")

// StreamInfo holds the audio stream properties Track does not carry, because
// github.com/dhowden/tag reads tag metadata only and does not decode the
// audio stream itself.
type StreamInfo struct {
	Duration     time.Duration
	BitrateKbps  int
	SampleRateHz int
	Channels     int
}

// ffprobeFormat mirrors the subset of `ffprobe -show_entries format` JSON
// output this package reads.
type ffprobeOutput struct {
	Format struct {
		Duration string `json:"duration"`
		BitRate  string `json:"bit_rate"`
	} `json:"format"`
	Streams []struct {
		SampleRate string `json:"sample_rate"`
		Channels   int    `json:"channels"`
	} `json:"streams"`
}

// Probe shells out to ffprobe to read stream duration, bitrate, sample rate,
// and channel count for path. It returns ErrProbeUnavailable if ffprobe is
// not installed; callers should degrade gracefully (e.g. display "n/a")
// rather than failing the whole command over it.
func Probe(path string) (StreamInfo, error) {
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		return StreamInfo{}, ErrProbeUnavailable
	}

	cmd := exec.Command(ffprobe,
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "format=duration,bit_rate:stream=sample_rate,channels",
		"-of", "json",
		path,
	)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return StreamInfo{}, fmt.Errorf("ffprobe %s: %w", path, err)
	}

	var out ffprobeOutput
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return StreamInfo{}, fmt.Errorf("parse ffprobe output for %s: %w", path, err)
	}

	info := StreamInfo{}
	if seconds, err := strconv.ParseFloat(out.Format.Duration, 64); err == nil {
		info.Duration = time.Duration(seconds * float64(time.Second))
	}
	if bps, err := strconv.ParseInt(out.Format.BitRate, 10, 64); err == nil {
		info.BitrateKbps = int(bps / 1000)
	}
	if len(out.Streams) > 0 {
		if rate, err := strconv.Atoi(out.Streams[0].SampleRate); err == nil {
			info.SampleRateHz = rate
		}
		info.Channels = out.Streams[0].Channels
	}
	return info, nil
}
