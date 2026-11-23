package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/assembler/variation"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/audio"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/mesh"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/tape"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/wgsl"
	"github.com/Gradient-Control-Laboratories/omega-core/cmd/compile/internal/config"
	"github.com/Gradient-Control-Laboratories/omega-core/eval"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/compile"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/meshimport"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/version"
	"github.com/Gradient-Control-Laboratories/omega-core/parser"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// companionMeta defines the structure of the JSON metadata file emitted alongside compiled targets.
type companionMeta struct {
	Target      string `json:"target"`
	Source      string `json:"source"`
	Sliders     any    `json:"sliders"`
	UIControls  any    `json:"uiControls"`
	Scene       any    `json:"scene,omitempty"`
	Diagnostics any    `json:"diagnostics,omitempty"`
}

// PrintUsage outputs formatted CLI help information.
func PrintUsage(w io.Writer) {
	fmt.Fprintf(w, "Usage of compile:\n")
	fmt.Fprintf(w, "  compile [flags] [source1.gsl source2.gsl ...]\n")
	fmt.Fprintf(w, "  compile eval [flags] REQUEST.json | --stdio   (query a document's field: compile eval --help)\n\n")
	fmt.Fprintf(w, "Flags:\n")
	fmt.Fprintf(w, "  -t, --target string\n")
	fmt.Fprintf(w, "        Compilation output target (wgsl, stl, 3mf, glb, tape, audio, ast, png, field) (default \"wgsl\")\n")
	fmt.Fprintf(w, "  -o, --output string\n")
	fmt.Fprintf(w, "        Output destination path ('-' for stdout)\n")
	fmt.Fprintf(w, "  -m, --meta string\n")
	fmt.Fprintf(w, "        Companion JSON metadata destination path\n")
	fmt.Fprintf(w, "      --no-meta\n")
	fmt.Fprintf(w, "        Suppress emission of companion metadata JSON\n")
	fmt.Fprintf(w, "  -I, --source-path string\n")
	fmt.Fprintf(w, "        Root directory for package resolution (may be specified multiple times)\n")
	fmt.Fprintf(w, "      --subdivision int\n")
	fmt.Fprintf(w, "        Octree subdivision depth for mesh polygonization (default 9)\n")
	fmt.Fprintf(w, "      --bbox string\n")
	fmt.Fprintf(w, "        Bounding box as minX,minY,minZ,maxX,maxY,maxZ, or auto: the box the field fills,\n")
	fmt.Fprintf(w, "        found within 1024 mm of the origin and grown by two cells (default: the returned value's .box,\n")
	fmt.Fprintf(w, "        declared or inferred, when it is bounded; else the @scene box, an edge of\n")
	fmt.Fprintf(w, "        its size centred on the origin, as the viewer frames it; else auto)\n")
	fmt.Fprintf(w, "      --algo string\n")
	fmt.Fprintf(w, "        Meshing algorithm (marching-cubes or mdc) (default \"mdc\")\n")
	fmt.Fprintf(w, "      --time float\n")
	fmt.Fprintf(w, "        Animation time in seconds (default 0)\n")
	fmt.Fprintf(w, "      --views string\n")
	fmt.Fprintf(w, "        Cameras for png: presets (%s) or x,y,z directions,\n", strings.Join(eval.ViewPresetNames(), ", "))
	fmt.Fprintf(w, "        comma-separated; several tile two across into one image (default \"iso\")\n")
	fmt.Fprintf(w, "      --size string\n")
	fmt.Fprintf(w, "        One png panel's size as WxH pixels (default \"640x480\")\n")
	fmt.Fprintf(w, "      --extent float\n")
	fmt.Fprintf(w, "        The png scene's half-size in mm; 0 frames the geometry (default 0)\n")
	fmt.Fprintf(w, "      --light\n")
	fmt.Fprintf(w, "        A light studio ground for png instead of the dark canvas\n")
	fmt.Fprintf(w, "  -O, --opt-level int\n")
	fmt.Fprintf(w, "        Optimization level (0 = debug, 1 = standard) (default 1)\n")
	fmt.Fprintf(w, "      --standalone\n")
	fmt.Fprintf(w, "        Append complete WebGPU raymarching runner harness to WGSL\n")
	fmt.Fprintf(w, "  -v, --verbose\n")
	fmt.Fprintf(w, "        Verbose diagnostic logging\n")
	fmt.Fprintf(w, "      --version\n")
	fmt.Fprintf(w, "        Print the version and exit\n")
	fmt.Fprintf(w, "  -h, --help\n")
	fmt.Fprintf(w, "        Display usage instructions\n")
}

// Run executes the compile pipeline with the given arguments and I/O streams.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	cfg, err := config.ParseFlags(args)
	if err != nil {
		if err == flag.ErrHelp {
			PrintUsage(stdout)
			return nil
		}
		return err
	}

	if cfg.Help {
		PrintUsage(stdout)
		return nil
	}
	if cfg.Version {
		fmt.Fprintf(stdout, "compile %s\n", version.Version)
		return nil
	}

	var (
		sourceTexts       []string
		sourceName        string
		primarySourceText string
	)

	if cfg.IsStdin {
		if stdin == nil {
			stdin = os.Stdin
		}
		data, readErr := io.ReadAll(stdin)
		if readErr != nil {
			return fmt.Errorf("reading stdin: %w", readErr)
		}
		if len(data) == 0 {
			return fmt.Errorf("stdin is empty; provide source files or pipe code via stdin")
		}

		sourceTexts = []string{string(data)}
		sourceName = "stdin"
		primarySourceText = string(data)
	} else {
		sourceTexts = make([]string, len(cfg.Sources))
		for i, s := range cfg.Sources {
			data, readErr := os.ReadFile(s)
			if readErr != nil {
				return fmt.Errorf("reading source file %q: %w", s, readErr)
			}
			sourceTexts[i] = string(data)
		}
		sourceName = cfg.Sources[0]
		primarySourceText = sourceTexts[0]
	}

	// The field target is a viewer's bundle, compiled as gsl serve's /api/compile compiles it
	// (internal/compile). A document with errors still gets one, listing them.
	if cfg.Target == config.TargetField {
		return runField(cfg, sourceTexts, sourceName, primarySourceText, stdout, stderr)
	}

	// Validate parser on primary package
	if _, parseErr := parser.ParsePackage(sourceTexts); parseErr != nil {
		return fmt.Errorf("parsing primary package: %w", parseErr)
	}

	// Configure assembler options
	asmOpts := []assembler.Option{
		assembler.ExtractUI(),
	}
	srcPaths := cfg.SourcePaths
	if len(srcPaths) == 0 && len(cfg.Sources) > 0 {
		dir := filepath.Dir(cfg.Sources[0])
		if dir == "" {
			dir = "."
		}
		srcPaths = []string{dir}
	}
	if len(srcPaths) > 0 {
		asmOpts = append(asmOpts, assembler.SourcePaths(srcPaths...))
	}
	if cfg.OptLevel == 0 {
		asmOpts = append(asmOpts, assembler.DisableDeadcodeRemoval())
	}
	if cfg.Verbose {
		asmOpts = append(asmOpts, assembler.WithProgress(func(stage string, completed, total uint32) {
			fmt.Fprintf(stderr, "[assemble] %s (%d/%d)\n", stage, completed, total)
		}))
	}

	res, asmErr := assembler.Assemble(sourceTexts, asmOpts...)
	if asmErr != nil {
		return fmt.Errorf("assembling program: %w", asmErr)
	}

	// The meshes bvhSdf and vdbSdf read: importMesh files beside the source (or in the working directory, for stdin).
	docDir := "."
	if len(cfg.Sources) > 0 {
		docDir = filepath.Dir(cfg.Sources[0])
	}
	// Only the targets that evaluate the field load them; wgsl leaves binding the textures to its host.
	meshDirs := meshDirsFor(docDir, srcPaths)
	loadMeshes := func() ([]meshimport.Use, error) {
		return meshimport.Resolve(res.Program, res.Slots, meshimport.Dirs(meshDirs...))
	}

	// Backend target emission
	var outBytes []byte
	var zipOutput bool

	switch cfg.Target {
	case config.TargetWGSL:
		wgslOpts := []wgsl.Option{
			wgsl.Standalone(cfg.Standalone),
		}
		if cfg.Verbose {
			wgslOpts = append(wgslOpts, wgsl.WithProgress(func(stage string, completed, total uint32) {
				fmt.Fprintf(stderr, "[wgsl] %s (%d/%d)\n", stage, completed, total)
			}))
		}
		renderRes, renderErr := wgsl.Render(primarySourceText, res.Program, wgslOpts...)
		if renderErr != nil {
			return fmt.Errorf("rendering WGSL shader: %w", renderErr)
		}
		outBytes = []byte(renderRes.WGSL)

	case config.TargetSTL, config.Target3MF, config.TargetGLB:
		meshOpts := []mesh.Option{
			mesh.WithSubdivision(cfg.Subdivision),
			mesh.WithAlgorithm(cfg.Algo),
			mesh.WithTime(cfg.Time),
		}
		declared, hasDeclared := [6]float32{}, false
		if !cfg.HasBbox && !cfg.AutoBbox && len(sourceTexts) == 1 {
			box, ok, boxErr := returnBox(primarySourceText, srcPaths, meshDirs)
			if boxErr != nil {
				return boxErr
			}
			declared, hasDeclared = box, ok
		}
		switch {
		case cfg.HasBbox:
			meshOpts = append(meshOpts, mesh.WithBbox(cfg.Bbox))
		case hasDeclared:
			// No box named: the returned value's .box, which says where the part is, padded by the mesher by one of
			// its own cells so a face lying on the box is meshed.
			if cfg.Verbose {
				fmt.Fprintf(stderr, "[mesh] returned .box: %g,%g,%g,%g,%g,%g\n", declared[0], declared[1], declared[2], declared[3], declared[4], declared[5])
			}
			meshOpts = append(meshOpts, mesh.WithBbox(declared), mesh.WithBboxPadCells(1))
		case cfg.AutoBbox || res.Scene.GetSize() <= 0:
			// --bbox auto, or no box named and no @scene to take one from: the box the field fills.
			if len(sourceTexts) > 1 {
				if cfg.AutoBbox {
					return fmt.Errorf("--bbox auto takes one source file, not %d: name the box outright", len(sourceTexts))
				}
				return fmt.Errorf("%d source files and no @scene to take a box from: name the box with --bbox", len(sourceTexts))
			}
			box, boxErr := autoBbox(primarySourceText, srcPaths, meshDirs)
			if boxErr != nil {
				return boxErr
			}
			if cfg.Verbose {
				fmt.Fprintf(stderr, "[mesh] --bbox auto: %g,%g,%g,%g,%g,%g\n", box[0], box[1], box[2], box[3], box[4], box[5])
			}
			meshOpts = append(meshOpts, mesh.WithBbox(box))
		default:
			// No box named: the @scene box, an edge of its size centred on the origin, as the viewer frames it.
			half := float32(res.Scene.GetSize() / 2)
			box := [6]float32{-half, -half, -half, half, half, half}
			if cfg.Verbose {
				fmt.Fprintf(stderr, "[mesh] @scene box: %g,%g,%g,%g,%g,%g\n", box[0], box[1], box[2], box[3], box[4], box[5])
			}
			meshOpts = append(meshOpts, mesh.WithBbox(box))
		}
		if cfg.Verbose {
			meshOpts = append(meshOpts, mesh.WithProgress(func(stage string, completed, total uint32) {
				fmt.Fprintf(stderr, "[mesh] %s (%d/%d)\n", stage, completed, total)
			}))
		}
		if regErr := registerMeshes(loadMeshes); regErr != nil {
			return regErr
		}
		// Each top-level return is its own body: one STL, a zip of <name>.stl for several, or a 3MF object or GLB node each.
		bodies, meshErr := mesh.MeshBodies(res.Program, meshOpts...)
		if meshErr != nil {
			return fmt.Errorf("generating STL mesh: %w", meshErr)
		}
		if cfg.Target == config.Target3MF {
			if len(bodies) == 1 {
				bodies[0].Name = bodyName(cfg)
			}
			var buf bytes.Buffer
			if threeMfErr := mesh.Write3MF(&buf, bodies, nil); threeMfErr != nil {
				return fmt.Errorf("writing 3MF: %w", threeMfErr)
			}
			outBytes = buf.Bytes()
			break
		}
		if cfg.Target == config.TargetGLB {
			if len(bodies) == 1 {
				bodies[0].Name = bodyName(cfg)
			}
			var buf bytes.Buffer
			if glbErr := mesh.WriteGLB(&buf, bodies, nil); glbErr != nil {
				return fmt.Errorf("writing GLB: %w", glbErr)
			}
			outBytes = buf.Bytes()
			break
		}
		packed, zipped, packErr := mesh.PackSTL(bodies)
		if packErr != nil {
			return fmt.Errorf("zipping bodies: %w", packErr)
		}
		if zipped && strings.EqualFold(filepath.Ext(cfg.OutputPath), ".stl") {
			return fmt.Errorf("%d bodies make a zip of STLs, not one STL: name the output .zip (-o %s.zip)",
				len(bodies), strings.TrimSuffix(cfg.OutputPath, filepath.Ext(cfg.OutputPath)))
		}
		zipOutput = zipped
		outBytes = packed

	case config.TargetTape:
		meshUses, loadErr := loadMeshes()
		if loadErr != nil {
			return loadErr
		}
		bvhOracles, vdbOracles := meshimport.Oracles(meshUses)
		tapeOpts := []tape.Option{tape.MeshOracles(bvhOracles), tape.VdbOracles(vdbOracles)}
		// The box a mesh of this document would cover, stamped in the tape for whoever reads it: --bbox, else the
		// returned value's .box, else the @scene box. With none of them it stays unset.
		switch {
		case cfg.HasBbox:
			tapeOpts = append(tapeOpts, tape.Bbox(cfg.Bbox))
		case len(sourceTexts) == 1:
			box, ok, boxErr := returnBox(primarySourceText, srcPaths, meshDirs)
			if boxErr != nil {
				return boxErr
			}
			if ok {
				tapeOpts = append(tapeOpts, tape.Bbox(box))
			} else if half := float32(res.Scene.GetSize() / 2); half > 0 {
				tapeOpts = append(tapeOpts, tape.Bbox([6]float32{-half, -half, -half, half, half, half}))
			}
		}
		if cfg.Verbose {
			tapeOpts = append(tapeOpts, tape.WithProgress(func(stage string, completed, total uint32) {
				fmt.Fprintf(stderr, "[tape] %s (%d/%d)\n", stage, completed, total)
			}))
		}
		tapeBytes, tapeErr := tape.Compile(res.Program, tapeOpts...)
		if tapeErr != nil {
			return fmt.Errorf("compiling bytecode tape: %w", tapeErr)
		}
		outBytes = tapeBytes

	case config.TargetAudio:
		// Compile the audio returns alone, as the viewer's compile does
		// (internal/compile): the whole program also holds the geometry
		// and library packages the audio backend does not translate.
		audioProg := res.Program
		if len(res.AudioReturns) > 0 {
			audioOps := make([]*program.Operand, len(res.AudioReturns))
			for i, ar := range res.AudioReturns {
				audioOps[i] = ar.Return
			}
			audioProg = variation.Program(res.Program, audioOps...)
		}
		jsCode, _, audioErr := audio.AudioToJs(audioProg)
		if audioErr != nil {
			return fmt.Errorf("compiling Web Audio JavaScript: %w", audioErr)
		}
		outBytes = []byte(jsCode)

	case config.TargetPNG:
		if regErr := registerMeshes(loadMeshes); regErr != nil {
			return regErr
		}
		pngBytes, pngErr := renderPNG(res.Program, cfg)
		if pngErr != nil {
			return pngErr
		}
		outBytes = pngBytes

	case config.TargetAST:
		astJson, jsonErr := json.MarshalIndent(res.Program, "", "  ")
		if jsonErr != nil {
			return fmt.Errorf("marshaling AST to JSON: %w", jsonErr)
		}
		outBytes = astJson

	default:
		return fmt.Errorf("unsupported target format: %s", cfg.Target)
	}

	// Output routing
	outputPath := cfg.ResolveOutputPath()
	if zipOutput && cfg.OutputPath == "" && outputPath != "-" {
		// The default name follows the source; several bodies make a zip, not an STL.
		outputPath = strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".zip"
	}
	if outputPath == "-" {
		if stdout == nil {
			stdout = os.Stdout
		}
		if _, writeErr := stdout.Write(outBytes); writeErr != nil {
			return fmt.Errorf("writing output to stdout: %w", writeErr)
		}
	} else {
		if dirErr := os.MkdirAll(filepath.Dir(outputPath), 0755); dirErr != nil {
			return fmt.Errorf("creating output directory: %w", dirErr)
		}
		if writeErr := os.WriteFile(outputPath, outBytes, 0644); writeErr != nil {
			return fmt.Errorf("writing output file %q: %w", outputPath, writeErr)
		}
	}

	// Companion metadata emission
	metaPath := cfg.ResolveMetaPath()
	if metaPath != "" {
		meta := companionMeta{
			Target:      string(cfg.Target),
			Source:      sourceName,
			Sliders:     res.Sliders,
			UIControls:  res.UIControls,
			Scene:       res.Scene,
			Diagnostics: res.Diagnostics,
		}
		metaBytes, metaJsonErr := json.MarshalIndent(meta, "", "  ")
		if metaJsonErr != nil {
			return fmt.Errorf("marshaling companion metadata: %w", metaJsonErr)
		}
		if dirErr := os.MkdirAll(filepath.Dir(metaPath), 0755); dirErr != nil {
			return fmt.Errorf("creating companion metadata directory: %w", dirErr)
		}
		if writeErr := os.WriteFile(metaPath, metaBytes, 0644); writeErr != nil {
			return fmt.Errorf("writing companion metadata file %q: %w", metaPath, writeErr)
		}
	}

	return nil
}

// runField writes source's field bundle (internal/compile) as JSON: the WGSL with its shapes, parts,
// sliders, UI controls, @scene and imported meshes' preview grids, the same as gsl serve's /api/compile answers
// for the same document. A document with errors gets a bundle too, holding them with their lines and columns, and
// then runField returns an error, so the exit status says the bundle draws nothing. There is no companion
// metadata: the bundle holds it.
func runField(cfg *config.Config, sourceTexts []string, sourceName, source string, stdout, stderr io.Writer) error {
	if len(sourceTexts) > 1 {
		return fmt.Errorf("-t field takes one source file, not %d", len(sourceTexts))
	}
	if cfg.MetaPath != "" {
		return fmt.Errorf("-t field writes no companion metadata (the bundle holds the sliders, UI controls and scene): drop -m")
	}
	srcPaths := cfg.SourcePaths
	docDir := "."
	if len(cfg.Sources) > 0 {
		docDir = filepath.Dir(cfg.Sources[0])
		if len(srcPaths) == 0 {
			srcPaths = []string{docDir}
		}
	}
	var progress compile.ProgressFunc
	if cfg.Verbose {
		progress = func(stage, label string, completed, total uint32) {
			fmt.Fprintf(stderr, "[%s] %s (%d/%d)\n", stage, label, completed, total)
		}
	}
	// The bundle's id is the document's file name, never its path: a host serves the bundle to a page, which is
	// told no more of the file system than the name it shows.
	bundle := compile.Process(compile.Request{
		ID:     filepath.Base(sourceName),
		Source: source,
		Meshes: meshimport.Dirs(meshDirsFor(docDir, srcPaths)...),
	}, srcPaths, progress)
	data, err := json.Marshal(bundle)
	if err != nil {
		return fmt.Errorf("marshaling the field bundle: %w", err)
	}
	data = append(data, '\n')

	outputPath := cfg.ResolveOutputPath()
	if outputPath == "-" {
		if stdout == nil {
			stdout = os.Stdout
		}
		if _, err := stdout.Write(data); err != nil {
			return fmt.Errorf("writing output to stdout: %w", err)
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			return fmt.Errorf("writing output file %q: %w", outputPath, err)
		}
	}

	if n := len(bundle.Errors); n > 0 {
		first := bundle.Errors[0]
		return fmt.Errorf("%s: %d error(s), listed in the field bundle; the first, %d:%d: %s",
			sourceName, n, first.LineNum, first.Column, first.Message)
	}
	return nil
}

// meshDirsFor is where a document's importMesh files are looked for: its own folder first, then each source path (a
// dependency package's files sit in its folder under one).
func meshDirsFor(docDir string, srcPaths []string) []string {
	dirs := []string{docDir}
	for _, p := range srcPaths {
		if filepath.Clean(p) != filepath.Clean(docDir) {
			dirs = append(dirs, p)
		}
	}
	return dirs
}

// registerMeshes loads the document's meshes and registers them as the oracles its bvhSdf and vdbSdf calls read.
func registerMeshes(load func() ([]meshimport.Use, error)) error {
	uses, err := load()
	if err != nil {
		return err
	}
	return meshimport.Register(uses)
}

// renderPNG is the png target: the program's field raymarched on the CPU by the viewer's lit renderer, from each
// of cfg.Views, tiled two across into one image. It draws the field itself, never a mesh of it.
func renderPNG(prog *program.Program, cfg *config.Config) ([]byte, error) {
	views, err := eval.ParseViews(cfg.Views)
	if err != nil {
		return nil, fmt.Errorf("--views: %w", err)
	}
	if len(views) == 0 {
		return nil, fmt.Errorf("--views names no camera")
	}
	panels := make([][]byte, 0, len(views))
	for i, v := range views {
		v.Light = cfg.Light
		panel, renderErr := eval.RenderLitSizedView(prog, cfg.Width, cfg.Height, cfg.Extent, cfg.Time, v)
		if renderErr != nil {
			return nil, fmt.Errorf("rendering view %d: %w", i+1, renderErr)
		}
		if len(panel) == 0 {
			return nil, fmt.Errorf("rendering view %d: the renderer returned no image", i+1)
		}
		panels = append(panels, panel)
	}
	if len(panels) == 1 {
		return panels[0], nil
	}
	return eval.ComposeSheet(panels, 2)
}

// autoReach is how far from the origin --bbox auto looks for the field: eval's Bounds' default search region.
const autoReach = 1024.0

// autoRescans bounds how many times --bbox auto widens its search around the box it has found.
const autoRescans = 8

// returnBox is the box the document says its returned value fills (eval's ReturnBox: a box declared on the name it
// returns, as written, or the box the analysis carries), handed to the mesher as it is; ok is false when that box is
// unbounded on some axis, and the caller falls back to @scene or --bbox auto.
func returnBox(code string, srcPaths, meshDirs []string) ([6]float32, bool, error) {
	var box [6]float32
	returns := eval.TopLevelReturnExprs(code)
	if len(returns) == 0 {
		return box, false, nil
	}
	var opts []eval.Option
	if len(srcPaths) > 0 {
		opts = append(opts, assembler.SourcePaths(srcPaths...))
	}
	ev := eval.New(nil, nil, opts...)
	ev.LoadMeshesFrom(meshDirs...)
	found, ok, err := ev.ReturnBox(context.Background(), nil, code, "", returns)
	if err != nil {
		return box, false, fmt.Errorf("reading the returned value's .box: %w", err)
	}
	if !ok {
		return box, false, nil
	}
	for k := range box {
		box[k] = float32(found[k])
	}
	return box, true, nil
}

// autoBbox is --bbox auto: the box the program's field fills, searched within autoReach of the origin, grown by two
// cells on every side so the surface lies inside the meshed box with room for the mesher's cells around it.
//
// eval's Bounds, a grid fit of where the union of its returns is negative, gives the first box. Its first pass samples
// cells 32 mm across, and its second looks again only within a cell of what the first caught: a part thinner than
// those cells can be caught by a few samples and the rest of it missed, the box stopping at the edge of where the
// second pass looked. So the search is widened around the box -- Bounds again, over the box grown on every side by
// its own size, where the grid's cells are a few mm -- and the box grows by what it finds, until it no longer reaches
// the edge of the region searched. It is a fit by sampling, not a proof: a field's interval bounds through its
// gradient-scaled blends (unionRoundExact and the like) are unbounded, so a box cannot be proven for them.
func autoBbox(code string, srcPaths, meshDirs []string) ([6]float32, error) {
	var box [6]float32
	returns := eval.TopLevelReturnExprs(code)
	if len(returns) == 0 {
		return box, fmt.Errorf("--bbox auto: the program returns nothing to mesh")
	}
	expr := returns[0]
	if len(returns) > 1 {
		expr = "min(" + strings.Join(returns, ", ") + ")"
	}
	var opts []eval.Option
	if len(srcPaths) > 0 {
		opts = append(opts, assembler.SourcePaths(srcPaths...))
	}
	ev := eval.New(nil, nil, opts...)
	ev.LoadMeshesFrom(meshDirs...)
	ctx := context.Background()
	found, err := ev.Bounds(ctx, nil, code, "", expr, [6]float64{}, 64)
	if err != nil {
		return box, fmt.Errorf("--bbox auto: %w", err)
	}
	if !found.Found {
		return box, fmt.Errorf("--bbox auto: the field is nowhere inside within 1024 mm of the origin (its least value %g, at %v)",
			found.MinValue, found.MinAt)
	}
	fit := [6]float64{found.BboxMin[0], found.BboxMin[1], found.BboxMin[2],
		found.BboxMax[0], found.BboxMax[1], found.BboxMax[2]}
	cell := found.CellSize
	for range autoRescans {
		var region [6]float64
		for k := 0; k < 3; k++ {
			grow := math.Max(fit[k+3]-fit[k], 4*cell[k])
			region[k] = math.Max(fit[k]-grow, -autoReach)
			region[k+3] = math.Min(fit[k+3]+grow, autoReach)
		}
		wider, err := ev.Bounds(ctx, nil, code, "", expr, region, 64)
		if err != nil {
			return box, fmt.Errorf("--bbox auto: %w", err)
		}
		if !wider.Found {
			break
		}
		cell = wider.CellSize
		reaches := false
		for k := 0; k < 3; k++ {
			fit[k] = math.Min(fit[k], wider.BboxMin[k])
			fit[k+3] = math.Max(fit[k+3], wider.BboxMax[k])
			if (region[k] > -autoReach && fit[k] < region[k]+cell[k]) ||
				(region[k+3] < autoReach && fit[k+3] > region[k+3]-cell[k]) {
				reaches = true
			}
		}
		if !reaches {
			break
		}
	}
	for k := 0; k < 3; k++ {
		pad := 2 * cell[k]
		box[k] = float32(fit[k] - pad)
		box[k+3] = float32(fit[k+3] + pad)
	}
	return box, nil
}

// bodyName names a one-body 3MF's object or GLB's node: the first source file's name without
// its extension, as the default output path takes it, or "model" from stdin.
// Several bodies keep their own names (mesh.MeshBodies).
func bodyName(cfg *config.Config) string {
	if cfg.IsStdin || len(cfg.Sources) == 0 {
		return "model"
	}
	base := filepath.Base(cfg.Sources[0])
	return strings.TrimSuffix(base, filepath.Ext(base))
}
