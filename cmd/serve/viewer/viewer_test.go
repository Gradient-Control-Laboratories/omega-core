package viewer

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// write puts content at dir/name, making the folders on the way.
func write(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestParse_FlagsBeforeAndAfterTheFile(t *testing.T) {
	o, err := parse([]string{"part.gsl", "--no-open", "-a", "127.0.0.1:9000"})
	require.NoError(t, err)
	assert.Equal(t, "part.gsl", o.file)
	assert.True(t, o.noOpen)
	assert.Equal(t, "127.0.0.1:9000", o.addr)

	o, err = parse([]string{"-v", "part.gsl"})
	require.NoError(t, err)
	assert.Equal(t, "part.gsl", o.file)
	assert.True(t, o.verbose)

	o, err = parse(nil)
	require.NoError(t, err)
	assert.Equal(t, "", o.file)
	assert.Equal(t, "", o.addr)

	_, err = parse([]string{"a.gsl", "b.gsl"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "one file at a time")

	_, err = parse([]string{"--bogus"})
	require.Error(t, err)
}

// A file is served from its own folder, which is also where its imports are found, and the browser starts on it.
func TestConfigure_AFileServesItsFolder(t *testing.T) {
	root := t.TempDir()
	viewerDir := filepath.Dir(write(t, root, "viewer/index.html", "<html></html>"))
	part := write(t, root, "work/my part.gsl", "return sphere(1);\n")

	cfg, start, err := configure(options{file: part, viewerDir: viewerDir}, "")
	require.NoError(t, err)
	assert.Equal(t, "my part.gsl", start)
	assert.Equal(t, filepath.Dir(part), cfg.ExamplesDir)
	assert.Equal(t, []string{filepath.Dir(part)}, cfg.SourcePaths)
	assert.Equal(t, viewerDir, cfg.ViewerDir)
}

func TestConfigure_RefusesWhatTheViewerCannotOpen(t *testing.T) {
	root := t.TempDir()
	viewerDir := filepath.Dir(write(t, root, "viewer/index.html", "<html></html>"))
	text := write(t, root, "work/notes.txt", "")
	dotted := write(t, root, "work/a..b.gsl", "")
	part := write(t, root, "work/part.gsl", "")

	for _, tc := range []struct {
		name string
		o    options
		want string
	}{
		{"missing", options{file: filepath.Join(root, "work/none.gsl"), viewerDir: viewerDir}, "no such file"},
		{"not gsl", options{file: text, viewerDir: viewerDir}, "not a .gsl file"},
		{"folder", options{file: filepath.Join(root, "work"), viewerDir: viewerDir}, "is a folder"},
		{"dots", options{file: dotted, viewerDir: viewerDir}, `contains ".."`},
		{"examples with a file", options{file: part, viewerDir: viewerDir, examplesDir: root}, "--examples-dir"},
		{"not a viewer", options{file: part, viewerDir: filepath.Join(root, "work")}, "no index.html"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := configure(tc.o, "")
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
		})
	}
}

// Installed as the wheel installs it, the program finds the viewer and the examples beside its own folder.
func TestConfigure_FindsTheWheelsViewerAndExamples(t *testing.T) {
	root := t.TempDir()
	write(t, root, "omega_core/viewer/index.html", "<html></html>")
	write(t, root, "omega_core/examples/gears.gsl", "")
	exeDir := filepath.Join(root, "omega_core", "_bin")
	require.NoError(t, os.MkdirAll(exeDir, 0o755))

	cfg, start, err := configure(options{}, exeDir)
	require.NoError(t, err)
	assert.Equal(t, "", start)
	assert.Equal(t, filepath.Join(root, "omega_core", "viewer"), cfg.ViewerDir)
	assert.Equal(t, filepath.Join(root, "omega_core", "examples"), cfg.ExamplesDir)
	assert.Equal(t, []string{cfg.ExamplesDir}, cfg.SourcePaths)
}

// Built into a checkout's bin/ (tool/build-cli.sh), it finds the checkout's examples/viewer.
func TestConfigure_FindsACheckoutsViewerFromBin(t *testing.T) {
	root := t.TempDir()
	write(t, root, "examples/viewer/index.html", "<html></html>")
	exeDir := filepath.Join(root, "bin")
	require.NoError(t, os.MkdirAll(exeDir, 0o755))

	cfg, _, err := configure(options{}, exeDir)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, "examples", "viewer"), cfg.ViewerDir)
	assert.Equal(t, filepath.Join(root, "examples"), cfg.ExamplesDir)
}

// run starts Run and returns the page address it prints, and a function that stops it and returns its exit code.
func run(t *testing.T, args []string) (string, func() int) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	out, w := io.Pipe()
	code := make(chan int, 1)
	go func() {
		code <- Run(ctx, args, w, io.Discard)
		w.Close()
	}()
	line, err := bufio.NewReader(out).ReadString('\n')
	require.NoError(t, err)
	go func() { _, _ = io.Copy(io.Discard, out) }()
	page := strings.TrimSpace(strings.TrimPrefix(line, "gsl viewer: "))
	return page, func() int {
		cancel()
		select {
		case c := <-code:
			return c
		case <-time.After(5 * time.Second):
			t.Fatal("the viewer did not stop")
			return -1
		}
	}
}

func get(t *testing.T, url string) (int, []byte) {
	t.Helper()
	res, err := http.Get(url)
	require.NoError(t, err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res.StatusCode, body
}

// The server answers for the file's folder, opens the browser on the file, and stops cleanly when interrupted.
func TestRun_ServesAFileAndOpensTheBrowserOnIt(t *testing.T) {
	root := t.TempDir()
	viewerDir := filepath.Dir(write(t, root, "viewer/index.html", "<html>viewer</html>"))
	part := write(t, root, "work/my part.gsl", "return sphere(1);\n")
	write(t, root, "work/other.gsl", "return sphere(2);\n")

	opened := make(chan string, 1)
	saved := openBrowser
	openBrowser = func(url string) error { opened <- url; return nil }
	t.Cleanup(func() { openBrowser = saved })

	page, stop := run(t, []string{"--viewer-dir", viewerDir, "-a", "127.0.0.1:0", part})
	assert.True(t, strings.HasSuffix(page, "/?file=my+part.gsl"), page)
	select {
	case url := <-opened:
		assert.Equal(t, page, url)
	case <-time.After(5 * time.Second):
		t.Fatal("the browser was not opened")
	}

	status, body := get(t, page)
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, string(body), "viewer")

	base := page[:strings.Index(page, "?")]
	status, body = get(t, base+"api/examples")
	assert.Equal(t, http.StatusOK, status)
	var list struct{ Examples []struct{ Name string } }
	require.NoError(t, json.Unmarshal(body, &list))
	require.Equal(t, 2, len(list.Examples))
	assert.Equal(t, "my part.gsl", list.Examples[0].Name)
	assert.Equal(t, "other.gsl", list.Examples[1].Name)

	status, body = get(t, base+"api/example?name=my%20part.gsl")
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, string(body), "sphere(1)")

	assert.Equal(t, 0, stop())
}

func TestRun_NoOpenLeavesTheBrowserAlone(t *testing.T) {
	root := t.TempDir()
	viewerDir := filepath.Dir(write(t, root, "viewer/index.html", "<html></html>"))
	examples := filepath.Dir(write(t, root, "examples/gears.gsl", "return sphere(1);\n"))

	saved := openBrowser
	openBrowser = func(url string) error { t.Errorf("opened %s with --no-open", url); return nil }
	t.Cleanup(func() { openBrowser = saved })

	page, stop := run(t, []string{"--no-open", "--viewer-dir", viewerDir, "--examples-dir", examples, "-a", "127.0.0.1:0"})
	assert.False(t, strings.Contains(page, "?file="), page)
	status, _ := get(t, page+"api/example?name=gears.gsl")
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, 0, stop())
}

func TestRun_ExitCodes(t *testing.T) {
	assert.Equal(t, 0, Run(context.Background(), []string{"--help"}, io.Discard, io.Discard))
	assert.Equal(t, 2, Run(context.Background(), []string{"--bogus"}, io.Discard, io.Discard))
	assert.Equal(t, 2, Run(context.Background(), []string{"--viewer-dir", t.TempDir(), "x.gsl"}, io.Discard, io.Discard))

	root := t.TempDir()
	viewerDir := filepath.Dir(write(t, root, "viewer/index.html", "<html></html>"))
	part := write(t, root, "part.gsl", "")
	assert.Equal(t, 1, Run(context.Background(), []string{"--no-open", "--viewer-dir", viewerDir, "-a", "256.0.0.1:1", part},
		io.Discard, io.Discard))
}
