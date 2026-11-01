package fix_test

import (
	"testing"

	"gradientcontrol.com/omega/core/cmd/fix/internal/fix"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestRegistry(t *testing.T) {
	all := fix.All()
	require.True(t, len(all) >= 1)

	sceneFixer, ok := fix.ByName("scene")
	require.True(t, ok)
	assert.Equal(t, "scene", sceneFixer.Name)
	assert.True(t, sceneFixer.Doc != "")
	assert.True(t, sceneFixer.Help != "")

	_, ok = fix.ByName("nonexistent")
	assert.False(t, ok)
}

func TestSceneFixer_Conversions(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		want        string
		wantChanged bool
	}{
		{
			name:        "division count equal to 10 emits single-argument @scene(size)",
			in:          "// Scene Size: 20\n// Scene Grid: 10\n\nreturn sphere(1);\n",
			want:        "@scene(20)\n\nreturn sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "scene size only without grid emits @scene(size)",
			in:          "// Scene Size: 40\nreturn sphere(1);\n",
			want:        "@scene(40)\nreturn sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "explicit mm grid equal to size/10 emits single-argument @scene(size)",
			in:          "// Scene Size: 50 mm\n// Scene Grid: 5 mm\n\nreturn sphere(1);\n",
			want:        "@scene(50)\n\nreturn sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "explicit mm grid different from size/10 emits @scene(size, grid)",
			in:          "// Scene Size: 50.5\n// Scene Grid: 5 mm\n\nreturn sphere(1);\n",
			want:        "@scene(50.5, 5)\n\nreturn sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "division count different from 10 emits @scene(size, size/count)",
			in:          "// Scene Size: 100\n// Scene Grid: 20\n\nreturn sphere(1);\n",
			want:        "@scene(100, 5)\n\nreturn sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "preserves non-scene header comments and blank lines",
			in:          "// Work Order: WO-1\n// Material: Aluminum\n// Scene Size: 60 mm\n// Scene Grid: 15 mm\n\nx = 1;\nreturn sphere(x);\n",
			want:        "// Work Order: WO-1\n// Material: Aluminum\n@scene(60, 15)\n\nx = 1;\nreturn sphere(x);\n",
			wantChanged: true,
		},
		{
			name:        "places @scene before first non-comment token when comment appears after code",
			in:          "// Top doc comment\nx = 10;\n// Scene Size: 30\n// Scene Grid: 10\nreturn sphere(x);\n",
			want:        "// Top doc comment\n@scene(30)\n\nx = 10;\nreturn sphere(x);\n",
			wantChanged: true,
		},
		{
			name:        "idempotent when @scene already present and strips leftover legacy comments",
			in:          "@scene(20)\n// Scene Size: 20\n// Scene Grid: 10\n\nreturn sphere(1);\n",
			want:        "@scene(20)\n\nreturn sphere(1);\n",
			wantChanged: true,
		},
		{
			name:        "idempotent no-op when @scene already present without legacy comments",
			in:          "@scene(20, 5)\n\nreturn sphere(1);\n",
			want:        "@scene(20, 5)\n\nreturn sphere(1);\n",
			wantChanged: false,
		},
		{
			name:        "no-op when no scene comments or directive exist",
			in:          "// Just a regular GSL file\nx = 2;\nreturn sphere(x);\n",
			want:        "// Just a regular GSL file\nx = 2;\nreturn sphere(x);\n",
			wantChanged: false,
		},
		{
			name:        "preserves CRLF line endings",
			in:          "// Scene Size: 24\r\n// Scene Grid: 10\r\n\r\nreturn sphere(1);\r\n",
			want:        "@scene(24)\r\n\r\nreturn sphere(1);\r\n",
			wantChanged: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, changed, err := fix.Apply(tc.in, nil)
			require.NoError(t, err)
			assert.Equal(t, tc.wantChanged, changed)
			assert.Equal(t, tc.want, got)

			// Verify second run is strictly idempotent
			got2, changed2, err2 := fix.Apply(got, nil)
			require.NoError(t, err2)
			assert.False(t, changed2)
			assert.Equal(t, got, got2)
		})
	}
}
