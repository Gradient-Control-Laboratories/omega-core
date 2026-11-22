package fix_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/cmd/fix/internal/fix"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestRangeFixer_Registered(t *testing.T) {
	rangeFixer, ok := fix.ByName("range")
	require.True(t, ok)
	assert.Equal(t, "range", rangeFixer.Name)
	assert.True(t, rangeFixer.Doc != "")
	assert.True(t, rangeFixer.Help != "")
}

func TestRangeFixer_Conversions(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "min and max share the assignment's line",
			in:   "r = 5;\nr.min = 1;\nr.max = 10;\nreturn sphere(r);\n",
			want: "@range(1, 10) r = 5;\nreturn sphere(r);\n",
		},
		{
			name: "period becomes the third argument, whatever order the properties were written in",
			in:   "r = 5;\nr.period = 4;\nr.max = 10;\nr.min = 1;\nreturn sphere(r);\n",
			want: "@range(1, 10, 4) r = 5;\nreturn sphere(r);\n",
		},
		{
			name: "bounds that read other variables are copied whole, and a long annotation takes the line above",
			in: "slideRadius = 60.00;\nbaseRadius = 40.00;\nslideOffset = 40;\n" +
				"slidePos = 20.2;\n" +
				"slidePos.min = abs(slideRadius - baseRadius) - slideOffset;\n" +
				"slidePos.max = slideRadius + baseRadius - slideOffset;\n" +
				"slidePos.period = 10;\n" +
				"return sphere(slidePos);\n",
			want: "slideRadius = 60.00;\nbaseRadius = 40.00;\nslideOffset = 40;\n" +
				"@range(abs(slideRadius - baseRadius) - slideOffset, slideRadius + baseRadius - slideOffset, 10)\n" +
				"slidePos = 20.2;\n" +
				"return sphere(slidePos);\n",
		},
		{
			name: "a bound that reads a variable declared between the assignment and its property",
			in:   "x = 5;\nlo = 1;\nhi = 10 + 2;\nx.min = lo;\nx.max = hi * 2;\nreturn sphere(x);\n",
			want: "@range(lo, hi * 2) x = 5;\nlo = 1;\nhi = 10 + 2;\nreturn sphere(x);\n",
		},
		{
			name: "a comment after a removed property stays on its own line",
			in:   "h = 2.5;\nh.min = 0.5; // lowest\nh.max = 8; // highest\nh.period = 4;\nreturn sphere(h);\n",
			want: "@range(0.5, 8, 4) h = 2.5;\n// lowest\n// highest\nreturn sphere(h);\n",
		},
		{
			name: "properties between two blank lines leave one",
			in:   "w = 3;\n\nw.min = 1;\nw.max = 6;\n\nreturn sphere(w);\n",
			want: "@range(1, 6) w = 3;\n\nreturn sphere(w);\n",
		},
		{
			name: "properties that end a section keep the blank line that follows",
			in:   "r = 5;\nr.min = 1;\nr.max = 10;\n\n// next\nreturn sphere(r);\n",
			want: "@range(1, 10) r = 5;\n\n// next\nreturn sphere(r);\n",
		},
		{
			name: "statements that share a line with other statements take only themselves",
			in:   "a = 1; b = 2; a.min = 0; a.max = 4; b.min = 1;\nreturn sphere(a + b);\n",
			want: "@range(0, 4) a = 1; b = 2; b.min = 1;\nreturn sphere(a + b);\n",
		},
		{
			name: "statements alone on a line with each other take the line",
			in:   "c = 7;\nc.min = 0; c.max = 9;\nreturn sphere(c);\n",
			want: "@range(0, 9) c = 7;\nreturn sphere(c);\n",
		},
		{
			name: "other properties of the variable stay",
			in:   "s = 3;\ns.min = 1;\ns.color = [1, 0, 0];\ns.max = 5;\nreturn sphere(s);\n",
			want: "@range(1, 5) s = 3;\ns.color = [1, 0, 0];\nreturn sphere(s);\n",
		},
		{
			name: "a bound over several lines is copied as written",
			in:   "r = 1;\nh = 2;\nt = 4;\nt.min = abs(r -\n            h);\nt.max = r + h;\nreturn sphere(t);\n",
			want: "r = 1;\nh = 2;\n@range(abs(r -\n            h), r + h) t = 4;\nreturn sphere(t);\n",
		},
		{
			name: "CRLF line endings are kept",
			in:   "r = 5;\r\nr.min = 1;\r\nr.max = 10;\r\nreturn sphere(r);\r\n",
			want: "@range(1, 10) r = 5;\r\nreturn sphere(r);\r\n",
		},
		{
			name: "each of several variables converts",
			in:   "a = 1;\na.min = 0;\na.max = 4;\nb = 2;\nb.min = 1;\nb.max = 9;\nreturn sphere(a + b);\n",
			want: "@range(0, 4) a = 1;\n@range(1, 9) b = 2;\nreturn sphere(a + b);\n",
		},
	}
	rangeFixer, _ := fix.ByName("range")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, changed, err := rangeFixer.Apply(tc.in)
			require.NoError(t, err)
			assert.True(t, changed)
			assert.Equal(t, tc.want, got)

			// A second run finds nothing left to rewrite.
			again, changedAgain, err := rangeFixer.Apply(got)
			require.NoError(t, err)
			assert.False(t, changedAgain)
			assert.Equal(t, got, again)
		})
	}
}

func TestRangeFixer_LeavesAlone(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"no properties", "r = 5;\nreturn sphere(r);\n"},
		{"an explicit @range already", "@range(0, 9) u = 3;\nu.min = 1;\nreturn sphere(u);\n"},
		{"only a minimum", "q = 1;\nq.min = 0;\nreturn sphere(q);\n"},
		{"only a period", "q = 1;\nq.period = 4;\nreturn sphere(q);\n"},
		{"a variable that shares its assignment", "m, n = [1, 2];\nm.min = 0;\nm.max = 3;\nreturn sphere(m + n);\n"},
		{"a comment inside a bound", "v = 1;\nv.min = 0 /* low */;\nv.max = 3;\nreturn sphere(v);\n"},
		{"source that does not parse", "r = ;\nr.min = 1;\nr.max = 10;\n"},
		{"an empty file", ""},
	}
	rangeFixer, _ := fix.ByName("range")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, changed, err := rangeFixer.Apply(tc.in)
			require.NoError(t, err)
			assert.False(t, changed)
			assert.Equal(t, tc.in, got)
		})
	}
}

// A file with some variables that convert and some that do not converts the
// first and leaves the rest exactly as they were.
func TestRangeFixer_MixedFile(t *testing.T) {
	in := strings.Join([]string{
		"r = 5;",
		"r.min = 1;",
		"r.max = 10;",
		"@range(0, 9) u = 3;",
		"u.min = 1;",
		"q = 1;",
		"q.min = 0;",
		"return sphere(r + u + q);",
		"",
	}, "\n")
	want := strings.Join([]string{
		"@range(1, 10) r = 5;",
		"@range(0, 9) u = 3;",
		"u.min = 1;",
		"q = 1;",
		"q.min = 0;",
		"return sphere(r + u + q);",
		"",
	}, "\n")
	rangeFixer, _ := fix.ByName("range")
	got, changed, err := rangeFixer.Apply(in)
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, want, got)
}
