package main_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/cmd/compile/internal/cli"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

// A program whose audio return goes through library code compiles to the
// same audio JavaScript the viewer plays: computeAudio, built from the audio
// returns only.
func TestRun_Audio_LibraryProgram(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "tone.gsl")
	src := "import \"synth\";\nreturn audio(synth.saw(110, time) * synth.decay(time, 0.5));\n"
	require.NoError(t, os.WriteFile(source, []byte(src), 0644))
	out := filepath.Join(dir, "tone.js")
	var stdout, stderr bytes.Buffer
	require.NoError(t, cli.Run([]string{"-t", "audio", "--no-meta", "-o", out, source}, nil, &stdout, &stderr))
	js, err := os.ReadFile(out)
	require.NoError(t, err)
	require.True(t, strings.Contains(string(js), "function computeAudio("), "no computeAudio in:\n%s", js)
}
