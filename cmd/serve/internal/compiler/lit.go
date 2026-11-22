package compiler

import (
	"errors"
	"fmt"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/assembler/variation"
	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/compile"
	"gradientcontrol.com/omega/core/internal/meshimport"
)

// MaxLitPixels bounds a CPU frame: 4096 x 4096.
const MaxLitPixels = 4096 * 4096

// ErrLitRequest is wrapped by the errors RenderLit returns for a request it will not draw, as opposed to a document
// that does not compile or a render that fails.
var ErrLitRequest = errors.New("bad CPU frame request")

// LitRequest describes a server-side CPU frame: the viewer's own view of a document, drawn by the lit renderer
// (eval.RenderLitLinear) the way the viewer's standard frame is drawn, for the viewer's CPU draw device.
type LitRequest struct {
	ID       string `json:"id"`
	ClientID string `json:"clientId"`
	Source   string `json:"source"`
	// Width and Height are the frame's size in pixels, at most MaxLitPixels together.
	Width  int `json:"width"`
	Height int `json:"height"`
	// WorldExtent is the viewer's scene extent, as for the viewer's own frame.
	WorldExtent float64 `json:"worldExtent"`
	Time        float64 `json:"time"`
	// Camera is where the viewer's camera is, in world Z-up coordinates. It is required.
	Camera *LitCamera `json:"camera"`
	// Colors is each return's colour where the document authors none: the viewer's per-shape colours.
	Colors [][3]float64 `json:"colors,omitempty"`
	// Mode is the viewer tier drawn: "standard" (or empty), "color" (the standard frame without the document's
	// materials), or "simple" (the Schematic frame, whose constants Simple carries).
	Mode   string     `json:"mode,omitempty"`
	Simple *LitSimple `json:"simple,omitempty"`
	// BackDim is how bright a surface seen from behind is drawn: 0.25 for the Render panel's Back faces at dim, 1 at
	// full. Zero is dim.
	BackDim float64 `json:"backDim,omitempty"`
	// Meshes finds the STL files the source's importMesh declarations name, as on compile.Request.
	Meshes meshimport.Finder `json:"-"`
}

// LitCamera is the viewer's camera.
type LitCamera struct {
	Pos    [3]float64 `json:"pos"`
	Right  [3]float64 `json:"right"`
	Up     [3]float64 `json:"up"`
	Fwd    [3]float64 `json:"fwd"`
	TanFov [2]float64 `json:"tanFov"`
	Ortho  bool       `json:"ortho"`
}

// LitSimple is the Schematic frame's constants, as the viewer's gsl-simple.js defines them.
type LitSimple struct {
	Steps       int     `json:"steps"`
	Fudge       float64 `json:"fudge"`
	HitFloor    float64 `json:"hitFloor"`
	HitPixel    float64 `json:"hitPixel"`
	NormalFloor float64 `json:"normalFloor"`
	NormalPixel float64 `json:"normalPixel"`
	Ambient     float64 `json:"ambient"`
}

// RenderLit assembles the request's source, resolves the meshes it imports (from req.Meshes), and draws it from the
// request's camera. progressFn, if not nil, is told how far the render has got as stage "render", in thousandths.
func RenderLit(req LitRequest, sourcePaths []string, progressFn compile.ProgressFunc) (*eval.LitLinear, error) {
	opts, err := litOptions(req)
	if err != nil {
		return nil, err
	}
	if progressFn != nil {
		opts.Progress = func(share float64) {
			progressFn("render", "CPU frame", uint32(share*1000), 1000)
		}
	}

	asmRes, err := assembler.Assemble([]string{req.Source}, compile.AssemblerOptions(sourcePaths, nil)...)
	if err != nil {
		return nil, err
	}
	prog := asmRes.Program
	if len(asmRes.AudioReturns) > 0 {
		prog = variation.PrunePrimary(prog)
	}
	var uses []meshimport.Use
	if meshimport.References(prog) {
		if req.Meshes == nil {
			return nil, compile.ErrNoMeshFolder
		}
		if uses, err = meshimport.Resolve(prog, asmRes.Slots, req.Meshes); err != nil {
			return nil, err
		}
	}
	return withImportedMeshes(uses, func() (*eval.LitLinear, error) {
		return eval.RenderLitLinear(prog, opts)
	})
}

// litOptions checks the request and says what it asks of the lit renderer.
func litOptions(req LitRequest) (eval.LitOptions, error) {
	bad := func(format string, args ...any) (eval.LitOptions, error) {
		return eval.LitOptions{}, fmt.Errorf("%w: %s", ErrLitRequest, fmt.Sprintf(format, args...))
	}
	if req.Width <= 0 || req.Height <= 0 || req.Width > MaxLitPixels || req.Height > MaxLitPixels || req.Width*req.Height > MaxLitPixels {
		return bad("a %dx%d frame is out of range (at most %d pixels)", req.Width, req.Height, MaxLitPixels)
	}
	if req.Camera == nil {
		return bad("a frame needs the view's camera")
	}
	c := req.Camera
	opts := eval.LitOptions{
		Width:       req.Width,
		Height:      req.Height,
		WorldExtent: req.WorldExtent,
		Time:        req.Time,
		Camera: &eval.LitCamera{
			Pos: c.Pos, Right: c.Right, Up: c.Up, Fwd: c.Fwd, TanFov: c.TanFov, Ortho: c.Ortho,
		},
		Colors:      req.Colors,
		ViewerMarch: true,
		BackDim:     req.BackDim,
	}
	switch req.Mode {
	case "", "standard":
	case "color":
		opts.NoMaterials = true
	case "simple":
		s := req.Simple
		if s == nil {
			return bad("a Schematic frame needs its constants")
		}
		opts.Simple = &eval.LitSimple{
			Steps: s.Steps, Fudge: s.Fudge,
			HitFloor: s.HitFloor, HitPixel: s.HitPixel,
			NormalFloor: s.NormalFloor, NormalPixel: s.NormalPixel,
			Ambient: s.Ambient,
		}
	default:
		return bad("unknown mode %q", req.Mode)
	}
	return opts, nil
}
