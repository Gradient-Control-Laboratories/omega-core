package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/program/preamble"
)

// TestSkillDocs: /docs/skill serves what `gsl skill` prints, for a model
// talking to the server: the index, a name's docs, and a search.
func TestSkillDocs(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	get := func(path string) (int, string) {
		t.Helper()
		resp, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(body)
	}

	if code, body := get("/docs/skill"); code != http.StatusOK || !strings.HasPrefix(body, preamble.GSLSkillIndex()) {
		t.Errorf("/docs/skill: status %d; want the skill index", code)
	}
	wood, _ := preamble.Lookup("material/bulk/wood")
	for path, want := range map[string]string{
		"/docs/skill/material/bulk/wood": wood,
		"/docs/skill/wood":               wood,
		"/docs/skill?member=wood":        wood,
	} {
		if code, body := get(path); code != http.StatusOK || body != want {
			t.Errorf("%s: status %d, %d bytes; want the wood docs", path, code, len(body))
		}
	}
	if code, body := get("/docs/skill?grep=countersink"); code != http.StatusOK || body != preamble.Grep("countersink") {
		t.Errorf("?grep=countersink: status %d", code)
	}
	for _, path := range []string{"/docs/skill/nope", "/docs/skill?grep=zzzqqqnothing"} {
		if code, _ := get(path); code != http.StatusNotFound {
			t.Errorf("%s: status %d, want 404", path, code)
		}
	}
}
