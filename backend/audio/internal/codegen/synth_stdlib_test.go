package codegen_test

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/assembler/variation"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/audio/internal/codegen"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// synthEveryVoice calls every drum voice and FM patch in the synth stdlib, so
// all of them go through the audio backend. It also passes a named voice
// function that calls another package (keyVoice) to the sequencer.
const synthEveryVoice = `import "synth";
import "synth/seq";
import "synth/drum";
import "synth/fm";
clk = seq.clock(time, 120, 4, 0.55);
h = [time, 1, 0, 57, 0.5];
drums = drum.kick(h) + drum.snare(h) + drum.clap(h) + drum.hatClosed(h) + drum.hatOpen(h) + drum.cymbal(h) + drum.tomLow(h) + drum.tomMid(h) + drum.tomHigh(h) + drum.rim(h) + drum.cowbell(h);
keys = fm.ePiano(h) + fm.bass(h) + fm.bell(h) + fm.brass(h) + fm.marimba(h) + fm.pad(h) + fm.organ(h);
tones = synth.saw(110, time) + synth.squareWave(110, time) + synth.tri(110, time) + synth.modes(220, time, [[1, 1, 1], [2.7, 0.5, 0.4]]) + synth.additive(55, time, [1, 0.5, 0.25]);
keyVoice(x) {
  return fm.ePiano(x);
}
groove = seq.play(clk, [9, 0, 5, 0], (x) {
  return drum.kick(x);
}, 1) + seq.playNotes(clk, [[45, 9, 2], seq.rest, [48, 7, 0]], (x) {
  return synth.saw(synth.mtof(x.w), x.x) * synth.adsr(x.x, x[4], 0.01, 0.1, 0.5, 0.1);
});
named = seq.playNotes(clk, [[57, 7, 2], seq.rest], keyVoice);
return audio(synth.softclip(drums + keys + tones + groove + named, 1));
`

func compileAudio(t *testing.T, src string) string {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
	require.NoError(t, err)
	require.True(t, len(res.AudioReturns) > 0)
	ops := make([]*program.Operand, len(res.AudioReturns))
	for i, ar := range res.AudioReturns {
		ops[i] = ar.Return
	}
	js, _, err := codegen.AudioToJs(variation.Program(res.Program, ops...))
	require.NoError(t, err)
	return js
}

// checkJS fails on defects the browser would only report at load or run
// time: two JS functions with one name (the later silently replaces the
// earlier for every caller), a const declared twice in one function (a
// SyntaxError), and calls to gsl_* runtime helpers that were never emitted.
func checkJS(t *testing.T, js string) {
	t.Helper()
	fnDecl := regexp.MustCompile(`^function ([A-Za-z_0-9$]+)\(`)
	constDecl := regexp.MustCompile(`^\s*const ([A-Za-z_0-9$]+) =`)
	fns := map[string]bool{}
	consts := map[string]bool{}
	for _, line := range strings.Split(js, "\n") {
		if m := fnDecl.FindStringSubmatch(line); m != nil {
			require.False(t, fns[m[1]], "duplicate JS function %s", m[1])
			fns[m[1]] = true
			consts = map[string]bool{}
			continue
		}
		if m := constDecl.FindStringSubmatch(line); m != nil {
			require.False(t, consts[m[1]], "const %s declared twice in one function", m[1])
			consts[m[1]] = true
		}
	}
	for _, m := range regexp.MustCompile(`\b(gsl_[A-Za-z0-9]+)\(`).FindAllStringSubmatch(js, -1) {
		require.True(t, fns[m[1]], "call to undefined helper %s", m[1])
	}
}

// TestSynthStdlibCompiles compiles every voice of the synth stdlib to audio JS.
func TestSynthStdlibCompiles(t *testing.T) {
	checkJS(t, compileAudio(t, synthEveryVoice))
}

// TestDrumMachineExampleCompiles compiles examples/drum_machine.gsl to audio JS.
func TestDrumMachineExampleCompiles(t *testing.T) {
	src, err := os.ReadFile("../../../../examples/drum_machine.gsl")
	require.NoError(t, err)
	checkJS(t, compileAudio(t, string(src)))
}
