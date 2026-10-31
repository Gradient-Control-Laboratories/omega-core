package server_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"gradientcontrol.com/omega/core/cmd/gslserve/internal/config"
	"gradientcontrol.com/omega/core/cmd/gslserve/internal/server"
)

func setupTestServer(t *testing.T) (*server.Server, string, string) {
	t.Helper()
	examplesDir := t.TempDir()
	viewerDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(examplesDir, "gear.gsl"), []byte("return sphere(5);\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(examplesDir, "box.gsl"), []byte("return box([4, 4, 4]);\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(examplesDir, "notes.txt"), []byte("ignore me"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(viewerDir, "index.html"), []byte("<!DOCTYPE html><title>Viewer</title>"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Addr:        "127.0.0.1:0",
		ExamplesDir: examplesDir,
		ViewerDir:   viewerDir,
		SourcePaths: []string{examplesDir},
	}
	return server.New(cfg), examplesDir, viewerDir
}

func TestStaticViewerServing(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}
	if cc := res.Header.Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("expected Cache-Control 'no-cache', got %q", cc)
	}
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), "<title>Viewer</title>") {
		t.Errorf("unexpected index.html body: %s", string(body))
	}
}

func TestExamplesAndExampleEndpoints(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	// 1. GET /api/examples
	res, err := http.Get(ts.URL + "/api/examples")
	if err != nil {
		t.Fatalf("GET /api/examples failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}
	var listResp struct {
		Examples []server.ExampleItem `json:"examples"`
	}
	if err := json.NewDecoder(res.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode examples failed: %v", err)
	}
	if len(listResp.Examples) != 2 {
		t.Fatalf("expected 2 .gsl examples, got %+v", listResp.Examples)
	}
	if listResp.Examples[0].Name != "box.gsl" || listResp.Examples[0].Label != "box" {
		t.Errorf("expected first sorted example 'box.gsl', got %+v", listResp.Examples[0])
	}
	if listResp.Examples[1].Name != "gear.gsl" || listResp.Examples[1].Label != "gear" {
		t.Errorf("expected second sorted example 'gear.gsl', got %+v", listResp.Examples[1])
	}

	// 2. GET /api/example?name=gear.gsl
	exRes, err := http.Get(ts.URL + "/api/example?name=gear.gsl")
	if err != nil {
		t.Fatalf("GET /api/example failed: %v", err)
	}
	defer exRes.Body.Close()
	if exRes.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", exRes.StatusCode)
	}
	var exPayload struct {
		Name   string `json:"name"`
		Source string `json:"source"`
	}
	if err := json.NewDecoder(exRes.Body).Decode(&exPayload); err != nil {
		t.Fatalf("decode example failed: %v", err)
	}
	if exPayload.Name != "gear.gsl" || exPayload.Source != "return sphere(5);\n" {
		t.Errorf("unexpected example payload: %+v", exPayload)
	}

	// 3. Path traversal & invalid extension rejection
	for _, badName := range []string{"", "../secret.gsl", "sub/gear.gsl", "notes.txt", ".gsl"} {
		badRes, err := http.Get(ts.URL + "/api/example?name=" + badName)
		if err != nil {
			t.Fatal(err)
		}
		badRes.Body.Close()
		if badRes.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for name=%q, got %d", badName, badRes.StatusCode)
		}
	}

	// 4. Non-existent file returns 404
	notFoundRes, err := http.Get(ts.URL + "/api/example?name=missing.gsl")
	if err != nil {
		t.Fatal(err)
	}
	notFoundRes.Body.Close()
	if notFoundRes.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 Not Found for missing.gsl, got %d", notFoundRes.StatusCode)
	}
}

func TestSaveEndpoint(t *testing.T) {
	srv, examplesDir, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	// Valid save
	saveBody, _ := json.Marshal(map[string]string{
		"name":   "gear_copy.gsl",
		"source": "return sphere(12);\n",
	})
	res, err := http.Post(ts.URL+"/api/save", "application/json", bytes.NewReader(saveBody))
	if err != nil {
		t.Fatalf("POST /api/save failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}
	savedBytes, err := os.ReadFile(filepath.Join(examplesDir, "gear_copy.gsl"))
	if err != nil {
		t.Fatalf("expected gear_copy.gsl on disk: %v", err)
	}
	if string(savedBytes) != "return sphere(12);\n" {
		t.Errorf("unexpected saved file content: %q", string(savedBytes))
	}

	// Invalid filename rejection
	for _, badName := range []string{"../evil.gsl", "/etc/passwd.gsl", "bad.txt", ""} {
		badBody, _ := json.Marshal(map[string]string{
			"name":   badName,
			"source": "return sphere(1);\n",
		})
		badRes, err := http.Post(ts.URL+"/api/save", "application/json", bytes.NewReader(badBody))
		if err != nil {
			t.Fatal(err)
		}
		badRes.Body.Close()
		if badRes.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 for save name=%q, got %d", badName, badRes.StatusCode)
		}
	}
}

func TestProgressStreamCompileAndMesh(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to GET /api/progress?clientId=client-1
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/api/progress?clientId=client-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	progRes, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/progress failed: %v", err)
	}
	defer progRes.Body.Close()

	if ct := progRes.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/x-ndjson") {
		t.Errorf("expected Content-Type application/x-ndjson, got %q", ct)
	}

	eventsCh := make(chan server.ProgressEvent, 64)
	go func() {
		scanner := bufio.NewScanner(progRes.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var ev server.ProgressEvent
			if err := json.Unmarshal([]byte(line), &ev); err == nil {
				eventsCh <- ev
			}
		}
	}()

	// Trigger POST /api/compile for client-1
	compileBody, _ := json.Marshal(map[string]any{
		"id":       "c-1",
		"clientId": "client-1",
		"source":   "r = 5; r.min = 1; r.max = 10; return sphere(r);",
	})
	cRes, err := http.Post(ts.URL+"/api/compile", "application/json", bytes.NewReader(compileBody))
	if err != nil {
		t.Fatalf("POST /api/compile failed: %v", err)
	}
	defer cRes.Body.Close()
	if cRes.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on /api/compile, got %d", cRes.StatusCode)
	}
	var compilePayload map[string]any
	if err := json.NewDecoder(cRes.Body).Decode(&compilePayload); err != nil {
		t.Fatalf("decode compile response failed: %v", err)
	}
	if compilePayload["id"] != "c-1" {
		t.Errorf("expected id 'c-1', got %v", compilePayload["id"])
	}
	if !strings.Contains(compilePayload["wgsl"].(string), "fn sdShape") {
		t.Errorf("expected wgsl in compile response")
	}

	// Verify at least one NDJSON progress event arrived for c-1
	select {
	case ev := <-eventsCh:
		if ev.Type != "progress" || ev.ID != "c-1" {
			t.Errorf("unexpected progress event: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("timed out waiting for NDJSON progress event from /api/compile")
	}

	// Trigger POST /api/compile with syntax error -> still 200 OK with errors array
	errCompileBody, _ := json.Marshal(map[string]any{
		"id":       "c-err",
		"clientId": "client-1",
		"source":   "return badCall(1);",
	})
	errRes, err := http.Post(ts.URL+"/api/compile", "application/json", bytes.NewReader(errCompileBody))
	if err != nil {
		t.Fatal(err)
	}
	defer errRes.Body.Close()
	if errRes.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK on compile error payload, got %d", errRes.StatusCode)
	}
	var errPayload struct {
		ID     string `json:"id"`
		Errors []any  `json:"errors"`
	}
	if err := json.NewDecoder(errRes.Body).Decode(&errPayload); err != nil {
		t.Fatal(err)
	}
	if len(errPayload.Errors) == 0 {
		t.Errorf("expected non-empty errors slice on invalid GSL")
	}

	// Trigger POST /api/mesh for client-1 -> raw binary model/stl response
	meshBody, _ := json.Marshal(map[string]any{
		"id":          "m-1",
		"clientId":    "client-1",
		"source":      "return sphere(5);",
		"subdivision": 2,
		"bbox":        []float32{-8, -8, -8, 8, 8, 8},
		"algo":        "marching-cubes",
		"time":        0,
	})
	mRes, err := http.Post(ts.URL+"/api/mesh", "application/json", bytes.NewReader(meshBody))
	if err != nil {
		t.Fatalf("POST /api/mesh failed: %v", err)
	}
	defer mRes.Body.Close()
	if mRes.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on /api/mesh, got %d", mRes.StatusCode)
	}
	if ct := mRes.Header.Get("Content-Type"); ct != "model/stl" {
		t.Errorf("expected Content-Type 'model/stl', got %q", ct)
	}
	stlBytes, err := io.ReadAll(mRes.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(stlBytes) < 84 {
		t.Fatalf("expected binary STL >= 84 bytes, got %d", len(stlBytes))
	}
	triCount := binary.LittleEndian.Uint32(stlBytes[80:84])
	if len(stlBytes) != 84+int(triCount)*50 {
		t.Errorf("invalid binary STL size %d for %d triangles", len(stlBytes), triCount)
	}
}

func TestServerValidationErrors(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	// Missing clientId on /api/progress
	res, err := http.Get(ts.URL + "/api/progress")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing clientId, got %d", res.StatusCode)
	}

	// Malformed JSON on POST endpoints
	for _, endpoint := range []string{"/api/save", "/api/compile", "/api/mesh"} {
		badRes, err := http.Post(ts.URL+endpoint, "application/json", strings.NewReader("{bad json"))
		if err != nil {
			t.Fatal(err)
		}
		badRes.Body.Close()
		if badRes.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 on malformed JSON for %s, got %d", endpoint, badRes.StatusCode)
		}
	}

	// Invalid GSL source on /api/mesh
	badMeshBody, _ := json.Marshal(map[string]any{
		"id":     "m-err",
		"source": "return unknownMeshFunc(1);",
	})
	mErrRes, err := http.Post(ts.URL+"/api/mesh", "application/json", bytes.NewReader(badMeshBody))
	if err != nil {
		t.Fatal(err)
	}
	mErrRes.Body.Close()
	if mErrRes.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 on /api/mesh error, got %d", mErrRes.StatusCode)
	}
}

func TestRealViewerAndExamplesIntegration(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	realExamplesDir := filepath.Join(repoRoot, "examples")
	realViewerDir := filepath.Join(realExamplesDir, "viewer")
	realStdlibDir := filepath.Join(repoRoot, "stdlib")

	srv := server.New(&config.Config{
		Addr:        "127.0.0.1:0",
		ExamplesDir: realExamplesDir,
		ViewerDir:   realViewerDir,
		SourcePaths: []string{realExamplesDir, realStdlibDir},
	})
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	// 1. Verify key viewer entrypoints and subdirectory modules are served.
	requiredPaths := []string{
		"/",
		"/index.html",
		"/main.js",
		"/style.css",
		"/render/viewer.js",
		"/panels/export-panel.js",
		"/panels/mesh-import.js",
		"/panels/slider-panel.js",
		"/ui/event-manager.js",
		"/ui/error.js",
		"/editor/editor.js",
		"/docs/help-panel.js",
	}
	for _, p := range requiredPaths {
		res, err := http.Get(ts.URL + p)
		if err != nil {
			t.Fatalf("GET %s failed: %v", p, err)
		}
		res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected HTTP 200 for %s, got %d", p, res.StatusCode)
		}
	}

	// 2. Verify all relative ES module imports in examples/viewer/**/*.js resolve on disk.
	relImportRe := regexp.MustCompile(`(?m)^\s*import\s+(?:[^;]*?\s+from\s+)?["'](\.\.?/[^"']+)["']`)
	err = filepath.Walk(realViewerDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			if info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".js") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		dir := filepath.Dir(path)
		matches := relImportRe.FindAllStringSubmatch(string(data), -1)
		for _, m := range matches {
			spec := m[1]
			target := filepath.Join(dir, filepath.FromSlash(spec))
			if _, statErr := os.Stat(target); statErr != nil {
				relFile, _ := filepath.Rel(realViewerDir, path)
				t.Errorf("%s has unresolved relative import %q (resolved to %s)", relFile, spec, target)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking viewer directory failed: %v", err)
	}

	// 3. Verify GET /api/examples lists real examples and POST /api/compile compiles gear.gsl.
	exRes, err := http.Get(ts.URL + "/api/examples")
	if err != nil {
		t.Fatal(err)
	}
	defer exRes.Body.Close()
	var listPayload struct {
		Examples []struct {
			Name string `json:"name"`
		} `json:"examples"`
	}
	if err := json.NewDecoder(exRes.Body).Decode(&listPayload); err != nil {
		t.Fatal(err)
	}
	if len(listPayload.Examples) == 0 {
		t.Fatalf("expected non-empty examples list from %s", realExamplesDir)
	}

	srcRes, err := http.Get(ts.URL + "/api/example?name=gear.gsl")
	if err != nil {
		t.Fatal(err)
	}
	defer srcRes.Body.Close()
	var exPayload struct {
		Name   string `json:"name"`
		Source string `json:"source"`
	}
	if err := json.NewDecoder(srcRes.Body).Decode(&exPayload); err != nil {
		t.Fatal(err)
	}
	if exPayload.Source == "" {
		t.Fatal("expected non-empty gear.gsl source")
	}

	compBody, _ := json.Marshal(map[string]any{
		"id":     "real-gear",
		"source": exPayload.Source,
	})
	compRes, err := http.Post(ts.URL+"/api/compile", "application/json", bytes.NewReader(compBody))
	if err != nil {
		t.Fatal(err)
	}
	defer compRes.Body.Close()
	var compOut struct {
		WGSL   string   `json:"wgsl"`
		Errors []string `json:"errors"`
	}
	if err := json.NewDecoder(compRes.Body).Decode(&compOut); err != nil {
		t.Fatal(err)
	}
	if len(compOut.Errors) > 0 || compOut.WGSL == "" {
		t.Fatalf("expected gear.gsl to compile cleanly, got errors=%v", compOut.Errors)
	}
}
