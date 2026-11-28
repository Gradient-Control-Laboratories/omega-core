package fix_test

import (
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/cmd/fix/internal/fix"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

func TestDomainFixer_Registry(t *testing.T) {
	f, ok := fix.ByName("domain")
	require.True(t, ok)
	assert.True(t, f.Doc != "")
	assert.True(t, f.Help != "")
}

func TestDomainFixer_Conversions(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "qualified call",
			in:   "return remap(sphere(1), map.rotateXY(p, 0.5));\n",
			want: "return remap(sphere(1), domain.rotateXY(p, 0.5));\n",
		},
		{
			name: "qualified value and nested calls",
			in:   "m = map.patternPolarXY(6);\nreturn remap(remap(box(-1, 1), map.checkerXY(2)), m);\n",
			want: "m = domain.patternPolarXY(6);\nreturn remap(remap(box(-1, 1), domain.checkerXY(2)), m);\n",
		},
		{
			name: "plain import",
			in:   "import \"map\";\nreturn remap(sphere(1), map.rotateZX(p, 1));\n",
			want: "import \"domain\";\nreturn remap(sphere(1), domain.rotateZX(p, 1));\n",
		},
		{
			name: "aliased import keeps its alias",
			in:   "import m \"map\";\nreturn remap(sphere(1), m.rotateZX(p, 1));\n",
			want: "import m \"domain\";\nreturn remap(sphere(1), m.rotateZX(p, 1));\n",
		},
		{
			name: "import named map follows the rename",
			in:   "import map \"map\";\nreturn remap(sphere(1), map.rotateXY(p, 1));\n",
			want: "import domain \"domain\";\nreturn remap(sphere(1), domain.rotateXY(p, 1));\n",
		},
		{
			name: "dot import",
			in:   "import . \"map\";\nreturn remap(sphere(1), rotateZX(p, 1));\n",
			want: "import . \"domain\";\nreturn remap(sphere(1), rotateZX(p, 1));\n",
		},
		{
			name: "CRLF line endings",
			in:   "a = map.triangleWave(p.x, 2);\r\nreturn sphere(a);\r\n",
			want: "a = domain.triangleWave(p.x, 2);\r\nreturn sphere(a);\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed, err := fix.Apply(tt.in, []fix.Fixer{mustFixer(t, "domain")})
			require.NoError(t, err)
			assert.True(t, changed)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDomainFixer_LeavesOtherMapsAlone(t *testing.T) {
	for _, src := range []string{
		// The map(fn, list) primitive.
		"xs = map((x) { return x * 2; }, [1, 2, 3]);\nreturn sphere(xs[0]);\n",
		// Comments and strings.
		"// map.rotateXY is old\nreturn sphere(1);\n",
		"/* map.rotateXY */ return sphere(1);\n",
		"t = text(\"map.rotateXY\");\nreturn t;\n",
		// Another package imported under the name map.
		"import map \"lattice\";\nreturn map.cubic(1);\n",
		// A string that is not an import path.
		"t = text(\"map\");\nreturn t;\n",
	} {
		got, changed, err := fix.Apply(src, []fix.Fixer{mustFixer(t, "domain")})
		require.NoError(t, err)
		assert.False(t, changed)
		assert.Equal(t, src, got)
	}
}

func mustFixer(t *testing.T, name string) fix.Fixer {
	t.Helper()
	f, ok := fix.ByName(name)
	require.True(t, ok)
	return f
}
