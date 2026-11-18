package viewer

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gradientcontrol.com/omega/core/internal/compile"
	"gradientcontrol.com/omega/core/internal/meshimport"
)

// `compile export`, which `gsl export` runs: the viewer saved as a folder that needs no server. The folder
// is the viewer's own web files with host-static.js in host.js's place, and beside them each document with
// the field bundle it compiles to (internal/compile, the same one /api/compile answers), so the page draws
// what `gsl viewer` draws and compiles nothing.

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

const (
	// staticHost is the viewer's host for an exported page, written there as servedHost.
	staticHost = "host-static.js"
	servedHost = "host.js"
	// modelsDir holds an export's documents, and modelsIndex lists them.
	modelsDir   = "models"
	modelsIndex = "index.json"
)

// exportSkips are the viewer's files an exported page never asks for, by slash path from the viewer's folder
// (a folder skips all under it): the material catalog's large previews, operator pictures and recoloring
// tables, which only the Materials panel reads, and Manifold, which only meshing does. Neither is offered
// by a page that cannot change or mesh its document. What is left of materials/ draws the Bodies panel's rows.
var exportSkips = []string{
	"README.md",
	"materials/large",
	"materials/operators",
	"materials/coat-basis.bin",
	"materials/finish-basis.bin",
	"vendor/manifold",
}

// exportOptions are the parsed command line.
type exportOptions struct {
	output      string
	viewerDir   string
	sourcePaths []string
	files       []string
	verbose     bool
	help        bool
}

// exportedExample is one row of models/index.json: a document's name, its label in the menu, and its file.
type exportedExample struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	File  string `json:"file"`
}

// exportedIndex is models/index.json.
type exportedIndex struct {
	Examples []exportedExample `json:"examples"`
}

// exportedModel is one document's file: its source and the field bundle the source compiles to.
type exportedModel struct {
	Name   string            `json:"name"`
	Source string            `json:"source"`
	Field  *compile.Response `json:"field"`
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
	if info, err := os.Stat(filepath.Join(dir, staticHost)); err != nil || info.IsDir() {
		return "", fmt.Errorf("%s has no %s: this viewer cannot be exported", dir, staticHost)
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
		// The page compiles a document as it shows it, trimmed (main.js), and the bundle's spans
		// are positions in that text.
		source := strings.TrimSpace(strings.TrimPrefix(string(data), "\ufeff"))

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
	if err := prepareOutput(out); err != nil {
		return 0, err
	}
	size, err := copyViewer(viewerDir, out)
	if err != nil {
		return 0, err
	}
	index := exportedIndex{Examples: make([]exportedExample, 0, len(models))}
	for _, m := range models {
		label := strings.TrimSuffix(m.Name, ".gsl")
		file := label + ".json"
		n, err := writeJSONFile(filepath.Join(out, modelsDir, file), m)
		if err != nil {
			return 0, err
		}
		size += n
		index.Examples = append(index.Examples, exportedExample{Name: m.Name, Label: label, File: file})
	}
	n, err := writeJSONFile(filepath.Join(out, modelsDir, modelsIndex), index)
	if err != nil {
		return 0, err
	}
	return size + n, nil
}

// prepareOutput makes out ready to write into: a new or empty folder, or an earlier export, whose models are
// removed so the new list replaces them. Any other folder with files in it is refused, so an export never
// lands among someone's files.
func prepareOutput(out string) error {
	info, err := os.Stat(out)
	if errors.Is(err, fs.ErrNotExist) {
		return os.MkdirAll(out, 0o755)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is a file: name a folder to write", out)
	}
	entries, err := os.ReadDir(out)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	indexPath := filepath.Join(out, modelsDir, modelsIndex)
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("%s has files in it and is not an earlier export: name a new or empty folder", out)
	}
	var old exportedIndex
	if err := json.Unmarshal(data, &old); err != nil {
		return fmt.Errorf("%s: %w", indexPath, err)
	}
	for _, e := range old.Examples {
		// Only a file of the models folder: the index is read back from disk.
		if e.File == "" || e.File != filepath.Base(e.File) {
			continue
		}
		if err := os.Remove(filepath.Join(out, modelsDir, e.File)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	return nil
}

// copyViewer copies the viewer's web files into out, but for exportSkips and dotfiles, with the static host
// in the served host's place. It returns the bytes written.
func copyViewer(viewerDir, out string) (int64, error) {
	var size int64
	err := filepath.WalkDir(viewerDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(viewerDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		slash := filepath.ToSlash(rel)
		if strings.HasPrefix(d.Name(), ".") || skipped(slash) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// copyFile makes a file's folders, so a folder with nothing left to copy is not made.
		if d.IsDir() || !d.Type().IsRegular() || slash == servedHost {
			return nil
		}
		if slash == staticHost {
			rel = servedHost
		}
		n, err := copyFile(path, filepath.Join(out, rel))
		size += n
		return err
	})
	return size, err
}

func skipped(slash string) bool {
	for _, skip := range exportSkips {
		if slash == skip || strings.HasPrefix(slash, skip+"/") {
			return true
		}
	}
	return false
}

func copyFile(from, to string) (int64, error) {
	in, err := os.Open(from)
	if err != nil {
		return 0, err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return 0, err
	}
	dst, err := os.Create(to)
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(dst, in)
	if closeErr := dst.Close(); err == nil {
		err = closeErr
	}
	return n, err
}

// writeJSONFile writes v as JSON, with <, > and & as they are (WGSL is full of them), and returns its size.
func writeJSONFile(path string, v any) (int64, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return 0, err
	}
	return int64(buf.Len()), os.WriteFile(path, buf.Bytes(), 0o644)
}
