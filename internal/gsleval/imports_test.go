package gsleval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

func TestStripRedundantLeadingImports(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "semicolon-less builtin import is dropped without eating the next statement",
			in:   "import \"matrix\"\nslab = 1.0;",
			want: "\nslab = 1.0;",
		},
		{
			name: "well-formed builtin import is dropped",
			in:   "import \"matrix\";\nslab = 1.0;",
			want: "\nslab = 1.0;",
		},
		{
			name: "omega import is preserved, matrix dropped",
			in:   "import \"omega\";\nimport \"matrix\";\nx = 1.0;",
			want: "import \"omega\";\n\nx = 1.0;",
		},
		{
			name: "no imports is a no-op",
			in:   "slab = length(p) - 5.0;\nreturn slab;",
			want: "slab = length(p) - 5.0;\nreturn slab;",
		},
		{
			name: "leading comment is preserved when the import is dropped",
			in:   "// hello\nimport \"vector\";\nx = 1.0;",
			want: "// hello\n\nx = 1.0;",
		},
		{
			name: "an identifier that starts with import is not touched",
			in:   "importScale = 2.0;\nreturn importScale;",
			want: "importScale = 2.0;\nreturn importScale;",
		},
	}
	for _, tc := range cases {
		if got := stripRedundantLeadingImports(tc.in); got != tc.want {
			t.Errorf("%s:\n  in   %q\n  got  %q\n  want %q", tc.name, tc.in, got, tc.want)
		}
	}
}

// TestEvalGslRecoversFromStrayImport is the gemini-class recovery: a
// redundant, semicolon-less built-in import used to fail the query parse; it is
// now stripped and the body evaluates.
func TestEvalGslRecoversFromStrayImport(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	prog := "import \"matrix\"\nslab = length(p) - 5.0;\nreturn slab;"
	groups, err := e.EvalGsl(context.Background(), nil, prog, "", []string{"slab"}, [][3]float64{{0, 0, 0}})
	if err != nil {
		t.Fatalf("expected recovery from the stray import, got: %v", err)
	}
	if len(groups) == 0 || len(groups[0].Samples) == 0 || len(groups[0].Samples[0]) == 0 {
		t.Fatal("no sample returned")
	}
	if got := float64(groups[0].Samples[0][0].Value); math.Abs(got+5.0) > 1e-4 {
		t.Errorf("slab at origin = %v, want -5 (length(0)-5)", got)
	}
}
