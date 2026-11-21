package server_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// exportableViewer gives the test server's viewer what a real one has that an export tells apart: both
// hosts, a script, and a file the export leaves out.
func exportableViewer(t *testing.T, viewerDir string) {
	t.Helper()
	for name, content := range map[string]string{
		"main.js":                     "import * as host from './host.js';",
		"host.js":                     "// the server's host",
		"host-static.js":              "// the static host",
		"vendor/manifold/manifold.js": "// manifold",
	} {
		path := filepath.Join(viewerDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func postExport(t *testing.T, url string, body map[string]any) *http.Response {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.Post(url+"/api/export", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("POST /api/export failed: %v", err)
	}
	t.Cleanup(func() { res.Body.Close() })
	return res
}

func exportError(t *testing.T, res *http.Response) string {
	t.Helper()
	var reply struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&reply); err != nil {
		t.Fatalf("the error reply is not JSON: %v", err)
	}
	return reply.Error
}

// POST /api/export answers a zip of what `gsl export` writes for the document sent, under one folder
// named after it.
func TestExport_ZipsTheViewerAndTheDocument(t *testing.T) {
	srv, _, viewerDir := setupTestServer(t)
	exportableViewer(t, viewerDir)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	// The source is not a file of the examples folder, and has space around it: the page exports
	// what its editor shows, as it compiles it.
	source := "r = 5; r.min = 1; r.max = 10; return sphere(r);"
	res := postExport(t, ts.URL, map[string]any{"name": "part.gsl", "source": "\n" + source + "\n\n"})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", res.StatusCode, exportError(t, res))
	}
	if got := res.Header.Get("Content-Type"); got != "application/zip" {
		t.Errorf("Content-Type = %q, want application/zip", got)
	}
	if got := res.Header.Get("Content-Disposition"); got != "attachment; filename=part-html.zip" {
		t.Errorf("Content-Disposition = %q", got)
	}
	if got := res.Header.Get("X-Export-Note"); got != "" {
		t.Errorf("a document that places no meshes has nothing to note, got %q", got)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("the reply is not a zip: %v", err)
	}
	files := map[string]string{}
	for _, f := range archive.File {
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		files[f.Name] = string(content)
	}

	for _, name := range []string{"part-html/index.html", "part-html/main.js", "part-html/host.js", "part-html/models/index.json", "part-html/models/part.json"} {
		if _, ok := files[name]; !ok {
			t.Errorf("the zip has no %s", name)
		}
	}
	for name := range files {
		if !strings.HasPrefix(name, "part-html/") {
			t.Errorf("%s is outside the export's folder", name)
		}
	}
	if got := files["part-html/host.js"]; got != "// the static host" {
		t.Errorf("host.js = %q, want the static host", got)
	}
	for _, name := range []string{"part-html/host-static.js", "part-html/vendor/manifold/manifold.js"} {
		if _, ok := files[name]; ok {
			t.Errorf("the zip should not carry %s", name)
		}
	}

	var index struct {
		Examples []struct{ Name, Label, File string }
	}
	if err := json.Unmarshal([]byte(files["part-html/models/index.json"]), &index); err != nil {
		t.Fatal(err)
	}
	if len(index.Examples) != 1 || index.Examples[0].Name != "part.gsl" || index.Examples[0].File != "part.json" {
		t.Errorf("index = %+v", index.Examples)
	}
	var model struct {
		Name   string
		Source string
		Field  struct {
			WGSL    string           `json:"wgsl"`
			Sliders []map[string]any `json:"sliders"`
		}
	}
	if err := json.Unmarshal([]byte(files["part-html/models/part.json"]), &model); err != nil {
		t.Fatal(err)
	}
	if model.Name != "part.gsl" || model.Source != source {
		t.Errorf("model = %q, source %q", model.Name, model.Source)
	}
	if !strings.Contains(model.Field.WGSL, "fn sdShape") {
		t.Errorf("the model carries no compiled field")
	}
	if len(model.Field.Sliders) != 1 {
		t.Errorf("sliders = %d, want the document's one", len(model.Field.Sliders))
	}
}

// A request the server cannot export is an error reply, in JSON, and no zip.
func TestExport_Refusals(t *testing.T) {
	srv, _, viewerDir := setupTestServer(t)
	exportableViewer(t, viewerDir)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	for _, tc := range []struct {
		what   string
		body   map[string]any
		status int
		want   string
	}{
		{"a document with an error", map[string]any{"name": "part.gsl", "source": "return badCall(1);"},
			http.StatusUnprocessableEntity, "does not compile (1:"},
		{"an empty document", map[string]any{"name": "part.gsl", "source": "  \n"},
			http.StatusBadRequest, "empty"},
		{"a name that is a path", map[string]any{"name": "../part.gsl", "source": "return sphere(5);"},
			http.StatusBadRequest, "invalid .gsl filename"},
		{"a name that is not a .gsl file", map[string]any{"name": "part.txt", "source": "return sphere(5);"},
			http.StatusBadRequest, "invalid .gsl filename"},
	} {
		res := postExport(t, ts.URL, tc.body)
		if res.StatusCode != tc.status {
			t.Errorf("%s: status %d, want %d", tc.what, res.StatusCode, tc.status)
		}
		if got := exportError(t, res); !strings.Contains(got, tc.want) {
			t.Errorf("%s: error %q, want it to say %q", tc.what, got, tc.want)
		}
	}
}

// A document sent without a name is exported as model.gsl.
func TestExport_AnUnnamedDocument(t *testing.T) {
	srv, _, viewerDir := setupTestServer(t)
	exportableViewer(t, viewerDir)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	res := postExport(t, ts.URL, map[string]any{"source": "return sphere(5);"})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", res.StatusCode, exportError(t, res))
	}
	if got := res.Header.Get("Content-Disposition"); got != "attachment; filename=model-html.zip" {
		t.Errorf("Content-Disposition = %q", got)
	}
}

// A viewer folder without the static host cannot be exported, and the server says so.
func TestExport_AViewerWithoutTheStaticHost(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	res := postExport(t, ts.URL, map[string]any{"name": "part.gsl", "source": "return sphere(5);"})
	if res.StatusCode != http.StatusNotImplemented {
		t.Errorf("status %d, want 501", res.StatusCode)
	}
	if got := exportError(t, res); !strings.Contains(got, "host-static.js") {
		t.Errorf("error %q should name the missing host", got)
	}
}
