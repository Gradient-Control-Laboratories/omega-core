package fix_test

import (
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/cmd/fix/internal/fix"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
	"github.com/Gradient-Control-Laboratories/omega-core/parser"
)

func TestOmegaImportFixer_Registry(t *testing.T) {
	f, ok := fix.ByName("omegaimport")
	require.True(t, ok)
	assert.Equal(t, "omegaimport", f.Name)
	assert.True(t, f.Doc != "")
	assert.True(t, f.Help != "")
}

func TestOmegaImportFixer_Conversions(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		want        string
		wantChanged bool
	}{
		{
			name:        "removes standalone top-of-file import and trailing blank line",
			in:          "import . \"omega\";\n\nreturn sphere(1);\n",
			want:        "return sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "removes standalone top-of-file import without blank line",
			in:          "import . \"omega\";\nreturn sphere(1);\n",
			want:        "return sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "removes indented and spaced import . \"omega\";",
			in:          "  import   .   \"omega\"  ;  \n\nreturn sphere(1);\n",
			want:        "return sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "preserves header comment and collapses redundant blank lines",
			in:          "// Part header\n\nimport . \"omega\";\n\nreturn sphere(1);\n",
			want:        "// Part header\n\nreturn sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "removes import . \"omega\" before other imports",
			in:          "import . \"omega\";\nimport \"color\";\n\nreturn sphere(1);\n",
			want:        "import \"color\";\n\nreturn sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "removes import . \"omega\" after other imports",
			in:          "import \"color\";\nimport . \"omega\";\n\nreturn sphere(1);\n",
			want:        "import \"color\";\n\nreturn sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "preserves trailing inline comment on same line",
			in:          "import . \"omega\"; // legacy note\nreturn sphere(1);\n",
			want:        "// legacy note\nreturn sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "removes import . \"omega\" sharing a line with another import",
			in:          "import . \"omega\"; import \"color\";\nreturn sphere(1);\n",
			want:        "import \"color\";\nreturn sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "removes trailing import . \"omega\" after another import on same line",
			in:          "import \"color\"; import . \"omega\";\nreturn sphere(1);\n",
			want:        "import \"color\";\nreturn sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "removes middle import . \"omega\" between two imports on same line",
			in:          "import \"color\"; import . \"omega\"; import \"map\";\nreturn sphere(1);\n",
			want:        "import \"color\"; import \"map\";\nreturn sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "removes import . \"omega\" at end of file after blank line",
			in:          "@scene(20)\n\nimport . \"omega\";\n",
			want:        "@scene(20)\n",
			wantChanged: true,
		},
		{
			name:        "removes import . \"omega\" at end of file without trailing newline",
			in:          "@scene(20)\nimport . \"omega\";",
			want:        "@scene(20)",
			wantChanged: true,
		},
		{
			name:        "removes sole import . \"omega\" line leaving empty file",
			in:          "import . \"omega\";\n",
			want:        "",
			wantChanged: true,
		},
		{
			name:        "empty file is a no-op",
			in:          "",
			want:        "",
			wantChanged: false,
		},
		{
			name:        "removes multiline import . \"omega\"; statement",
			in:          "import .\n  \"omega\";\n\nreturn sphere(1);\n",
			want:        "return sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "removes duplicate import . \"omega\" and resolves conflict with import \"vector\"",
			in:          "import \"vector\";\nimport . \"omega\";\nimport . \"omega\";\n\nreturn sphere(1);\n",
			want:        "import \"vector\";\n\nreturn sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "removes import . \"omega\" even when trailing declaration has syntax error",
			in:          "import . \"omega\";\n\nx = 10;\nbroken = ;\n",
			want:        "x = 10;\nbroken = ;\n",
			wantChanged: true,
		},
		{
			name:        "preserves CRLF line endings",
			in:          "@scene(20)\r\n\r\nimport . \"omega\";\r\n\r\nreturn sphere(1);\r\n",
			want:        "@scene(20)\r\n\r\nreturn sphere(1);\r\n",
			wantChanged: true,
		},
		{
			name:        "fixes both legacy Scene Size comments and import . \"omega\" in one pass",
			in:          "// Scene Size: 20\n// Scene Grid: 10\nimport . \"omega\";\n\nreturn sphere(1);\n",
			want:        "@scene(20)\n\nreturn sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "does not touch line comments, block comments, or string literals",
			in:          "// import . \"omega\";\n/* import . \"omega\"; */\ns = [\"import . \\\"omega\\\";\"];\nreturn sphere(1);\n",
			want:        "// import . \"omega\";\n/* import . \"omega\"; */\ns = [\"import . \\\"omega\\\";\"];\nreturn sphere(1);\n",
			wantChanged: false,
		},
		{
			name:        "does not touch non-dot import \"omega\" or aliased import",
			in:          "import \"omega\";\nreturn sphere(1);\n",
			want:        "import \"omega\";\nreturn sphere(1);\n",
			wantChanged: false,
		},
		{
			name:        "does not touch aliased omega import or dot import of other packages",
			in:          "import o \"omega\";\nimport . \"vector\";\nreturn sphere(1);\n",
			want:        "import o \"omega\";\nimport . \"vector\";\nreturn sphere(1);\n",
			wantChanged: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, changed, err := fix.Apply(tc.in, nil)
			require.NoError(t, err)
			assert.Equal(t, tc.wantChanged, changed)
			assert.Equal(t, tc.want, got)

			// Verify second run is strictly idempotent.
			got2, changed2, err2 := fix.Apply(got, nil)
			require.NoError(t, err2)
			assert.False(t, changed2)
			assert.Equal(t, got, got2)
		})
	}
}

func TestApply_CombinedSceneAndOmegaImportProducesZeroParserWarnings(t *testing.T) {
	orig := "// Scene Size: 40\n// Scene Grid: 8\nimport . \"omega\";\n\nreturn sphere(5);\n"

	origRes, err := parser.ParsePackage([]string{orig})
	require.NoError(t, err)
	assert.Equal(t, 2, len(origRes.Diagnostics))

	fixed, changed, err := fix.Apply(orig, nil)
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, "@scene(40, 5)\n\nreturn sphere(5);\n", fixed)

	fixedRes, err := parser.ParsePackage([]string{fixed})
	require.NoError(t, err)
	assert.Equal(t, 0, len(fixedRes.Diagnostics))
	require.True(t, fixedRes.Package.Scene != nil)
	assert.Equal(t, 40.0, fixedRes.Package.Scene.Size)
	assert.Equal(t, 5.0, fixedRes.Package.Scene.Grid)
}
