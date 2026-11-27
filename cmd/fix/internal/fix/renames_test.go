package fix_test

import (
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/cmd/fix/internal/fix"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

func TestStdlibFixer_RenamedMembers(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "default prefix",
			in:   "return honeycomb.squareHoneycomb([2, 2]) - 0.1;\n",
			want: "return honeycomb.rectangular([2, 2]) - 0.1;\n",
		},
		{
			name: "every honeycomb rename, longest name first",
			in:   "a = honeycomb.squareHoneycombReinforced([2, 2], 0.5);\nb = honeycomb.squareDiagonalHoneycomb([2, 2]);\nc = honeycomb.hexagonalHoneycomb(2);\nd = honeycomb.triangularHoneycomb(2);\ne = honeycomb.octetHoneycomb([2, 2]);\nreturn min(a, b, c, d, e);\n",
			want: "a = honeycomb.rectangularReinforced([2, 2], 0.5);\nb = honeycomb.rectangularDiagonal([2, 2]);\nc = honeycomb.hexagonal(2);\nd = honeycomb.triangular(2);\ne = honeycomb.octet([2, 2]);\nreturn min(a, b, c, d, e);\n",
		},
		{
			name: "explicit import name",
			in:   "import h \"lattice/honeycomb\";\nreturn h.hexagonalHoneycomb(2, unionRound(0.2));\n",
			want: "import h \"lattice/honeycomb\";\nreturn h.hexagonal(2, unionRound(0.2));\n",
		},
		{
			name: "dot import",
			in:   "import . \"lattice/honeycomb\";\nreturn triangularHoneycomb(2);\n",
			want: "import . \"lattice/honeycomb\";\nreturn triangular(2);\n",
		},
		{
			name: "matrix.scale",
			in:   "m = matrix.scale(2, matrix.identity3());\nreturn sphere(m[0][0]);\n",
			want: "m = matrix.multiplyScalar(2, matrix.identity(3));\nreturn sphere(m[0][0]);\n",
		},
		{
			name: "sized matrix names",
			in:   "a = matrix.identity3();\nb = matrix.transpose4(matrix.identity4());\nreturn sphere(matrix.determinant3(a) + matrix.inverse2([[2, 0], [0, 2]])[0][0] + b[0][0]);\n",
			want: "a = matrix.identity(3);\nb = matrix.transpose(matrix.identity(4));\nreturn sphere(matrix.determinant(a) + matrix.inverse([[2, 0], [0, 2]])[0][0] + b[0][0]);\n",
		},
		{
			name: "passed as a value",
			in:   "f = honeycomb.octetHoneycomb;\nreturn f([2, 2]);\n",
			want: "f = honeycomb.octet;\nreturn f([2, 2]);\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed, err := fix.Apply(tt.in, []fix.Fixer{mustFixer(t, "stdlib")})
			require.NoError(t, err)
			assert.True(t, changed)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStdlibFixer_RenamedMembersLeftAlone(t *testing.T) {
	for _, src := range []string{
		// Not dot-imported: a bare name is the file's own.
		"squareHoneycomb(s) { return s; }\nreturn sphere(squareHoneycomb(1));\n",
		// Another package's member of the same name.
		"return lattice.octetHoneycomb(2);\n",
		// The prefix names another package.
		"import honeycomb \"lattice/foam\";\nreturn honeycomb.octetHoneycomb(2);\n",
		// Comments and strings.
		"// honeycomb.squareHoneycomb\nreturn sphere(1);\n",
		"t = text(\"honeycomb.squareHoneycomb\");\nreturn t;\n",
	} {
		got, changed, err := fix.Apply(src, []fix.Fixer{mustFixer(t, "stdlib")})
		require.NoError(t, err)
		assert.False(t, changed)
		assert.Equal(t, src, got)
	}
}

func TestStdlibFixer_DeprecatedLatticeBcc(t *testing.T) {
	tests := []struct{ in, want string }{
		{"return lattice.bcc(10, 1);\n", "return (beam.bcc(10) - 1);\n"},
		{"return lattice.bcc([10, 12, 8], r, unionRound(0.5));\n", "return (beam.bcc([10, 12, 8], unionRound(0.5)) - r);\n"},
		{"return lattice.bcc(10, r + 0.5) * 2;\n", "return (beam.bcc(10) - (r + 0.5)) * 2;\n"},
		{"import l \"lattice\";\nreturn l.bcc(10, 1);\n", "import l \"lattice\";\nreturn (beam.bcc(10) - 1);\n"},
	}
	for _, tt := range tests {
		got, changed, err := fix.Apply(tt.in, []fix.Fixer{mustFixer(t, "stdlib")})
		require.NoError(t, err)
		assert.True(t, changed)
		assert.Equal(t, tt.want, got)
	}
	// beam.bcc and foam.bcc are not the deprecated one.
	for _, src := range []string{"return beam.bcc(10) - 1;\n", "return foam.bcc([10, 10, 10]);\n"} {
		got, changed, err := fix.Apply(src, []fix.Fixer{mustFixer(t, "stdlib")})
		require.NoError(t, err)
		assert.False(t, changed)
		assert.Equal(t, src, got)
	}
}
