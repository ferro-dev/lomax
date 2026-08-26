// Package acoustid fingerprints audio files via the Chromaprint `fpcalc`
// binary and looks matches up against the AcoustID web service, for
// resolving metadata on files with missing or unreliable tags.
package acoustid

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os/exec"
)

// ErrFpcalcUnavailable is returned by Fingerprint when fpcalc is not on
// $PATH. Like ffprobe (see internal/audio/probe.go), fpcalc is a runtime
// dependency lomax checks for rather than bundles — callers should degrade
// gracefully (skip the AcoustID lookup) rather than failing outright.
var ErrFpcalcUnavailable = errors.New("fpcalc not found on PATH (install the chromaprint/libchromaprint-tools package)")

// Fingerprint is a Chromaprint audio fingerprint plus the duration fpcalc
// measured it over. The AcoustID lookup API requires both.
type Fingerprint struct {
	DurationSeconds int
	Data            string
}

// fpcalcOutput mirrors `fpcalc -json`'s output.
type fpcalcOutput struct {
	Duration    float64 `json:"duration"`
	Fingerprint string  `json:"fingerprint"`
}

// Compute runs fpcalc against path and returns its fingerprint and duration.
func Compute(path string) (Fingerprint, error) {
	fpcalc, err := exec.LookPath("fpcalc")
	if err != nil {
		return Fingerprint{}, ErrFpcalcUnavailable
	}

	cmd := exec.Command(fpcalc, "-json", path)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return Fingerprint{}, fmt.Errorf("acoustid: fpcalc %s: %w", path, err)
	}

	var out fpcalcOutput
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return Fingerprint{}, fmt.Errorf("acoustid: parse fpcalc output for %s: %w", path, err)
	}
	return Fingerprint{
		DurationSeconds: int(math.Round(out.Duration)),
		Data:            out.Fingerprint,
	}, nil
}
