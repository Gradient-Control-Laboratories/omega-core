package compiler

import (
	"errors"
	"strings"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/assembler/slider"
	"gradientcontrol.com/omega/core/assembler/ui"
	"gradientcontrol.com/omega/core/backend/audio"
	"gradientcontrol.com/omega/core/backend/mesh"
	"gradientcontrol.com/omega/core/backend/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// ProgressFunc reports compilation or meshing progress milestones.
type ProgressFunc func(stage, label string, completed, total uint32)

// CompileRequest describes an interactive GSL compilation request.
type CompileRequest struct {
	ID       string `json:"id"`
	ClientID string `json:"clientId"`
	Source   string `json:"source"`
}

// AudioResult contains compiled Web Audio synthesis JavaScript and duration.
type AudioResult struct {
	JSCode   string `json:"jsCode"`
	Duration int    `json:"duration"`
}

// CompileResponse contains the compiled WGSL shader, UI metadata, optional audio, or structured diagnostics.
type CompileResponse struct {
	ID             string                `json:"id"`
	WGSL           string                `json:"wgsl,omitempty"`
	ShapeCount     int                   `json:"shapeCount,omitempty"`
	ShapeNames     []string              `json:"shapeNames,omitempty"`
	ShapeLocations []wgsl.ShapeLocations `json:"shapeLocations,omitempty"`
	MeshSdfSlots   []wgsl.MeshSdfSlot    `json:"meshSdfSlots,omitempty"`
	Parts          []wgsl.Part           `json:"parts,omitempty"`
	Sliders        []slider.SliderInfo   `json:"sliders,omitempty"`
	UIControls     []ui.Control          `json:"uiControls,omitempty"`
	Scene          *program.Scene        `json:"scene,omitempty"`
	MeshDecls      any                   `json:"meshDecls,omitempty"`
	Audio          *AudioResult          `json:"audio,omitempty"`
	Errors         []program.Error       `json:"errors,omitempty"`
}

// MeshRequest describes a server-side binary STL meshing request.
type MeshRequest struct {
	ID          string     `json:"id"`
	ClientID    string     `json:"clientId"`
	Source      string     `json:"source"`
	Subdivision int        `json:"subdivision"`
	Bbox        [6]float32 `json:"bbox"`
	HasBbox     bool       `json:"-"`
	Algo        string     `json:"algo"`
	Time        float64    `json:"time"`
}

// Compile executes the in-process GSL compilation pipeline (audio extraction, AST assembly, WGSL codegen, and audio JS codegen).
func Compile(req CompileRequest, sourcePaths []string, progressFn ProgressFunc) *CompileResponse {
	resp := &CompileResponse{ID: req.ID}

	if strings.TrimSpace(req.Source) == "" {
		resp.Errors = []program.Error{
			{
				LineNum: 1,
				Column:  1,
				Length:  1,
				Message: "source is empty",
			},
		}
		return resp
	}

	audioSrc, shapeSrc, err := audio.ExtractAudio(req.Source)
	if err != nil {
		resp.Errors = UnwrapProgramErrors(req.Source, err)
		return resp
	}
	if audioSrc == "" && shapeSrc == "" {
		shapeSrc = req.Source
	}

	asmOpts := buildAssemblerOptions(sourcePaths, progressFn)

	if strings.TrimSpace(shapeSrc) != "" {
		asmRes, err := assembler.Assemble([]string{shapeSrc}, asmOpts...)
		if err != nil {
			resp.Errors = UnwrapProgramErrors(shapeSrc, err)
			return resp
		}

		wgslOpts := []wgsl.Option{}
		if progressFn != nil {
			wgslOpts = append(wgslOpts, wgsl.WithProgress(func(stage string, completed, total uint32) {
				progressFn("wgsl-build", stage, completed, total)
			}))
		}

		renderRes, err := wgsl.Render(shapeSrc, asmRes.Program, wgslOpts...)
		if err != nil {
			resp.Errors = UnwrapProgramErrors(shapeSrc, err)
			return resp
		}

		resp.WGSL = renderRes.WGSL
		resp.ShapeCount = renderRes.ShapeCount
		resp.ShapeNames = renderRes.ShapeNames
		resp.ShapeLocations = renderRes.ShapeLocations
		resp.MeshSdfSlots = renderRes.MeshSdfSlots
		resp.Parts = renderRes.Parts
		resp.Sliders = asmRes.Sliders
		resp.UIControls = asmRes.UIControls
		resp.Scene = asmRes.Scene
	}

	if strings.TrimSpace(audioSrc) != "" {
		audioAsmRes, err := assembler.Assemble([]string{audioSrc}, buildAssemblerOptions(sourcePaths, nil)...)
		if err != nil {
			resp.Errors = UnwrapProgramErrors(audioSrc, err)
			return resp
		}
		jsCode, duration, err := audio.AudioToJs(audioAsmRes.Program)
		if err != nil {
			resp.Errors = UnwrapProgramErrors(audioSrc, err)
			return resp
		}
		resp.Audio = &AudioResult{
			JSCode:   jsCode,
			Duration: duration,
		}
	}

	return resp
}

// Mesh assembles the GSL source and evaluates a binary STL mesh using the Fidget/Manifold mesh backend.
func Mesh(req MeshRequest, sourcePaths []string, progressFn ProgressFunc) ([]byte, error) {
	_, shapeSrc, err := audio.ExtractAudio(req.Source)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(shapeSrc) == "" {
		shapeSrc = req.Source
	}

	asmRes, err := assembler.Assemble([]string{shapeSrc}, buildAssemblerOptions(sourcePaths, progressFn)...)
	if err != nil {
		return nil, err
	}

	subdiv := req.Subdivision
	if subdiv <= 0 {
		subdiv = 4
	}
	algo := req.Algo
	if algo == "" {
		algo = "marching-cubes"
	}

	meshOpts := []mesh.Option{
		mesh.WithSubdivision(subdiv),
		mesh.WithAlgorithm(algo),
		mesh.WithTime(req.Time),
	}
	if req.HasBbox {
		meshOpts = append(meshOpts, mesh.WithBbox(req.Bbox))
	}
	if progressFn != nil {
		meshOpts = append(meshOpts, mesh.WithProgress(func(stage string, completed, total uint32) {
			progressFn("mesh", stage, completed, total)
		}))
	}

	return mesh.Mesh(asmRes.Program, meshOpts...)
}

func buildAssemblerOptions(sourcePaths []string, progressFn ProgressFunc) []assembler.Option {
	opts := []assembler.Option{
		assembler.ExtractUI(),
	}
	if len(sourcePaths) > 0 {
		opts = append(opts, assembler.SourcePaths(sourcePaths...))
	}
	if progressFn != nil {
		opts = append(opts, assembler.WithProgress(func(stage string, completed, total uint32) {
			progressFn("ast-parse", stage, completed, total)
		}))
	}
	return opts
}

// UnwrapProgramErrors enriches and flattens an error tree into structured program.Error values.
func UnwrapProgramErrors(src string, err error) []program.Error {
	enriched := build.ErrWithSrc(src, err)
	var out []program.Error
	collectErrors(src, enriched, &out)
	if len(out) == 0 && err != nil {
		out = append(out, program.Error{
			LineNum: 1,
			Column:  1,
			Length:  1,
			Message: err.Error(),
		})
	}
	return out
}

// tokenWidth is how many characters of line, from the 1-based column, an error
// there should underline: the whole word or number when the column starts one
// (a dotted name like lib.fn counts as one), else the single character.
func tokenWidth(line string, column int) int {
	start := column - 1
	if start < 0 || start >= len(line) || !isWordByte(line[start]) {
		return 1
	}
	end := start
	for end < len(line) && (isWordByte(line[end]) || line[end] == '.') {
		end++
	}
	return end - start
}

func isWordByte(b byte) bool {
	return b == '_' || ('0' <= b && b <= '9') || ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z')
}

func collectErrors(src string, err error, out *[]program.Error) {
	if err == nil {
		return
	}
	if multi, ok := err.(interface{ Unwrap() []error }); ok {
		for _, sub := range multi.Unwrap() {
			collectErrors(src, sub, out)
		}
		return
	}
	var pErr *program.Error
	if errors.As(err, &pErr) && pErr != nil {
		cp := *pErr
		if cp.Line == "" && cp.LineNum > 0 {
			cp.Line = build.LineFromSource(src, cp.LineNum)
		}
		if cp.Length <= 1 {
			cp.Length = tokenWidth(cp.Line, cp.Column)
		}
		*out = append(*out, cp)
		return
	}
	*out = append(*out, program.Error{
		LineNum: 1,
		Column:  1,
		Length:  1,
		Message: err.Error(),
	})
}
