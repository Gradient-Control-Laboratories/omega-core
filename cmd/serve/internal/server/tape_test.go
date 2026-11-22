package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"gradientcontrol.com/omega/core/backend/tape"
)

func postTape(t *testing.T, url string, req map[string]any) (*http.Response, []byte) {
	t.Helper()
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.Post(url+"/api/tape", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/tape: %v", err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return res, data
}

func TestTape_IsTheDocumentsTape(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	res, data := postTape(t, ts.URL, map[string]any{"id": "tape-1", "source": "a = sphere(5);\nb = box([0, 0, 0], [4, 4, 4]);\nreturn a, b;\n"})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d: %s", res.StatusCode, data)
	}
	if got := res.Header.Get("Content-Type"); got != "application/octet-stream" {
		t.Errorf("Content-Type %q, want application/octet-stream", got)
	}
	var tp tape.Tape
	if err := proto.Unmarshal(data, &tp); err != nil {
		t.Fatalf("the reply is not a tape: %v", err)
	}
	if len(tp.Nodes) == 0 {
		t.Error("the tape has no nodes")
	}
	if len(tp.Outputs) != 2 {
		t.Errorf("the tape has %d outputs, want 2: every return, not only the first", len(tp.Outputs))
	}
}

func TestTape_AppearanceIsOptIn(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	const src = "a = sphere(5);\na.color = [1, 0, 0];\nreturn a;\n"
	_, plain := postTape(t, ts.URL, map[string]any{"source": src})
	res, lit := postTape(t, ts.URL, map[string]any{"source": src, "appearance": true})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d: %s", res.StatusCode, lit)
	}
	if len(lit) <= len(plain) {
		t.Errorf("a tape with appearance is %d bytes, without %d: the colour field is missing", len(lit), len(plain))
	}
}

func TestTape_ErrorsAreJSON(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	cases := []struct {
		name string
		req  map[string]any
		want string
	}{
		{"does not compile", map[string]any{"source": "return y;"}, "undefined identifier: y"},
		{"reads a mesh", map[string]any{"source": "importMesh 0 \"box.stl\";\nreturn bvhSdf(0);\n"}, "imported mesh"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, data := postTape(t, ts.URL, c.req)
			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("status %d, want 400: %.200s", res.StatusCode, data)
			}
			var e struct{ Error string }
			if err := json.Unmarshal(data, &e); err != nil {
				t.Fatalf("body is not JSON: %.200s", data)
			}
			if !strings.Contains(e.Error, c.want) {
				t.Errorf("error %q does not say %q", e.Error, c.want)
			}
		})
	}

	res, err := http.Post(ts.URL+"/api/tape", "application/json", strings.NewReader("{"))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("a body that is not JSON got status %d, want 400", res.StatusCode)
	}
}
