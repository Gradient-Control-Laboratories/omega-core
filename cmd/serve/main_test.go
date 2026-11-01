package main

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRun_HelpFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--help"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error on --help: %v", err)
	}
	if !strings.Contains(stdout.String(), "Usage: serve") {
		t.Errorf("expected help text in stdout, got %q", stdout.String())
	}
}

func TestRun_InvalidFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--bad-flag"}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected error on invalid flag")
	}
}

func TestRun_InvalidListenAddr(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--addr", "127.0.0.1:99999"}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected error on invalid listen address")
	}
}

func TestRun_ServerLifecycle(t *testing.T) {
	exDir := t.TempDir()
	viewDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(exDir, "demo.gsl"), []byte("return sphere(5);\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(viewDir, "index.html"), []byte("<!DOCTYPE html><html></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stdout, stderr bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, []string{
			"--addr", "127.0.0.1:0",
			"--examples-dir", exDir,
			"--viewer-dir", viewDir,
			"--verbose",
		}, &stdout, &stderr)
	}()

	// Wait for listener address in stdout
	var baseURL string
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		out := stdout.String()
		if idx := strings.Index(out, "http://"); idx != -1 {
			rest := out[idx:]
			if end := strings.IndexAny(rest, " \n"); end != -1 {
				baseURL = rest[:end]
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	if baseURL == "" {
		t.Fatalf("server did not report listening URL; stdout=%q stderr=%q", stdout.String(), stderr.String())
	}

	res, err := http.Get(baseURL + "/api/examples")
	if err != nil {
		t.Fatalf("GET /api/examples failed: %v", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}

	cancel()
	select {
	case runErr := <-done:
		if runErr != nil {
			t.Errorf("unexpected error on clean shutdown: %v", runErr)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for server shutdown")
	}
}

func TestRun_OpenBrowserFlag(t *testing.T) {
	exDir := t.TempDir()
	viewDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(exDir, "demo.gsl"), []byte("return sphere(5);\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(viewDir, "index.html"), []byte("<!DOCTYPE html><html></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	openedCh := make(chan string, 1)
	origOpen := openBrowserFunc
	openBrowserFunc = func(url string) error {
		openedCh <- url
		return nil
	}
	defer func() { openBrowserFunc = origOpen }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stdout, stderr bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, []string{
			"--addr", "127.0.0.1:0",
			"--examples-dir", exDir,
			"--viewer-dir", viewDir,
			"--open",
		}, &stdout, &stderr)
	}()

	select {
	case url := <-openedCh:
		if !strings.HasPrefix(url, "http://127.0.0.1:") {
			t.Errorf("expected opened URL to start with http://127.0.0.1:, got %q", url)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for browser open call")
	}

	cancel()
	select {
	case runErr := <-done:
		if runErr != nil {
			t.Errorf("unexpected error on clean shutdown: %v", runErr)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for server shutdown")
	}
}
