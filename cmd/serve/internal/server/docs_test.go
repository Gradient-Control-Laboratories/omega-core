package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gradientcontrol.com/omega/core/program/preamble"
)

// TestHelpDocsAreGenerated: the Help panel's two documents come from the
// preamble, not from files in the viewer directory.
func TestHelpDocsAreGenerated(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	for path, want := range map[string]string{
		"/docs/gsl_description.md": preamble.GSLDescription,
		"/docs/gsl_libraries.md":   preamble.LibrariesDescriptiveText(),
	} {
		resp, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK || string(body) != want {
			t.Errorf("%s: status %d, %d bytes; want the preamble's %d", path, resp.StatusCode, len(body), len(want))
		}
		if ct := resp.Header.Get("Content-Type"); ct != "text/markdown; charset=utf-8" {
			t.Errorf("%s: Content-Type %q", path, ct)
		}
	}
}
