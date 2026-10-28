package mesh

import (
	lib "gradientcontrol.com/omega/core/backend/internal/fidget/lib"
	"gradientcontrol.com/omega/core/program"
)

// Option configures 3D mesh generation.
type Option func(*options)

type options struct {
	subdivision int
	bbox        [6]float32
	hasBbox     bool
	algo        string
	time        float64
	progressFn  ProgressFunc
}

// ProgressFunc is invoked during mesh generation milestones.
type ProgressFunc func(stage string, completed, total uint32)

// WithProgress supplies a progress callback to report meshing progress.
func WithProgress(fn ProgressFunc) Option {
	return func(o *options) {
		o.progressFn = fn
	}
}

// WithSubdivision sets the octree subdivision depth (default: 4).
func WithSubdivision(level int) Option {
	return func(o *options) {
		o.subdivision = level
	}
}

// WithBbox sets the explicit world-space bounding box for meshing.
// If omitted, the bounding box is automatically derived from the program's @scene or AST bounds.
func WithBbox(bbox [6]float32) Option {
	return func(o *options) {
		o.bbox = bbox
		o.hasBbox = true
	}
}

// WithAlgorithm selects the polygonization algorithm ("marching-cubes" or "mdc").
func WithAlgorithm(algo string) Option {
	return func(o *options) {
		o.algo = algo
	}
}

// WithTime sets the animation timestamp for evaluating dynamic geometry (default: 0.0).
func WithTime(timeSec float64) Option {
	return func(o *options) {
		o.time = timeSec
	}
}

// Body represents a distinct solid body in a multi-body export.
type Body struct {
	Name string
	STL  []byte
}

// Mesh converts an assembled GSL program into a single watertight binary STL mesh.
// It automatically coordinates the underlying geometry engines:
//   - Uses the Fidget dual-contouring mesher for analytical SDF expressions.
//   - Uses the Manifold boolean engine for exact mesh booleans and imported mesh assets.
func Mesh(prog *program.Program, opts ...Option) ([]byte, error) {
	o := options{
		subdivision: 4,
		algo:        "marching-cubes",
		time:        0.0,
		bbox:        [6]float32{-10, -10, -10, 10, 10, 10},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	var internalProgress func(completed, total uint32, status string)
	if o.progressFn != nil {
		internalProgress = func(completed, total uint32, status string) {
			o.progressFn(status, completed, total)
		}
	}

	return lib.MeshAtTime(prog, o.subdivision, o.bbox, o.algo, o.time, internalProgress)
}

// MeshBodies converts an assembled program into discrete named bodies.
func MeshBodies(prog *program.Program, opts ...Option) ([]Body, error) {
	stl, err := Mesh(prog, opts...)
	if err != nil {
		return nil, err
	}
	return []Body{
		{
			Name: "default",
			STL:  stl,
		},
	}, nil
}
