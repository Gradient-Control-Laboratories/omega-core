package config_test

import (
	"testing"

	"gradientcontrol.com/omega/core/cmd/fix/internal/config"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestParseFlags_NoArgs(t *testing.T) {
	cfg, err := config.ParseFlags(nil)
	require.NoError(t, err)
	assert.Equal(t, config.ModeUsage, cfg.Mode)
}

func TestParseFlags_Help(t *testing.T) {
	for _, args := range [][]string{{"help"}, {"-h"}, {"-help"}, {"--help"}} {
		cfg, err := config.ParseFlags(args)
		require.NoError(t, err)
		assert.Equal(t, config.ModeHelp, cfg.Mode)
	}
}

func TestParseFlags_FixerHelp(t *testing.T) {
	cfg, err := config.ParseFlags([]string{"help", "scene"})
	require.NoError(t, err)
	assert.Equal(t, config.ModeFixerHelp, cfg.Mode)
	assert.Equal(t, "scene", cfg.HelpTopic)

	_, err = config.ParseFlags([]string{"help", "unknown_fixer"})
	require.Error(t, err)
}

func TestParseFlags_TargetsAndDiff(t *testing.T) {
	cfg, err := config.ParseFlags([]string{"-diff", "a.gsl", "dir"})
	require.NoError(t, err)
	assert.Equal(t, config.ModeRun, cfg.Mode)
	assert.True(t, cfg.Diff)
	assert.Equal(t, []string{"a.gsl", "dir"}, cfg.Paths)
	assert.Equal(t, 1, len(cfg.ActiveFixers))
	assert.Equal(t, "scene", cfg.ActiveFixers[0].Name)
}

func TestParseFlags_TriStateFixerSelection(t *testing.T) {
	// Explicit -scene=true enables scene
	cfg, err := config.ParseFlags([]string{"-scene", "a.gsl"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(cfg.ActiveFixers))
	assert.Equal(t, "scene", cfg.ActiveFixers[0].Name)

	// Explicit -scene=false disables scene
	cfgDisabled, err := config.ParseFlags([]string{"-scene=false", "a.gsl"})
	require.NoError(t, err)
	assert.Equal(t, 0, len(cfgDisabled.ActiveFixers))
}

func TestParseFlags_FlagsWithoutTargets(t *testing.T) {
	cfg, err := config.ParseFlags([]string{"-diff"})
	require.NoError(t, err)
	assert.Equal(t, config.ModeUsage, cfg.Mode)
}
