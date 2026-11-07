package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/backend/audio"
	"gradientcontrol.com/omega/core/backend/mesh"
	"gradientcontrol.com/omega/core/backend/tape"
	"gradientcontrol.com/omega/core/backend/wgsl"
	"gradientcontrol.com/omega/core/cmd/compile/internal/config"
	"gradientcontrol.com/omega/core/parser"
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
	fmt.Fprintf(w, "  compile [flags] [source1.gsl source2.gsl ...]\n\n")
	fmt.Fprintf(w, "Flags:\n")
	fmt.Fprintf(w, "  -t, --target string\n")
	fmt.Fprintf(w, "        Compilation output target (wgsl, stl, tape, audio, ast) (default \"wgsl\")\n")
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
	fmt.Fprintf(w, "        Bounding box as minX,minY,minZ,maxX,maxY,maxZ\n")
	fmt.Fprintf(w, "      --algo string\n")
	fmt.Fprintf(w, "        Meshing algorithm (marching-cubes or mdc) (default \"marching-cubes\")\n")
	fmt.Fprintf(w, "      --time float\n")
	fmt.Fprintf(w, "        Animation time in seconds (default 0)\n")
	fmt.Fprintf(w, "  -O, --opt-level int\n")
	fmt.Fprintf(w, "        Optimization level (0 = debug, 1 = standard) (default 1)\n")
	fmt.Fprintf(w, "      --standalone\n")
	fmt.Fprintf(w, "        Append complete WebGPU raymarching runner harness to WGSL\n")
	fmt.Fprintf(w, "  -v, --verbose\n")
	fmt.Fprintf(w, "        Verbose diagnostic logging\n")
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
	if parseErr := parser.ParsePackage(sourceTexts).Error(); parseErr != nil {
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
			wgsl.WithStandalone(cfg.Standalone),
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
