// Package viewer is `compile viewer`, which `gsl viewer` runs: the viewer server (cmd/serve) started on one
// GSL file, or on the examples, with the default browser opened on it. It lives in compile, as eval and skill
// do, so the toolchain keeps one Fidget-linked program, and the wheel ships it without a serve program.
//
// It sits under cmd/serve, not in compile, to use serve's server and configuration (cmd/serve/internal)
// as they are.
package viewer

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gradientcontrol.com/omega/core/cmd/serve/internal/config"
	"gradientcontrol.com/omega/core/cmd/serve/internal/server"
)

const usage = `Usage of compile viewer (gsl viewer):
  gsl viewer [flags] [part.gsl]

Starts the interactive WebGPU viewer and opens it in the default browser. With a file, the viewer
lists the .gsl files in that file's folder and opens it; Save writes back to the folder, and imports
are found there. Without one, it opens the examples that come with omega-core. Ctrl-C stops it.

Flags:
  -a, --addr <host:port>   where to listen (default 127.0.0.1:8080, or a free port when that is taken)
      --viewer-dir <dir>   the viewer's web files (default: the ones installed with this program)
      --examples-dir <dir> the folder opened without a file (default: the installed examples)
      --no-open            print the address and do not open a browser
  -v, --verbose            log compilation timings and HTTP requests
  -h, --help               display this help

Exit codes: 0 stopped, 1 the viewer could not start, 2 a bad argument.
`

// defaultAddr is tried first, so the viewer keeps one origin from run to run (it keeps its render tier and
// environment in the browser's local storage, which is per origin); a free port is taken when it is busy.
const defaultAddr = "127.0.0.1:8080"

// openBrowser launches the system's default browser on url. Overridable in tests.
var openBrowser = func(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

// options are the parsed command line.
type options struct {
	addr        string
	viewerDir   string
	examplesDir string
	file        string
	noOpen      bool
	verbose     bool
	help        bool
}

// parse reads flags before and after the one file argument, so `gsl viewer part.gsl --no-open` works as
// `gsl viewer --no-open part.gsl` does.
func parse(args []string) (options, error) {
	var o options
	fs := flag.NewFlagSet("viewer", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&o.addr, "addr", "", "")
	fs.StringVar(&o.addr, "a", "", "")
	fs.StringVar(&o.viewerDir, "viewer-dir", "", "")
	fs.StringVar(&o.examplesDir, "examples-dir", "", "")
	fs.BoolVar(&o.noOpen, "no-open", false, "")
	fs.BoolVar(&o.verbose, "verbose", false, "")
	fs.BoolVar(&o.verbose, "v", false, "")
	fs.BoolVar(&o.help, "help", false, "")
	fs.BoolVar(&o.help, "h", false, "")
	var files []string
	for {
		if err := fs.Parse(args); err != nil {
			return o, err
		}
		if fs.NArg() == 0 {
			break
		}
		files = append(files, fs.Arg(0))
		args = fs.Args()[1:]
	}
	if len(files) > 1 {
		return o, fmt.Errorf("one file at a time, not %d: %s", len(files), strings.Join(files, " "))
	}
	if len(files) == 1 {
		o.file = files[0]
	}
	return o, nil
}

// Run is `compile viewer ARGS`: it serves until ctx is cancelled and returns the exit code.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	o, err := parse(args)
	if err != nil {
		fmt.Fprintf(stderr, "gsl viewer: %v\n\n%s", err, usage)
		return 2
	}
	if o.help {
		fmt.Fprint(stdout, usage)
		return 0
	}
	cfg, start, err := configure(o, executableDir())
	if err != nil {
		fmt.Fprintf(stderr, "gsl viewer: %v\n", err)
		return 2
	}
	if err := serve(ctx, cfg, start, o, stdout, stderr); err != nil {
		fmt.Fprintf(stderr, "gsl viewer: %v\n", err)
		return 1
	}
	return 0
}

// configure is the server's configuration and the file the browser opens on ("" for the viewer's default).
func configure(o options, exeDir string) (*config.Config, string, error) {
	viewerDir := o.viewerDir
	if viewerDir == "" {
		viewerDir = findViewer(exeDir)
		if viewerDir == "" {
			return nil, "", errors.New("found no viewer next to this program or above the current folder: " +
				"name it with --viewer-dir")
		}
	} else if !isViewer(viewerDir) {
		return nil, "", fmt.Errorf("%s has no index.html: it is not the viewer's folder", viewerDir)
	}

	if o.file == "" {
		examplesDir := o.examplesDir
		if examplesDir == "" {
			examplesDir = findExamples(exeDir)
			if examplesDir == "" {
				return nil, "", errors.New("found no examples next to this program or above the current folder: " +
					"name a .gsl file, or the folder with --examples-dir")
			}
		} else if info, err := os.Stat(examplesDir); err != nil || !info.IsDir() {
			return nil, "", fmt.Errorf("%s is not a folder", examplesDir)
		}
		return &config.Config{Addr: o.addr, ExamplesDir: examplesDir, ViewerDir: viewerDir,
			SourcePaths: []string{examplesDir}, Verbose: o.verbose}, "", nil
	}

	if o.examplesDir != "" {
		return nil, "", errors.New("--examples-dir is for opening without a file; the file's own folder is served")
	}
	path, err := filepath.Abs(o.file)
	if err != nil {
		return nil, "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", err
	}
	name := filepath.Base(path)
	switch {
	case info.IsDir():
		return nil, "", fmt.Errorf("%s is a folder: name a .gsl file in it", o.file)
	case !strings.HasSuffix(name, ".gsl") || name == ".gsl":
		return nil, "", fmt.Errorf("%s is not a .gsl file", o.file)
	case strings.Contains(name, ".."):
		// The server refuses such a name (it guards against paths that climb out of the folder).
		return nil, "", fmt.Errorf("the viewer cannot open a file whose name contains \"..\": %s", name)
	}
	dir := filepath.Dir(path)
	return &config.Config{Addr: o.addr, ExamplesDir: dir, ViewerDir: viewerDir,
		SourcePaths: []string{dir}, Verbose: o.verbose}, name, nil
}

// executableDir is the folder this program is in, through symlinks; "" when it cannot be told.
func executableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return filepath.Dir(exe)
}

// installLayouts are where the viewer and the examples are relative to the program's folder: the wheel's
// (omega_core/_bin/compile, omega_core/viewer, omega_core/examples) and a checkout's bin/ (tool/build-cli.sh).
var installLayouts = []struct{ viewer, examples string }{
	{"../viewer", "../examples"},
	{"../examples/viewer", "../examples"},
}

// findViewer is the viewer's folder: installed beside the program, else examples/viewer in the current
// folder or one above it (a checkout, `go run ./cmd/compile viewer`). "" when there is none.
func findViewer(exeDir string) string {
	if exeDir != "" {
		for _, l := range installLayouts {
			if dir := filepath.Join(exeDir, l.viewer); isViewer(dir) {
				return dir
			}
		}
	}
	if root := checkoutRoot(); root != "" {
		return filepath.Join(root, "examples", "viewer")
	}
	return ""
}

// findExamples is the examples' folder, found as findViewer finds the viewer.
func findExamples(exeDir string) string {
	if exeDir != "" {
		for _, l := range installLayouts {
			if isViewer(filepath.Join(exeDir, l.viewer)) && isDir(filepath.Join(exeDir, l.examples)) {
				return filepath.Join(exeDir, l.examples)
			}
		}
	}
	if root := checkoutRoot(); root != "" {
		return filepath.Join(root, "examples")
	}
	return ""
}

// checkoutRoot is the nearest folder at or above the current one with examples/viewer in it, looking five
// levels up as serve does; "" when there is none.
func checkoutRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for i := 0; i < 5; i++ {
		if isViewer(filepath.Join(dir, "examples", "viewer")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func isViewer(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, "index.html"))
	return err == nil && !info.IsDir()
}

func isDir(dir string) bool {
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// listen listens on the configured address, or on defaultAddr and then a free port when none was named.
func listen(addr string) (net.Listener, error) {
	if addr != "" {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("listening on %s: %w", addr, err)
		}
		return ln, nil
	}
	if ln, err := net.Listen("tcp", defaultAddr); err == nil {
		return ln, nil
	}
	host, _, _ := net.SplitHostPort(defaultAddr)
	ln, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err != nil {
		return nil, fmt.Errorf("listening on %s: %w", host, err)
	}
	return ln, nil
}

// pageURL is the address the browser opens: the viewer, and the file it starts on when there is one.
func pageURL(addr net.Addr, start string) string {
	u := "http://" + addr.String() + "/"
	if start != "" {
		u += "?file=" + url.QueryEscape(start)
	}
	return u
}

func serve(ctx context.Context, cfg *config.Config, start string, o options, stdout, stderr io.Writer) error {
	ln, err := listen(cfg.Addr)
	if err != nil {
		return err
	}
	httpSrv := &http.Server{Handler: server.New(cfg).Handler()}
	page := pageURL(ln.Addr(), start)
	fmt.Fprintf(stdout, "gsl viewer: %s\n  folder: %s\n  viewer: %s\nCtrl-C to stop.\n", page, cfg.ExamplesDir, cfg.ViewerDir)

	done := make(chan error, 1)
	go func() {
		if err := httpSrv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			done <- err
			return
		}
		done <- nil
	}()

	if !o.noOpen {
		if err := openBrowser(page); err != nil {
			fmt.Fprintf(stderr, "gsl viewer: could not open a browser (%v); open %s\n", err, page)
		}
	}

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutCtx)
		return <-done
	case err := <-done:
		return err
	}
}
