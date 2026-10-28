package wgsl

import (
	api "gradientcontrol.com/omega/core/backend/wgsl/internal/lib"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/lib"
	"gradientcontrol.com/omega/core/program"
)

// RenderResult contains the emitted WGSL shader source code and geometry metadata.
type RenderResult = api.RenderResult

// ShapeLocations contains source spans that contribute to a shape.
type ShapeLocations = api.ShapeLocations

// Part represents a part bounding box in world coordinates.
type Part = api.Part

// Option configures WGSL shader rendering.
type Option func(*config)

type config struct {
	standalone bool
	progressFn ProgressFunc
}

// ProgressFunc reports compilation progress.
type ProgressFunc func(stage string, completed, total uint32)

// WithStandalone configures WGSL emission in standalone mode.
func WithStandalone(standalone bool) Option {
	return func(c *config) {
		c.standalone = standalone
	}
}

// WithProgress supplies a progress callback during WGSL emission.
func WithProgress(fn ProgressFunc) Option {
	return func(c *config) {
		c.progressFn = fn
	}
}

// Render emits WebGPU shader code from the given GSL source and assembled program AST.
func Render(source string, prog *program.Program, opts ...Option) (*RenderResult, error) {
	var cfg config
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.progressFn != nil {
		cfg.progressFn("Emitting WGSL shader", 1, 1)
	}
	return lib.Render(source, prog)
}
