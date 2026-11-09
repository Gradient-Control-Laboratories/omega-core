package config

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

// Target specifies the compilation output artifact format.
type Target string

const (
	TargetWGSL  Target = "wgsl"
	TargetSTL   Target = "stl"
	TargetTape  Target = "tape"
	TargetAudio Target = "audio"
	TargetAST   Target = "ast"
)

// multiStringFlag collects multiple instances of a flag into a string slice.
type multiStringFlag []string

func (m *multiStringFlag) String() string {
	return strings.Join(*m, ", ")
}

func (m *multiStringFlag) Set(val string) error {
	*m = append(*m, val)
	return nil
}

// Config encapsulates parsed CLI flags and configuration for compile.
type Config struct {
	Target      Target
	OutputPath  string
	MetaPath    string
	NoMeta      bool
	SourcePaths []string
	Subdivision int
	Bbox        [6]float32
	HasBbox     bool
	// AutoBbox: --bbox auto, the box the field fills, worked out from the program (a mesh target's).
	AutoBbox   bool
	Algo       string
	Time       float64
	OptLevel   int
	Standalone bool
	Verbose    bool
	Help       bool
	Version    bool
	Sources    []string
	IsStdin    bool
}

// defaultExt returns the file extension corresponding to the configured target.
func (c *Config) defaultExt() string {
	switch c.Target {
	case TargetWGSL:
		return ".wgsl"
	case TargetSTL:
		return ".stl"
	case TargetTape:
		return ".tape"
	case TargetAudio:
		return ".js"
	case TargetAST:
		return ".ast.json"
	default:
		return ".out"
	}
}

// ResolveOutputPath determines the final destination path for the compiled artifact.
func (c *Config) ResolveOutputPath() string {
	if c.OutputPath == "-" {
		return "-"
	}
	if c.OutputPath != "" {
		return c.OutputPath
	}
	if c.IsStdin {
		return "-"
	}
	ext := c.defaultExt()
	base := strings.TrimSuffix(c.Sources[0], filepath.Ext(c.Sources[0]))
	return base + ext
}

// ResolveMetaPath determines the companion JSON metadata destination path.
func (c *Config) ResolveMetaPath() string {
	if c.NoMeta {
		return ""
	}
	if c.MetaPath != "" {
		return c.MetaPath
	}
	out := c.ResolveOutputPath()
	if out == "-" {
		return ""
	}
	base := strings.TrimSuffix(out, filepath.Ext(out))
	if strings.HasSuffix(out, ".ast.json") {
		base = strings.TrimSuffix(out, ".ast.json")
	}
	return base + ".json"
}

// ParseFlags parses CLI arguments into a validated Config.
func ParseFlags(args []string) (*Config, error) {
	fs := flag.NewFlagSet("compile", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var (
		targetStr   string
		outputStr   string
		metaStr     string
		noMeta      bool
		sourcePaths multiStringFlag
		subdivision int
		bboxStr     string
		algo        string
		timeSec     float64
		optLevel    int
		standalone  bool
		verbose     bool
		help        bool
		showVersion bool
	)

	// Target flag (-t, --target)
	fs.StringVar(&targetStr, "t", "wgsl", "Compilation output target (wgsl, stl, tape, audio, ast)")
	fs.StringVar(&targetStr, "target", "wgsl", "Compilation output target (wgsl, stl, tape, audio, ast)")

	// Output flag (-o, --output)
	fs.StringVar(&outputStr, "o", "", "Output destination path ('-' for stdout)")
	fs.StringVar(&outputStr, "output", "", "Output destination path ('-' for stdout)")

	// Meta flag (-m, --meta, --no-meta)
	fs.StringVar(&metaStr, "m", "", "Companion JSON metadata destination path")
	fs.StringVar(&metaStr, "meta", "", "Companion JSON metadata destination path")
	fs.BoolVar(&noMeta, "no-meta", false, "Suppress emission of companion metadata JSON")

	// Source paths (-I, --source-path)
	fs.Var(&sourcePaths, "I", "Root directory for package resolution (may be specified multiple times)")
	fs.Var(&sourcePaths, "source-path", "Root directory for package resolution (may be specified multiple times)")

	// Meshing flags
	fs.IntVar(&subdivision, "subdivision", 4, "Octree subdivision depth for mesh polygonization")
	fs.StringVar(&bboxStr, "bbox", "", "Bounding box as minX,minY,minZ,maxX,maxY,maxZ, or auto: the box the field fills")
	fs.StringVar(&algo, "algo", "marching-cubes", "Meshing algorithm (marching-cubes or mdc)")
	fs.Float64Var(&timeSec, "time", 0.0, "Animation time in seconds")

	// Optimization & diagnostics
	fs.IntVar(&optLevel, "O", 1, "Optimization level (0 = debug, 1 = standard)")
	fs.IntVar(&optLevel, "opt-level", 1, "Optimization level (0 = debug, 1 = standard)")
	fs.BoolVar(&standalone, "standalone", false, "Append complete WebGPU raymarching runner harness to WGSL")
	fs.BoolVar(&verbose, "v", false, "Verbose diagnostic logging")
	fs.BoolVar(&verbose, "verbose", false, "Verbose diagnostic logging")

	// Help
	fs.BoolVar(&help, "h", false, "Display usage instructions")
	fs.BoolVar(&help, "help", false, "Display usage instructions")
	fs.BoolVar(&showVersion, "version", false, "Print the version and exit")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	target := Target(strings.ToLower(strings.TrimSpace(targetStr)))
	switch target {
	case TargetWGSL, TargetSTL, TargetTape, TargetAudio, TargetAST:
	default:
		return nil, fmt.Errorf("unsupported target %q; must be one of wgsl, stl, tape, audio, ast", targetStr)
	}

	if subdivision < 1 {
		return nil, fmt.Errorf("subdivision depth must be >= 1, got %d", subdivision)
	}

	var (
		bbox     [6]float32
		hasBbox  bool
		autoBbox bool
	)
	if strings.EqualFold(strings.TrimSpace(bboxStr), "auto") {
		autoBbox = true
	} else if bboxStr != "" {
		parts := strings.Split(bboxStr, ",")
		if len(parts) != 6 {
			return nil, fmt.Errorf("invalid --bbox %q; expected 6 comma-separated floats (minX,minY,minZ,maxX,maxY,maxZ)", bboxStr)
		}
		for i, part := range parts {
			val, err := strconv.ParseFloat(strings.TrimSpace(part), 32)
			if err != nil {
				return nil, fmt.Errorf("invalid float in --bbox %q: %w", part, err)
			}
			bbox[i] = float32(val)
		}
		if bbox[0] > bbox[3] || bbox[1] > bbox[4] || bbox[2] > bbox[5] {
			return nil, fmt.Errorf("invalid --bbox; min bounds must not exceed max bounds: %v", bbox)
		}
		hasBbox = true
	}

	sources := fs.Args()
	isStdin := len(sources) == 0

	return &Config{
		Target:      target,
		OutputPath:  outputStr,
		MetaPath:    metaStr,
		NoMeta:      noMeta,
		SourcePaths: sourcePaths,
		Subdivision: subdivision,
		Bbox:        bbox,
		HasBbox:     hasBbox,
		AutoBbox:    autoBbox,
		Algo:        algo,
		Time:        timeSec,
		OptLevel:    optLevel,
		Standalone:  standalone,
		Verbose:     verbose,
		Help:        help,
		Version:     showVersion,
		Sources:     sources,
		IsStdin:     isStdin,
	}, nil
}
