package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/Gradient-Control-Laboratories/omega-core/backend/tape"
)

const tapeJSONSource = "a = sphere(5);\nb = box([0, 0, 0], [4, 4, 4]);\nreturn a, b;\n"

func TestTapeJSON_IsTheSameTape(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	_, binary := postTape(t, ts.URL, map[string]any{"source": tapeJSONSource})
	var want tape.Tape
	if err := proto.Unmarshal(binary, &want); err != nil {
		t.Fatalf("the binary reply is not a tape: %v", err)
	}

	res, data := postTape(t, ts.URL, map[string]any{"source": tapeJSONSource, "format": "json"})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d: %s", res.StatusCode, data)
	}
	if got := res.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type %q, want application/json", got)
	}
	if !json.Valid(data) {
		t.Fatalf("the reply is not JSON: %.200s", data)
	}
	var got tape.Tape
	if err := protojson.Unmarshal(data, &got); err != nil {
		t.Fatalf("the reply is not a tape as protojson writes it: %v", err)
	}
	if !proto.Equal(&got, &want) {
		t.Error("the JSON tape differs from the binary tape of the same document")
	}
	if len(got.Outputs) != 2 {
		t.Errorf("the JSON tape has %d outputs, want 2: every return", len(got.Outputs))
	}
}

func TestTapeJSON_SameTapeSameBytes(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	_, first := postTape(t, ts.URL, map[string]any{"source": tapeJSONSource, "format": "json"})
	_, second := postTape(t, ts.URL, map[string]any{"source": tapeJSONSource, "format": "json"})
	if !bytes.Equal(first, second) {
		t.Error("two requests for the same document gave different JSON")
	}
	if !bytes.HasPrefix(first, []byte("{\n  \"")) || !bytes.HasSuffix(first, []byte("}\n")) {
		t.Errorf("the JSON is not indented two spaces and newline-terminated: %.60q ... %q", first, first[len(first)-4:])
	}
}

func TestTapeJSON_BinaryFormatIsTheDefault(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	_, plain := postTape(t, ts.URL, map[string]any{"source": tapeJSONSource})
	res, named := postTape(t, ts.URL, map[string]any{"source": tapeJSONSource, "format": "binary"})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d: %s", res.StatusCode, named)
	}
	if !bytes.Equal(plain, named) {
		t.Error(`"format": "binary" differs from no format`)
	}
}

func TestTapeJSON_UnknownFormatIsRefused(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	res, data := postTape(t, ts.URL, map[string]any{"source": tapeJSONSource, "format": "yaml"})
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d, want 400: %.200s", res.StatusCode, data)
	}
	var e struct{ Error string }
	if err := json.Unmarshal(data, &e); err != nil {
		t.Fatalf("body is not JSON: %.200s", data)
	}
	if !strings.Contains(e.Error, "yaml") {
		t.Errorf("error %q does not name the format", e.Error)
	}
}
