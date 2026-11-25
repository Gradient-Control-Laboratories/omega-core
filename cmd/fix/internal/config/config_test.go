package config_test

import (
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/cmd/fix/internal/config"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
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
	for _, topic := range []string{"scene", "omegaimport"} {
		cfg, err := config.ParseFlags([]string{"help", topic})
		require.NoError(t, err)
		assert.Equal(t, config.ModeFixerHelp, cfg.Mode)
		assert.Equal(t, topic, cfg.HelpTopic)
	}

	_, err := config.ParseFlags([]string{"help", "unknown_fixer"})
	require.Error(t, err)
}

func TestParseFlags_TargetsAndDiff(t *testing.T) {
	cfg, err := config.ParseFlags([]string{"-range=false", "-stdlib=false", "-diff", "a.gsl", "dir"})
	require.NoError(t, err)
	assert.Equal(t, config.ModeRun, cfg.Mode)
	assert.True(t, cfg.Diff)
	assert.Equal(t, []string{"a.gsl", "dir"}, cfg.Paths)
	require.Equal(t, 2, len(cfg.ActiveFixers))
	assert.Equal(t, "scene", cfg.ActiveFixers[0].Name)
	assert.Equal(t, "omegaimport", cfg.ActiveFixers[1].Name)
}

func TestParseFlags_TriStateFixerSelection(t *testing.T) {
	// Explicit -scene=true enables only scene
	cfg, err := config.ParseFlags([]string{"-scene", "a.gsl"})
	require.NoError(t, err)
	require.Equal(t, 1, len(cfg.ActiveFixers))
	assert.Equal(t, "scene", cfg.ActiveFixers[0].Name)

	// Explicit -omegaimport=true enables only omegaimport
	cfgOmega, err := config.ParseFlags([]string{"-omegaimport", "a.gsl"})
	require.NoError(t, err)
	require.Equal(t, 1, len(cfgOmega.ActiveFixers))
	assert.Equal(t, "omegaimport", cfgOmega.ActiveFixers[0].Name)

	// Explicit -scene=false leaves remaining fixers enabled
	cfgSceneOff, err := config.ParseFlags([]string{"-scene=false", "a.gsl"})
	require.NoError(t, err)
	require.Equal(t, 3, len(cfgSceneOff.ActiveFixers))
	assert.Equal(t, "range", cfgSceneOff.ActiveFixers[0].Name)
	assert.Equal(t, "stdlib", cfgSceneOff.ActiveFixers[1].Name)
	assert.Equal(t, "omegaimport", cfgSceneOff.ActiveFixers[2].Name)

	// Disabling all fixers leaves ActiveFixers empty
	cfgAllOff, err := config.ParseFlags([]string{"-scene=false", "-range=false", "-stdlib=false", "-omegaimport=false", "a.gsl"})
	require.NoError(t, err)
	assert.Equal(t, 0, len(cfgAllOff.ActiveFixers))
}

func TestParseFlags_FlagsWithoutTargets(t *testing.T) {
	cfg, err := config.ParseFlags([]string{"-diff"})
	require.NoError(t, err)
	assert.Equal(t, config.ModeUsage, cfg.Mode)
}
