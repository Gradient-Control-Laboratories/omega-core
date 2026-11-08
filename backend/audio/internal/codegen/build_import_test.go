package codegen_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/backend/audio/internal/codegen"
	"gradientcontrol.com/omega/core/backend/audio/internal/extract"
	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// TestImportConstants — top-level constants in an imported package flow
// into the generated audio output and references resolve. Regression for
// the original stub where BuildImport returned an error.
func TestImportConstants(t *testing.T) {
	srcs := map[string][]string{
		"libk": {`pi = atan(1) * 4;
two_pi = pi * 2;
`},
	}
	primary := `import . "libk";
tone = sin(time * two_pi * 440);
return audio(tone, tone);
`
	primaryAudioSrc, _, err := extract.ExtractAudio(string(primary))
	require.NoError(t, err)

	mgr := &source.MockManager{PkgSources: srcs}
	res, err := assembler.Assemble([]string{primaryAudioSrc}, assembler.SourceManager(mgr))
	require.NoError(t, err)

	js, _, err := codegen.AudioToJs(res.Program)
	require.NoError(t, err)
	// pi is folded into two_pi and dropped as dead; two_pi is what the primary reads.
	require.True(t, strings.Contains(js, "const libk_two_pi"))
	require.True(t, strings.Contains(js, "libk_two_pi"))
}

// TestImportFunctions — functions in an imported package are emitted into
// the output and references resolve through the build framework's scope
// chain.
func TestImportFunctions(t *testing.T) {
	srcs := map[string][]string{
		"audiolib": {`osc(freq) { return sin(time * freq); }
env(t) { return exp(-t * 2); }
`},
	}
	primary := `import . "audiolib";
// add remap and non-const args to prevent inlining of the imported functions
note = remap(osc(time * 440) * env(time), [1, 2, 3]);
return audio(note, note);
`
	primaryAudioSrc, _, err := extract.ExtractAudio(string(primary))
	require.NoError(t, err)

	mgr := &source.MockManager{PkgSources: srcs}
	res, err := assembler.Assemble([]string{primaryAudioSrc}, assembler.SourceManager(mgr))
	require.NoError(t, err)

	js, _, err := codegen.AudioToJs(res.Program)
	require.NoError(t, err)
	require.True(t, strings.Contains(js, "function audiolib_osc("))
	require.True(t, strings.Contains(js, "function audiolib_env("))
}

// TestImportCollisionResolved — two imports declaring the same identifier
// don't trample each other in the flat JS namespace; the pkg path prefixes
// keep them distinct. Without the prefixing fix this generated two
// `function helper()` definitions and the second silently won.
func TestImportCollisionResolved(t *testing.T) {
	srcs := map[string][]string{
		"a": {`helper() { return p.x + 1; }`},
		"b": {`helper() { return p.x + 2; }`},
	}
	primary := `import a "a";
	import b "b";
both = a.helper() + b.helper();
return audio(both, both);
`
	primaryAudioSrc, _, err := extract.ExtractAudio(string(primary))
	require.NoError(t, err)

	mgr := &source.MockManager{PkgSources: srcs}
	res, err := assembler.Assemble([]string{primaryAudioSrc}, assembler.SourceManager(mgr))
	require.NoError(t, err)

	js, _, err := codegen.AudioToJs(res.Program)
	require.NoError(t, err)
	// Assert the sum of the two inlined helpers
	require.True(t, strings.Contains(js, "gsl_p[0] + 1 + (gsl_p[0] + 2)"))
}
