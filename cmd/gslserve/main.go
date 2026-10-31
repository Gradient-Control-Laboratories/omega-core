package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gradientcontrol.com/omega/core/cmd/gslserve/internal/config"
	"gradientcontrol.com/omega/core/cmd/gslserve/internal/server"
)

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

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", cfg.Addr, err)
	}

	srv := server.New(cfg)
	httpSrv := &http.Server{
		Handler: srv.Handler(),
	}

	fmt.Fprintf(stdout, "gslserve listening on http://%s (examples: %s, viewer: %s)\n",
		ln.Addr().String(), cfg.ExamplesDir, cfg.ViewerDir)

	errCh := make(chan error, 1)
	go func() {
		if serveErr := httpSrv.Serve(ln); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		} else {
			errCh <- nil
		}
	}()

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
