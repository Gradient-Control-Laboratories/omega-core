package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/backend/audio"
	"gradientcontrol.com/omega/core/backend/mesh"
	"gradientcontrol.com/omega/core/backend/tape"
	"gradientcontrol.com/omega/core/backend/wgsl"
	"gradientcontrol.com/omega/core/cmd/compile/internal/config"
	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/version"
	"gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
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
	fmt.Fprintf(w, "        Compilation output target (wgsl, stl, tape, audio, ast, png) (default \"wgsl\")\n")
	fmt.Fprintf(w, "  -o, --output string\n")
	fmt.Fprintf(w, "        Output destination path ('-' for stdout)\n")
	fmt.Fprintf(w, "  -m, --meta string\n")
	fmt.Fprintf(w, "        Companion JSON metadata destination path\n")
	fmt.Fprintf(w, "      --no-meta\n")
	fmt.Fprintf(w, "        Suppress emission of companion metadata JSON\n")
	fmt.Fprintf(w, "  -I, --source-path string\n")
	fmt.Fprintf(w, "        Root directory for package resolution (may be specified multiple times)\n")
	fmt.Fprintf(w, "      --subdivision int\n")
	fmt.Fprintf(w, "        Octree subdivision depth for mesh polygonization (default 4)\n")
	fmt.Fprintf(w, "      --bbox string\n")
	fmt.Fprintf(w, "        Bounding box as minX,minY,minZ,maxX,maxY,maxZ, or auto: the box the field fills,\n")
	fmt.Fprintf(w, "        found within 1024 mm of the origin and grown by two cells (default -10..10 on each axis)\n")
	fmt.Fprintf(w, "      --algo string\n")
	fmt.Fprintf(w, "        Meshing algorithm (marching-cubes or mdc) (default \"marching-cubes\")\n")
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

	// Backend target emission
	var outBytes []byte

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

	case config.TargetSTL:
		meshOpts := []mesh.Option{
			mesh.WithSubdivision(cfg.Subdivision),
			mesh.WithAlgorithm(cfg.Algo),
			mesh.WithTime(cfg.Time),
		}
		if cfg.HasBbox {
			meshOpts = append(meshOpts, mesh.WithBbox(cfg.Bbox))
		}
		if cfg.AutoBbox {
			if len(sourceTexts) > 1 {
				return fmt.Errorf("--bbox auto takes one source file, not %d: name the box outright", len(sourceTexts))
			}
			box, boxErr := autoBbox(primarySourceText, srcPaths)
			if boxErr != nil {
				return boxErr
			}
			if cfg.Verbose {
				fmt.Fprintf(stderr, "[mesh] --bbox auto: %g,%g,%g,%g,%g,%g\n", box[0], box[1], box[2], box[3], box[4], box[5])
			}
			meshOpts = append(meshOpts, mesh.WithBbox(box))
		}
		if cfg.Verbose {
			meshOpts = append(meshOpts, mesh.WithProgress(func(stage string, completed, total uint32) {
				fmt.Fprintf(stderr, "[mesh] %s (%d/%d)\n", stage, completed, total)
			}))
		}
		stlBytes, meshErr := mesh.Mesh(res.Program, meshOpts...)
		if meshErr != nil {
			return fmt.Errorf("generating STL mesh: %w", meshErr)
		}
		outBytes = stlBytes

	case config.TargetTape:
		tapeOpts := []tape.Option{}
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
		jsCode, _, audioErr := audio.AudioToJs(res.Program)
		if audioErr != nil {
			return fmt.Errorf("compiling Web Audio JavaScript: %w", audioErr)
		}
		outBytes = []byte(jsCode)

	case config.TargetPNG:
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

// autoBbox is --bbox auto: the box the program's field fills -- eval's Bounds, a grid fit of where the union of
// its returns is negative, searched within 1024 mm of the origin -- grown by two of the fit's cells on every side,
// so the surface lies inside the meshed box with room for the mesher's cells around it.
func autoBbox(code string, srcPaths []string) ([6]float32, error) {
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
	found, err := eval.New(nil, nil, opts...).Bounds(context.Background(), nil, code, "", expr, [6]float64{}, 64)
	if err != nil {
		return box, fmt.Errorf("--bbox auto: %w", err)
	}
	if !found.Found {
		return box, fmt.Errorf("--bbox auto: the field is nowhere inside within 1024 mm of the origin (its least value %g, at %v)",
			found.MinValue, found.MinAt)
	}
	for k := 0; k < 3; k++ {
		pad := 2 * found.CellSize[k]
		box[k] = float32(found.BboxMin[k] - pad)
		box[k+3] = float32(found.BboxMax[k] + pad)
	}
	return box, nil
}
