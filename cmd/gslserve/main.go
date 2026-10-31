package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"gradientcontrol.com/omega/core/cmd/gslserve/internal/config"
	"gradientcontrol.com/omega/core/cmd/gslserve/internal/server"
)

// openBrowserFunc launches the system default web browser for url. Overridable in tests.
var openBrowserFunc = openBrowser

func openBrowser(url string) error {
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

func resolveDefaultDirs(cfg *config.Config) {
	if _, err := os.Stat(filepath.Join(cfg.ViewerDir, "index.html")); err == nil {
		return
	}
	if cfg.ExamplesDir != "examples" || cfg.ViewerDir != "examples/viewer" {
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	dir := cwd
	for i := 0; i < 5; i++ {
		candidateViewer := filepath.Join(dir, "examples", "viewer")
		if _, err := os.Stat(filepath.Join(candidateViewer, "index.html")); err == nil {
			cfg.ExamplesDir = filepath.Join(dir, "examples")
			cfg.ViewerDir = candidateViewer
			if len(cfg.SourcePaths) == 1 && cfg.SourcePaths[0] == "examples" {
				cfg.SourcePaths = []string{cfg.ExamplesDir}
			}
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
}

// Run parses CLI arguments and runs the gslserve HTTP server until ctx is canceled.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	cfg, err := config.ParseFlags(args)
	if err != nil {
		fmt.Fprint(stderr, config.Usage())
		return err
	}

	if cfg.Help {
		fmt.Fprint(stdout, config.Usage())
		return nil
	}

	resolveDefaultDirs(cfg)

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", cfg.Addr, err)
	}

	srv := server.New(cfg)
	httpSrv := &http.Server{
		Handler: srv.Handler(),
	}

	url := fmt.Sprintf("http://%s", ln.Addr().String())
	fmt.Fprintf(stdout, "gslserve listening on %s (examples: %s, viewer: %s)\n",
		url, cfg.ExamplesDir, cfg.ViewerDir)

	errCh := make(chan error, 1)
	go func() {
		if serveErr := httpSrv.Serve(ln); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		} else {
			errCh <- nil
		}
	}()

	if cfg.Open {
		if openErr := openBrowserFunc(url); openErr != nil {
			fmt.Fprintf(stderr, "gslserve: failed to open browser: %v\n", openErr)
		}
	}

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutCtx)
		return <-errCh
	case serveErr := <-errCh:
		return serveErr
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := Run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "gslserve: %v\n", err)
		os.Exit(1)
	}
}
