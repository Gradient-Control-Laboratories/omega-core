package lib

import (
	api "gradientcontrol.com/omega/core/backend/internal/fidget"
	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/mesh"
	"gradientcontrol.com/omega/core/program"
)

// Mesh converts a program strictly to an STL using the new convert logic.
// This is exposed as a library function for monolithic clients and serves as a thin
// wrapper around the internal service.GenerateMesh function.
func Mesh(
	prog *program.Program,
	subdivisionLevel int, bbox [6]float32,
	algorithm string,
	progressFn func(completed, total uint32, status string),
) ([]byte, error) {
	return MeshAtTime(prog, subdivisionLevel, bbox, algorithm, 0, progressFn)
}

// MeshAtTime is Mesh with the animation clock set: every `time` reference in
// the program evaluates to timeAt, so the mesh matches that frame. Mesh keeps
// the 0 a document opens at.
func MeshAtTime(
	prog *program.Program,
	subdivisionLevel int, bbox [6]float32,
	algorithm string,
	timeAt float64,
	progressFn func(completed, total uint32, status string),
) ([]byte, error) {
	return mesh.GenerateMesh(
		prog,
		subdivisionLevel, bbox,
		algorithm,
		timeAt,
		progressFn)
}

// EvalGrad evaluates each scalar return of the module at the given points,
// returning per-return samples (value + spatial gradient). Vector-valued
// returns flatten into consecutive scalar results, matching Mesh's treatment
// of returns.
func EvalGrad(prog *program.Program, xs, ys, zs []float32) ([][]api.GradSample, error) {
	internal, err := mesh.EvalGrad(prog, xs, ys, zs)
	if err != nil {
		return nil, err
	}
	results := make([][]api.GradSample, len(internal))
	for i, samples := range internal {
		results[i] = convertGradSamples(samples)
	}
	return results, nil
}

// EvalGradGroups evaluates every return of the module at the given points,
// accepting vector- and matrix-valued returns: each return becomes one
// GradGroup (GSL type name + per-component samples, [component][point]).
func EvalGradGroups(prog *program.Program, xs, ys, zs []float32) ([]api.GradGroup, error) {
	internal, err := mesh.EvalGradGroups(prog, xs, ys, zs)
	if err != nil {
		return nil, err
	}
	results := make([]api.GradGroup, len(internal))
	for i, group := range internal {
		converted := api.GradGroup{TypeName: group.TypeName}
		for _, samples := range group.Samples {
			converted.Samples = append(converted.Samples, convertGradSamples(samples))
		}
		results[i] = converted
	}
	return results, nil
}

// EvalIntervalParamBoxes interval-bounds the module's single scalar return
// over each box with per-box parameter ranges (names orders the parameters;
// bounds[i] is box i's [lo, hi] per name). See mesh.EvalIntervalParamBoxes.
func EvalIntervalParamBoxes(prog *program.Program, names []string, boxes [][6]float32, bounds [][][2]float64) ([]api.IntervalSample, error) {
	internal, err := mesh.EvalIntervalParamBoxes(prog, names, boxes, bounds)
	if err != nil {
		return nil, err
	}
	out := make([]api.IntervalSample, len(internal))
	for i, s := range internal {
		out[i] = api.IntervalSample{Lower: s.Lower, Upper: s.Upper}
	}
	return out, nil
}

// EvalIntervalParams is EvalInterval with named document parameters ranging
// over intervals — the bracket is sound over each box and the whole parameter
// ranges at once. The module must have a single scalar return; each named
// parameter must be a slider-style literal declaration.
func EvalIntervalParams(prog *program.Program, boxes [][6]float32, params map[string][2]float64) ([]api.IntervalSample, error) {
	internal, err := mesh.EvalIntervalParams(prog, boxes, params)
	if err != nil {
		return nil, err
	}
	out := make([]api.IntervalSample, len(internal))
	for i, s := range internal {
		out[i] = api.IntervalSample{Lower: s.Lower, Upper: s.Upper}
	}
	return out, nil
}

// EvalInterval interval-bounds each scalar return of the module over every
// axis-aligned box, returning a sound [lower, upper] bracket per box
// (oracle-aware; no Lipschitz constant). A query module has one return, so
// results[0] holds the brackets. Each box is [xMin,yMin,zMin,xMax,yMax,zMax].
func EvalInterval(prog *program.Program, boxes [][6]float32) ([][]api.IntervalSample, error) {
	internal, err := mesh.EvalInterval(prog, boxes)
	if err != nil {
		return nil, err
	}
	results := make([][]api.IntervalSample, len(internal))
	for i, samples := range internal {
		converted := make([]api.IntervalSample, len(samples))
		for j, s := range samples {
			converted[j] = api.IntervalSample{Lower: s.Lower, Upper: s.Upper}
		}
		results[i] = converted
	}
	return results, nil
}

func convertGradSamples(samples []fidget.GradResult) []api.GradSample {
	converted := make([]api.GradSample, len(samples))
	for j, s := range samples {
		converted[j] = api.GradSample{Value: s.Value, Dx: s.Dx, Dy: s.Dy, Dz: s.Dz}
	}
	return converted
}

// Render generates a png and glsl representation of the model.
// This is a thin wrapper over the internal render generator for structural integration.
func Render(prog *program.Program) ([]byte, *api.RenderGlslResponse, error) {
	pngBytes, internalResult, err := mesh.GenerateRender(prog)
	return pngBytes, convertRenderResult(internalResult), err
}

// RenderFramed is Render with the camera framed on the geometry instead of
// the fixed default view: the union field's occupied box centers the view
// and sets its scale, falling back to the fixed camera when the frame scan
// finds nothing. For offline consumers (e.g. the eval benchmark's saved
// previews) where no interactive camera exists.
func RenderFramed(prog *program.Program) ([]byte, *api.RenderGlslResponse, error) {
	pngBytes, internalResult, err := mesh.GenerateRenderFramed(prog)
	return pngBytes, convertRenderResult(internalResult), err
}

// RenderLit renders the module the way the webdm viewer shades it — the same
// three-light rig, ambient occlusion, speculars, and the document's own
// authored color fields — by sphere-tracing the field on the CPU.
func RenderLit(prog *program.Program, opts mesh.LitOptions) ([]byte, error) {
	return mesh.RenderLit(prog, opts)
}

// RenderLitSized is RenderLit for callers outside this module: it takes the
// options as plain scalars, because mesh.LitOptions lives in an internal
// package and so cannot be named from another module. Zero width/height take
// the renderer's defaults; zero worldExtent frames from the document's
// "Scene Size" header or, failing that, from the geometry.
func RenderLitSized(
	prog *program.Program,
	width, height int,
	worldExtent, timeAt float64,
) ([]byte, error) {
	return RenderLit(prog, mesh.LitOptions{
		Width:       width,
		Height:      height,
		WorldExtent: worldExtent,
		Time:        timeAt,
	})
}

// LitView is the camera and ground of a lit render, for callers outside this
// module. The zero value is the viewer's own framing: the corner-on camera,
// Z up, the dark canvas, the world-axis overlay.
type LitView struct {
	// Dir is the direction from the look-at point toward the camera; zero
	// keeps the corner-on (1, 1, 1) view.
	Dir [3]float64
	// Up is the world direction that reads as up in the image; zero is +Z.
	Up [3]float64
	// Light swaps the dark canvas for a near-white studio ground.
	Light bool
	// NoAxes omits the world-axis overlay.
	NoAxes bool
	// Environment names an environment light preset, the viewers' Light row;
	// empty is the viewer's own lighting.
	Environment string
	// EnvironmentTurn turns the environment about Z by this many degrees
	// beyond the camera's own turn, as the viewers' Turn slider does.
	EnvironmentTurn float64
	// GlassSamples is the refracted rays per pixel a rough transmitting
	// surface takes; zero is the viewer's own count.
	GlassSamples int
}

// RenderLitSizedView is RenderLitSized with a chosen camera and ground.
func RenderLitSizedView(
	prog *program.Program,
	width, height int,
	worldExtent, timeAt float64,
	view LitView,
) ([]byte, error) {
	opts := mesh.LitOptions{
		Width:       width,
		Height:      height,
		WorldExtent: worldExtent,
		Time:        timeAt,
		View:        view.Dir,
		Up:          view.Up,
		Light:       view.Light,
		NoAxes:      view.NoAxes,
		Environment: view.Environment,
	}
	opts.EnvironmentTurn = view.EnvironmentTurn
	opts.GlassSamples = view.GlassSamples
	return RenderLit(prog, opts)
}

// RenderLitSizedViewRadiance is RenderLitSizedView returning, beside the PNG,
// each opaque pixel's radiance before tone mapping — linear light, three per
// pixel, row-major from the top left, zero where nothing opaque was hit. A
// caller recolouring the render works there, where shading is linear in the
// albedo (see mesh.LitLinear.Radiance).
func RenderLitSizedViewRadiance(
	prog *program.Program,
	width, height int,
	worldExtent, timeAt float64,
	view LitView,
) ([]byte, []float32, error) {
	opts := mesh.LitOptions{
		Width:       width,
		Height:      height,
		WorldExtent: worldExtent,
		Time:        timeAt,
		View:        view.Dir,
		Up:          view.Up,
		Light:       view.Light,
		NoAxes:      view.NoAxes,
		Environment: view.Environment,
	}
	opts.EnvironmentTurn = view.EnvironmentTurn
	opts.GlassSamples = view.GlassSamples
	png, lin, err := mesh.RenderLitWithLinear(prog, opts)
	if err != nil {
		return nil, nil, err
	}
	return png, lin.Radiance, nil
}

func convertRenderResult(internalResult *mesh.GlslResult) *api.RenderGlslResponse {
	if internalResult == nil {
		return nil
	}
	inputs := make(map[string]*api.RenderInput, len(internalResult.Inputs))
	for k, v := range internalResult.Inputs {
		inputs[k] = &api.RenderInput{
			Value:  v.Value,
			HasMin: v.HasMin,
			HasMax: v.HasMax,
			Min:    v.Min,
			Max:    v.Max,
		}
	}
	return &api.RenderGlslResponse{
		Inputs: inputs,
		Glsl:   internalResult.GlslSrc,
	}
}
