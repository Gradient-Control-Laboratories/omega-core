package config_test

import (
	"testing"

	"gradientcontrol.com/omega/core/cmd/gslc/internal/config"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestParseFlags_Defaults(t *testing.T) {
	cfg, err := config.ParseFlags([]string{"foo.gsl"})
	require.NoError(t, err)

	assert.Equal(t, config.TargetWGSL, cfg.Target)
	assert.Equal(t, []string{"foo.gsl"}, cfg.Sources)
	assert.Equal(t, false, cfg.IsStdin)
	assert.Equal(t, "foo.wgsl", cfg.ResolveOutputPath())
	assert.Equal(t, "foo.json", cfg.ResolveMetaPath())
	assert.Equal(t, false, cfg.NoMeta)
	assert.Equal(t, 4, cfg.Subdivision)
	assert.Equal(t, "marching-cubes", cfg.Algo)
	assert.Equal(t, 0.0, cfg.Time)
	assert.Equal(t, 1, cfg.OptLevel)
	assert.Equal(t, false, cfg.Standalone)
	assert.Equal(t, false, cfg.Verbose)
	assert.Equal(t, false, cfg.Help)
}

func TestParseFlags_Stdin(t *testing.T) {
	cfg, err := config.ParseFlags([]string{})
	require.NoError(t, err)

	assert.Equal(t, true, cfg.IsStdin)
	assert.Empty(t, cfg.Sources)
	assert.Equal(t, "-", cfg.ResolveOutputPath())
	// Auto meta emission is suppressed when streaming to stdout
	assert.Equal(t, "", cfg.ResolveMetaPath())
}

func TestParseFlags_Targets(t *testing.T) {
	tests := []struct {
		target   string
		expected config.Target
		ext      string
	}{
		{"wgsl", config.TargetWGSL, "model.wgsl"},
		{"stl", config.TargetSTL, "model.stl"},
		{"tape", config.TargetTape, "model.tape"},
		{"audio", config.TargetAudio, "model.js"},
		{"ast", config.TargetAST, "model.ast.json"},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			cfg, err := config.ParseFlags([]string{"-t", tt.target, "model.gsl"})
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg.Target)
			assert.Equal(t, tt.ext, cfg.ResolveOutputPath())
		})
	}
}

func TestParseFlags_InvalidTarget(t *testing.T) {
	_, err := config.ParseFlags([]string{"-t", "invalid", "foo.gsl"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported target")
}

func TestParseFlags_ExplicitOutput(t *testing.T) {
	cfg, err := config.ParseFlags([]string{"-o", "out/custom.shader", "model.gsl"})
	require.NoError(t, err)
	assert.Equal(t, "out/custom.shader", cfg.ResolveOutputPath())
	assert.Equal(t, "out/custom.json", cfg.ResolveMetaPath())
}

func TestParseFlags_StdoutOutput(t *testing.T) {
	cfg, err := config.ParseFlags([]string{"-o", "-", "model.gsl"})
	require.NoError(t, err)
	assert.Equal(t, "-", cfg.ResolveOutputPath())
	assert.Equal(t, "", cfg.ResolveMetaPath())
}

func TestParseFlags_ExplicitMeta(t *testing.T) {
	cfg, err := config.ParseFlags([]string{"-o", "-", "-m", "custom_meta.json", "model.gsl"})
	require.NoError(t, err)
	assert.Equal(t, "-", cfg.ResolveOutputPath())
	assert.Equal(t, "custom_meta.json", cfg.ResolveMetaPath())
}

func TestParseFlags_NoMeta(t *testing.T) {
	cfg, err := config.ParseFlags([]string{"--no-meta", "model.gsl"})
	require.NoError(t, err)
	assert.Equal(t, true, cfg.NoMeta)
	assert.Equal(t, "", cfg.ResolveMetaPath())
}

func TestParseFlags_SourcePaths(t *testing.T) {
	cfg, err := config.ParseFlags([]string{"-I", "/lib/pkg1", "-I", "/lib/pkg2", "model.gsl"})
	require.NoError(t, err)
	assert.Equal(t, []string{"/lib/pkg1", "/lib/pkg2"}, cfg.SourcePaths)
}

func TestParseFlags_MeshingOptions(t *testing.T) {
	cfg, err := config.ParseFlags([]string{
		"--subdivision", "6",
		"--bbox", "-10,-20,-30,10,20,30",
		"--algo", "mdc",
		"--time", "2.5",
		"model.gsl",
	})
	require.NoError(t, err)
	assert.Equal(t, 6, cfg.Subdivision)
	assert.Equal(t, [6]float32{-10, -20, -30, 10, 20, 30}, cfg.Bbox)
	assert.Equal(t, true, cfg.HasBbox)
	assert.Equal(t, "mdc", cfg.Algo)
	assert.Equal(t, 2.5, cfg.Time)
}

func TestParseFlags_InvalidBbox(t *testing.T) {
	// Not enough components
	_, err := config.ParseFlags([]string{"--bbox", "-10,10", "model.gsl"})
	require.Error(t, err)

	// Min > Max
	_, err = config.ParseFlags([]string{"--bbox", "10,0,0,0,0,0", "model.gsl"})
	require.Error(t, err)
}

func TestParseFlags_FlagsAndOptions(t *testing.T) {
	cfg, err := config.ParseFlags([]string{
		"-O", "0",
		"--standalone",
		"-v",
		"a.gsl", "b.gsl",
	})
	require.NoError(t, err)
	assert.Equal(t, 0, cfg.OptLevel)
	assert.Equal(t, true, cfg.Standalone)
	assert.Equal(t, true, cfg.Verbose)
	assert.Equal(t, []string{"a.gsl", "b.gsl"}, cfg.Sources)
}

func TestParseFlags_Help(t *testing.T) {
	cfg, err := config.ParseFlags([]string{"--help"})
	require.NoError(t, err)
	assert.Equal(t, true, cfg.Help)
}
