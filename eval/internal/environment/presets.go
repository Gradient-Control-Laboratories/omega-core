// Package environment holds the viewers' environment light presets. The one
// copy is presets.json beside this file: common/web/gsl-environment.js imports
// it for the viewers, and this package embeds it for litrender, so the GPU and
// CPU renderers light a scene from the same numbers.
package environment

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed presets.json
var presetsJSON []byte

// Default is the preset the viewers open with: their own lights, as before.
const Default = "Viewer"

// MaxPanels is how many panels a preset may have: the shaders carry four.
const MaxPanels = 4

// Panel is one soft light, a spherical Gaussian: an axis (Z up, laid out for
// the viewer's starting view from the (+x, +y) side; it need not be unit), a
// sharpness lambda (the lobe's angular radius is about 1/sqrt(lambda)
// radians) and a linear RGB amplitude.
type Panel struct {
	Axis      [3]float64 `json:"axis"`
	Sharpness float64    `json:"sharpness"`
	Color     [3]float64 `json:"color"`
}

// Preset is one environment. Mode 0 is the viewer's own lighting, shaded by
// its original expressions; mode 1 is analytic, a step sky (Sky above the
// horizon, Ground below, linear radiance) plus the panels. Rig is how strongly
// the viewer's three rig lights still shine.
type Preset struct {
	Description string     `json:"description"`
	Mode        int        `json:"mode"`
	Rig         float64    `json:"rig"`
	Sky         [3]float64 `json:"sky"`
	Ground      [3]float64 `json:"ground"`
	Panels      []Panel    `json:"panels"`
}

var (
	loadOnce sync.Once
	presets  map[string]Preset
	loadErr  error
)

// Presets returns every preset by name.
func Presets() (map[string]Preset, error) {
	loadOnce.Do(func() {
		if err := json.Unmarshal(presetsJSON, &presets); err != nil {
			loadErr = fmt.Errorf("environment presets: %w", err)
			return
		}
		for name, p := range presets {
			if p.Mode != 0 && p.Mode != 1 {
				loadErr = fmt.Errorf("environment preset %q: mode %d", name, p.Mode)
				return
			}
			if len(p.Panels) > MaxPanels {
				loadErr = fmt.Errorf("environment preset %q: %d panels, at most %d", name, len(p.Panels), MaxPanels)
				return
			}
		}
		if _, ok := presets[Default]; !ok {
			loadErr = fmt.Errorf("environment presets: no %q", Default)
		}
	})
	return presets, loadErr
}

// Lookup returns the preset called name.
func Lookup(name string) (Preset, error) {
	all, err := Presets()
	if err != nil {
		return Preset{}, err
	}
	p, ok := all[name]
	if !ok {
		return Preset{}, fmt.Errorf("no environment preset %q", name)
	}
	return p, nil
}
