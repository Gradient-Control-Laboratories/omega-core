package config_test

import (
	"testing"

	"gradientcontrol.com/omega/core/cmd/compile/internal/config"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestParseFlags_BboxAuto(t *testing.T) {
	for _, given := range []string{"auto", "AUTO", " auto "} {
		cfg, err := config.ParseFlags([]string{"-t", "stl", "--bbox", given, "foo.gsl"})
		require.NoError(t, err)
		assert.Equal(t, true, cfg.AutoBbox)
		assert.Equal(t, false, cfg.HasBbox)
	}
	cfg, err := config.ParseFlags([]string{"foo.gsl"})
	require.NoError(t, err)
	assert.Equal(t, false, cfg.AutoBbox)
}

func TestParseFlags_Version(t *testing.T) {
	cfg, err := config.ParseFlags([]string{"--version"})
	require.NoError(t, err)
	assert.Equal(t, true, cfg.Version)
}
