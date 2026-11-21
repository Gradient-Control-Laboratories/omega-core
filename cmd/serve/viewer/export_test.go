package viewer

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

const exportedPart = `
r = 10;
r.min = 2;
r.max = 20;
body = sphere(r);
return body;
`

// fakeViewer is a folder with the files copyViewer tells apart: the page, both hosts, a file of each
// skipped kind, and a dotfile.
func fakeViewer(t *testing.T, root string) string {
	t.Helper()
	dir := filepath.Dir(write(t, root, "viewer/index.html", "<html></html>"))
	write(t, dir, "main.js", "import * as host from './host.js';")
	write(t, dir, "host.js", "// the server's host")
	write(t, dir, "host-static.js", "// the static host")
	write(t, dir, "render/viewer.js", "// the viewport")
	write(t, dir, "materials/catalog.json", "{}")
	write(t, dir, "materials/large/metal.chrome.png", "png")
	write(t, dir, "materials/operators/coat.png", "png")
	write(t, dir, "materials/coat-basis.bin", "bin")
	write(t, dir, "materials/finish-basis.bin", "bin")
	write(t, dir, "vendor/manifold/manifold.js", "// manifold")
	write(t, dir, "README.md", "# viewer")
	write(t, dir, ".DS_Store", "x")
	return dir
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TestParseExport_FlagsAmongTheFiles(t *testing.T) {
	o, err := parseExport([]string{"a.gsl", "-o", "site", "b.gsl", "-I", "lib", "--source-path", "more", "-v"})
	require.NoError(t, err)
	assert.Equal(t, []string{"a.gsl", "b.gsl"}, o.files)
	assert.Equal(t, "site", o.output)
	assert.Equal(t, []string{"lib", "more"}, o.sourcePaths)
	assert.True(t, o.verbose)

	_, err = parseExport([]string{"--bogus"})
	require.Error(t, err)
}

func TestExport_BadArguments(t *testing.T) {
	root := t.TempDir()
	viewerDir := fakeViewer(t, root)
	part := write(t, root, "work/part.gsl", exportedPart)
	other := write(t, root, "other/part.gsl", exportedPart)
	notGSL := write(t, root, "work/notes.txt", "notes")
	noStatic := filepath.Dir(write(t, root, "old/index.html", "<html></html>"))

	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"no file", []string{"--viewer-dir", viewerDir}, "name a .gsl file"},
		{"not gsl", []string{"--viewer-dir", viewerDir, notGSL}, "is not a .gsl file"},
		{"a folder", []string{"--viewer-dir", viewerDir, filepath.Dir(part)}, "is a folder"},
		{"one name twice", []string{"--viewer-dir", viewerDir, part, other}, "have the same name"},
		{"not a viewer", []string{"--viewer-dir", filepath.Dir(part), part}, "has no index.html"},
		{"a viewer without the static host", []string{"--viewer-dir", noStatic, part}, "cannot be exported"},
	} {
		var stdout, stderr bytes.Buffer
		out := filepath.Join(root, "out")
		code := Export(append([]string{"-o", out}, tc.args...), &stdout, &stderr)
		assert.Equal(t, 2, code, tc.name)
		assert.Contains(t, stderr.String(), tc.want, tc.name)
		assert.False(t, exists(out), tc.name)
	}

	var stdout, stderr bytes.Buffer
	assert.Equal(t, 0, Export([]string{"--help"}, &stdout, &stderr))
	assert.Contains(t, stdout.String(), "gsl export [flags] part.gsl")
}

// The folder is the viewer with the static host in the served host's place, and each document with the field
// it compiles to, listed in the order named.
func TestExport_WritesTheViewerAndEachModel(t *testing.T) {
	root := t.TempDir()
	viewerDir := fakeViewer(t, root)
	part := write(t, root, "work/my part.gsl", exportedPart)
	second := write(t, root, "work/ball.gsl", "return sphere(3);\n")
	out := filepath.Join(root, "site")

	var stdout, stderr bytes.Buffer
	code := Export([]string{"-o", out, "--viewer-dir", viewerDir, part, second}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	assert.Contains(t, stdout.String(), "2 models")

	assert.Equal(t, "<html></html>", readFile(t, filepath.Join(out, "index.html")))
	assert.Equal(t, "// the viewport", readFile(t, filepath.Join(out, "render", "viewer.js")))
	assert.Equal(t, "{}", readFile(t, filepath.Join(out, "materials", "catalog.json")))
	assert.Equal(t, "// the static host", readFile(t, filepath.Join(out, "host.js")))
	for _, gone := range []string{"host-static.js", "README.md", ".DS_Store", "materials/large", "materials/operators",
		"materials/coat-basis.bin", "materials/finish-basis.bin", "vendor/manifold", "vendor"} {
		assert.False(t, exists(filepath.Join(out, filepath.FromSlash(gone))), gone)
	}

	var index exportedIndex
	require.NoError(t, json.Unmarshal([]byte(readFile(t, filepath.Join(out, "models", "index.json"))), &index))
	assert.Equal(t, []exportedExample{
		{Name: "my part.gsl", Label: "my part", File: "my part.json"},
		{Name: "ball.gsl", Label: "ball", File: "ball.json"},
	}, index.Examples)

	raw := readFile(t, filepath.Join(out, "models", "my part.json"))
	var model exportedModel
	require.NoError(t, json.Unmarshal([]byte(raw), &model))
	assert.Equal(t, "my part.gsl", model.Name)
	// The source is the document trimmed, as the page compiles it, so the bundle's spans are positions in it.
	assert.Equal(t, strings.TrimSpace(exportedPart), model.Source)
	require.NotNil(t, model.Field)
	assert.Equal(t, 0, len(model.Field.Errors))
	assert.Contains(t, model.Field.WGSL, "fn sdShape")
	require.Equal(t, 1, len(model.Field.Sliders))
	assert.Equal(t, "r", model.Field.Sliders[0].VariableName)
	assert.Equal(t, 1, model.Field.Sliders[0].Span.StartLine)
	// WGSL keeps its angle brackets in the file.
	assert.Contains(t, raw, "vec3<f32>")
}

// A document that does not compile is reported with its place, and no folder is written, whatever else was named.
func TestExport_ADocumentWithErrorsWritesNothing(t *testing.T) {
	root := t.TempDir()
	viewerDir := fakeViewer(t, root)
	good := write(t, root, "work/ball.gsl", "return sphere(3);\n")
	bad := write(t, root, "work/broken.gsl", "return nothingByThisName(3);\n")
	noReturn := write(t, root, "work/no return.gsl", "x = 3;\n")
	out := filepath.Join(root, "site")

	var stdout, stderr bytes.Buffer
	code := Export([]string{"-o", out, "--viewer-dir", viewerDir, good, bad}, &stdout, &stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), bad+":1:")
	assert.Contains(t, stderr.String(), "nothing was written")
	assert.False(t, exists(out))

	stderr.Reset()
	code = Export([]string{"-o", out, "--viewer-dir", viewerDir, noReturn}, &stdout, &stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), noReturn+":1:1: ")
	assert.False(t, exists(out))
}

// An export goes into a new or empty folder, or over an earlier export, whose models it replaces; a folder with
// anything else in it is left alone.
func TestExport_TheFolderItWritesInto(t *testing.T) {
	root := t.TempDir()
	viewerDir := fakeViewer(t, root)
	ball := write(t, root, "work/ball.gsl", "return sphere(3);\n")
	part := write(t, root, "work/part.gsl", exportedPart)

	theirs := filepath.Join(root, "theirs")
	kept := write(t, theirs, "thesis.txt", "do not lose")
	var stdout, stderr bytes.Buffer
	code := Export([]string{"-o", theirs, "--viewer-dir", viewerDir, ball}, &stdout, &stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), "is not an earlier export")
	assert.Equal(t, "do not lose", readFile(t, kept))
	assert.False(t, exists(filepath.Join(theirs, "index.html")))

	stderr.Reset()
	code = Export([]string{"-o", kept, "--viewer-dir", viewerDir, ball}, &stdout, &stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), "is a file")

	out := filepath.Join(root, "site")
	require.NoError(t, os.MkdirAll(out, 0o755))
	stderr.Reset()
	require.Equal(t, 0, Export([]string{"-o", out, "--viewer-dir", viewerDir, ball}, &stdout, &stderr), stderr.String())
	assert.True(t, exists(filepath.Join(out, "models", "ball.json")))
	added := write(t, out, "CNAME", "example.org")

	require.Equal(t, 0, Export([]string{"-o", out, "--viewer-dir", viewerDir, part}, &stdout, &stderr), stderr.String())
	assert.False(t, exists(filepath.Join(out, "models", "ball.json")))
	assert.True(t, exists(filepath.Join(out, "models", "part.json")))
	assert.Equal(t, "example.org", readFile(t, added))
	var index exportedIndex
	require.NoError(t, json.Unmarshal([]byte(readFile(t, filepath.Join(out, "models", "index.json"))), &index))
	assert.Equal(t, []exportedExample{{Name: "part.gsl", Label: "part", File: "part.json"}}, index.Examples)
}

// Without -o the folder is named after the first file, in the working directory.
func TestExport_DefaultFolder(t *testing.T) {
	root := t.TempDir()
	viewerDir := fakeViewer(t, root)
	ball := write(t, root, "work/ball.gsl", "return sphere(3);\n")
	t.Chdir(root)

	var stdout, stderr bytes.Buffer
	require.Equal(t, 0, Export([]string{"--viewer-dir", viewerDir, ball}, &stdout, &stderr), stderr.String())
	assert.True(t, exists(filepath.Join(root, "ball-html", "models", "ball.json")))
}

// realViewer is the repository's viewer, which the tests below read.
const realViewer = "../../../examples/viewer"

var hostExport = regexp.MustCompile(`(?m)^export (?:const|let|async function|function) (\w+)`)

// main.js calls the host through one module, which is host.js for a server and host-static.js in an exported
// page, so the two export the same names.
func TestRealViewer_TheHostsExportTheSameNames(t *testing.T) {
	names := func(file string) []string {
		var out []string
		for _, m := range hostExport.FindAllStringSubmatch(readFile(t, filepath.Join(realViewer, file)), -1) {
			out = append(out, m[1])
		}
		sort.Strings(out)
		return out
	}
	served := names(servedHost)
	require.True(t, len(served) > 0)
	assert.Equal(t, served, names(staticHost))

	// And main.js asks for nothing the hosts do not export.
	has := map[string]bool{}
	for _, name := range served {
		has[name] = true
	}
	uses := regexp.MustCompile(`\bhost\.(\w+)`).FindAllStringSubmatch(readFile(t, filepath.Join(realViewer, "main.js")), -1)
	require.True(t, len(uses) > 0)
	for _, m := range uses {
		// "host.js" is the file's name, in the import and the comments.
		if m[1] != "js" {
			assert.True(t, has[m[1]], "main.js uses host."+m[1])
		}
	}
}

// What an export leaves out is named by path, so each path is one the viewer has.
func TestRealViewer_TheSkippedFilesExist(t *testing.T) {
	for _, skip := range exportSkips {
		assert.True(t, exists(filepath.Join(realViewer, filepath.FromSlash(skip))), skip)
	}
}

// The repository's viewer exports to a page with its modules, and without the material catalog's bulk.
func TestRealViewer_Exports(t *testing.T) {
	root := t.TempDir()
	ball := write(t, root, "work/ball.gsl", "return sphere(3);\n")
	out := filepath.Join(root, "site")

	var stdout, stderr bytes.Buffer
	require.Equal(t, 0, Export([]string{"-o", out, "--viewer-dir", realViewer, ball}, &stdout, &stderr), stderr.String())

	assert.Equal(t, readFile(t, filepath.Join(realViewer, staticHost)), readFile(t, filepath.Join(out, servedHost)))
	for _, file := range []string{"index.html", "main.js", "style.css", "render/viewer.js", "render/noise_tile.bin",
		"vendor/three/three.webgpu.js", "vendor/codemirror/codemirror.js", "materials/catalog.json", "models/ball.json"} {
		assert.True(t, exists(filepath.Join(out, filepath.FromSlash(file))), file)
	}

	var size int64
	require.NoError(t, filepath.Walk(out, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return err
	}))
	assert.True(t, size < 16<<20, "the export is %d bytes", size)
}
