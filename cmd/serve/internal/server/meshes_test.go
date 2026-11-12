package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/test/stlbox"
)

var (
	boxLo = [3]float64{-10, -5, -3}
	boxHi = [3]float64{10, 5, 3}
)

func postCompile(t *testing.T, url, source string) map[string]any {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"id": "1", "clientId": "c", "source": source})
	res, err := http.Post(url+"/api/compile", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

// An STL the Imports panel uploads is what the document's importMesh of that
// name reads.
func TestAssetUploadFeedsCompile(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	source := "importMesh 0 \"up.stl\";\nreturn bvhSdf(0);\n"
	if out := postCompile(t, ts.URL, source); out["errors"] == nil {
		t.Fatalf("compiled before the upload: %v", out)
	}

	res, err := http.Post(ts.URL+"/api/asset?name=up.stl", "application/octet-stream",
		bytes.NewReader(stlbox.Bytes(boxLo, boxHi)))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("upload: HTTP %d", res.StatusCode)
	}

	out := postCompile(t, ts.URL, source)
	if out["errors"] != nil {
		t.Fatalf("compile errors: %v", out["errors"])
	}
	if bindings, _ := out["meshSdfBindings"].([]any); len(bindings) != 1 {
		t.Fatalf("got %v mesh bindings, want 1", out["meshSdfBindings"])
	}
}

// A file beside the examples, or in examples/assets, is found without an
// upload.
func TestCompileFindsMeshesOnDisk(t *testing.T) {
	srv, examplesDir, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	stlbox.Write(t, examplesDir, "beside.stl", boxLo, boxHi)
	assets := filepath.Join(examplesDir, "assets")
	if err := os.MkdirAll(assets, 0o755); err != nil {
		t.Fatal(err)
	}
	stlbox.Write(t, assets, "asset.stl", boxLo, boxHi)

	for _, name := range []string{"beside.stl", "asset.stl"} {
		out := postCompile(t, ts.URL, "importMesh 0 \""+name+"\";\nreturn vdbSdf(0);\n")
		if out["errors"] != nil {
			t.Errorf("%s: %v", name, out["errors"])
		}
	}
}

func TestAssetUploadRefusals(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	for _, c := range []struct {
		name string
		body []byte
		says string
	}{
		{"part.obj", stlbox.Bytes(boxLo, boxHi), ".stl filename"},
		{"../up.stl", stlbox.Bytes(boxLo, boxHi), ".stl filename"},
		{".hidden.stl", stlbox.Bytes(boxLo, boxHi), ".stl filename"},
		{"junk.stl", []byte("not an stl"), "junk.stl"},
	} {
		res, err := http.Post(ts.URL+"/api/asset?name="+c.name, "application/octet-stream", bytes.NewReader(c.body))
		if err != nil {
			t.Fatal(err)
		}
		var out map[string]string
		_ = json.NewDecoder(res.Body).Decode(&out)
		res.Body.Close()
		if res.StatusCode != http.StatusBadRequest || !strings.Contains(out["error"], c.says) {
			t.Errorf("%s: HTTP %d, %q", c.name, res.StatusCode, out["error"])
		}
	}
}
