package viewer

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gradientcontrol.com/omega/core/cmd/serve/internal/staticpage"
	"gradientcontrol.com/omega/core/internal/compile"
	"gradientcontrol.com/omega/core/internal/meshimport"
)

// `compile export`, which `gsl export` runs: the viewer saved as a folder that needs no server. What the
// folder holds is staticpage's to say (cmd/serve/internal/staticpage), and the viewer's Export panel
// downloads the same thing as a zip; this file is the command: its flags, the documents it is given, and
// where it writes.

const exportUsage = `Usage of compile export (gsl export):
  gsl export [flags] part.gsl [more.gsl ...]

Writes a folder that shows the files in the WebGPU viewer with no server behind it: the viewer's page,
and each file's source with the field it compiles to. Sliders, gizmos, time, sections and the render
settings work as they do in gsl viewer. The source is shown and cannot be changed, since the page
compiles nothing. Several files share the one page and are picked from its menu, the first one first;
?file=part.gsl on the address opens another.

Browsers do not load the page from a file:// address: put the folder on any static web server.

Flags:
  -o, --output <dir>        the folder to write: a new or empty one, or an earlier export, whose models
                            are replaced (default: the first file's name, -html in place of .gsl)
  -I, --source-path <dir>   where imported packages are found, in place of each file's own folder
                            (may be given more than once)
      --viewer-dir <dir>    the viewer's web files (default: the ones installed with this program)
  -v, --verbose             log each file's compilation stages
  -h, --help                display this help

Exit codes: 0 written, 1 a file did not compile or the folder could not be written, 2 a bad argument.
`

// The export's layout, by the names this command and its tests use.
const (
	staticHost = staticpage.StaticHost
	servedHost = staticpage.ServedHost
)

var exportSkips = staticpage.Skips

type (
	exportedExample = staticpage.Example
	exportedIndex   = staticpage.Index
	exportedModel   = staticpage.Model
)

// exportOptions are the parsed command line.
type exportOptions struct {
	output      string
	viewerDir   string
	sourcePaths []string
	files       []string
	verbose     bool
	help        bool
}

type pathList []string

func (p *pathList) String() string { return strings.Join(*p, ",") }

func (p *pathList) Set(val string) error {
	*p = append(*p, val)
	return nil
}

// parseExport reads flags before, between and after the files.
func parseExport(args []string) (exportOptions, error) {
	var o exportOptions
	flags := flag.NewFlagSet("export", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&o.output, "output", "", "")
	flags.StringVar(&o.output, "o", "", "")
	flags.Var((*pathList)(&o.sourcePaths), "source-path", "")
	flags.Var((*pathList)(&o.sourcePaths), "I", "")
	flags.StringVar(&o.viewerDir, "viewer-dir", "", "")
	flags.BoolVar(&o.verbose, "verbose", false, "")
	flags.BoolVar(&o.verbose, "v", false, "")
	flags.BoolVar(&o.help, "help", false, "")
	flags.BoolVar(&o.help, "h", false, "")
	for {
		if err := flags.Parse(args); err != nil {
			return o, err
		}
		if flags.NArg() == 0 {
			break
		}
		o.files = append(o.files, flags.Arg(0))
		args = flags.Args()[1:]
	}
	return o, nil
}

// errDidNotCompile is a document's errors, already written to stderr.
var errDidNotCompile = errors.New("nothing was written")

// Export is `compile export ARGS`: it writes the folder and returns the exit code.
func Export(args []string, stdout, stderr io.Writer) int {
	o, err := parseExport(args)
	if err != nil {
		fmt.Fprintf(stderr, "gsl export: %v\n\n%s", err, exportUsage)
		return 2
	}
	if o.help {
		fmt.Fprint(stdout, exportUsage)
		return 0
	}
	if len(o.files) == 0 {
		fmt.Fprintf(stderr, "gsl export: name a .gsl file to export\n\n%s", exportUsage)
		return 2
	}
	viewerDir, err := exportViewerDir(o.viewerDir, executableDir())
	if err != nil {
		fmt.Fprintf(stderr, "gsl export: %v\n", err)
		return 2
	}
	docs, err := exportDocuments(o.files)
	if err != nil {
		fmt.Fprintf(stderr, "gsl export: %v\n", err)
		return 2
	}
	out := o.output
	if out == "" {
		out = strings.TrimSuffix(docs[0].name, ".gsl") + "-html"
	}

	models, err := compileDocuments(docs, o, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "gsl export: %v\n", err)
		return 1
	}
	size, err := writeExport(out, viewerDir, models)
	if err != nil {
		fmt.Fprintf(stderr, "gsl export: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "gsl export: %s (%d %s, %.1f MB)\n  viewer: %s\n"+
		"Put the folder on a static web server to open it; browsers do not load it from file://.\n",
		out, len(models), plural(len(models), "model", "models"), float64(size)/1e6, viewerDir)
	return 0
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// exportViewerDir is the viewer's folder, found as `gsl viewer` finds it; it must hold the static host.
func exportViewerDir(named, exeDir string) (string, error) {
	dir := named
	if dir == "" {
		dir = findViewer(exeDir)
		if dir == "" {
			return "", errors.New("found no viewer next to this program or above the current folder: " +
				"name it with --viewer-dir")
		}
	} else if !isViewer(dir) {
		return "", fmt.Errorf("%s has no index.html: it is not the viewer's folder", dir)
	}
	if err := staticpage.CheckViewer(dir); err != nil {
		return "", err
	}
	return dir, nil
}

// document is a file to export: where it is, and the name the page lists it under (its file name).
type document struct {
	path string
	name string
}

// exportDocuments checks the files named: each a .gsl file, no two of one name (the page lists them by it).
func exportDocuments(files []string) ([]document, error) {
	docs := make([]document, 0, len(files))
	seen := map[string]string{}
	for _, file := range files {
		path, err := filepath.Abs(file)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		name := filepath.Base(path)
		switch {
		case info.IsDir():
			return nil, fmt.Errorf("%s is a folder: name the .gsl files in it", file)
		case !strings.HasSuffix(name, ".gsl") || name == ".gsl":
			return nil, fmt.Errorf("%s is not a .gsl file", file)
		}
		if other, dup := seen[name]; dup {
			return nil, fmt.Errorf("%s and %s have the same name, and the page lists files by name", other, file)
		}
		seen[name] = file
		docs = append(docs, document{path: path, name: name})
	}
	return docs, nil
}

// compileDocuments compiles each document as the viewer's server does. A document with errors has them
// written to stderr, and then no folder is written.
func compileDocuments(docs []document, o exportOptions, stderr io.Writer) ([]exportedModel, error) {
	models := make([]exportedModel, 0, len(docs))
	failed := false
	for _, doc := range docs {
		data, err := os.ReadFile(doc.path)
		if err != nil {
			return nil, err
		}
		source := staticpage.Source(string(data))

		docDir := filepath.Dir(doc.path)
		srcPaths := o.sourcePaths
		if len(srcPaths) == 0 {
			srcPaths = []string{docDir}
		}
		// importMesh files are looked for where `gsl viewer part.gsl` looks: the document's
		// folder, its assets, then the source paths.
		meshDirs := []string{docDir, filepath.Join(docDir, "assets")}
		for _, p := range srcPaths {
			if filepath.Clean(p) != filepath.Clean(docDir) {
				meshDirs = append(meshDirs, p)
			}
		}
		var progress compile.ProgressFunc
		if o.verbose {
			progress = func(stage, label string, completed, total uint32) {
				fmt.Fprintf(stderr, "[%s] %s: %s (%d/%d)\n", doc.name, stage, label, completed, total)
			}
		}
		field := compile.Process(compile.Request{
			ID:     doc.name,
			Source: source,
			Meshes: meshimport.Dirs(meshDirs...),
		}, srcPaths, progress)

		if len(field.Errors) > 0 {
			failed = true
			for _, e := range field.Errors {
				fmt.Fprintf(stderr, "%s:%d:%d: %s\n", doc.path, e.LineNum, e.Column, e.Message)
			}
			continue
		}
		if field.WGSL == "" {
			failed = true
			fmt.Fprintf(stderr, "%s: returns no shape to draw\n", doc.path)
			continue
		}
		if field.MeshDecls != nil {
			fmt.Fprintf(stderr, "gsl export: %s places meshes with mesh(), which the exported page does not carry: "+
				"it draws the rest and reports them missing\n", doc.name)
		}
		models = append(models, exportedModel{Name: doc.name, Source: source, Field: field})
	}
	if failed {
		return nil, errDidNotCompile
	}
	return models, nil
}

// writeExport writes the folder and returns its size in bytes.
func writeExport(out, viewerDir string, models []exportedModel) (int64, error) {
	if err := staticpage.PrepareDir(out); err != nil {
		return 0, err
	}
	return staticpage.Write(staticpage.Dir(out), viewerDir, models)
}
