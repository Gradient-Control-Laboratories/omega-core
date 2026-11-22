package server_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gradientcontrol.com/omega/core/cmd/serve/internal/server"
)

// litBody is a render-lit request for a sphere of radius 5, seen along +Y from 50 units away.
func litBody(t *testing.T, edit func(map[string]any)) []byte {
	t.Helper()
	req := map[string]any{
		"id":          "lit-1",
		"source":      "return sphere(5);",
		"width":       32,
		"height":      24,
		"worldExtent": 40,
		"camera": map[string]any{
			"pos": []float64{0, -50, 0}, "right": []float64{1, 0, 0}, "up": []float64{0, 0, 1}, "fwd": []float64{0, 1, 0},
			"tanFov": []float64{0.4, 0.3}, "ortho": false,
		},
	}
	if edit != nil {
		edit(req)
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func postLit(t *testing.T, url string, body []byte) (*http.Response, []byte) {
	t.Helper()
	res, err := http.Post(url+"/api/render-lit", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/render-lit: %v", err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return res, data
}

func TestRenderLit_FrameIsFiveFloatsAPixel(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	res, data := postLit(t, ts.URL, litBody(t, nil))
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d: %s", res.StatusCode, data)
	}
	if got := res.Header.Get("Content-Type"); got != "application/octet-stream" {
		t.Errorf("Content-Type %q, want application/octet-stream", got)
	}
	if res.Header.Get("X-Frame-Width") != "32" || res.Header.Get("X-Frame-Height") != "24" {
		t.Errorf("frame size headers %q x %q, want 32 x 24", res.Header.Get("X-Frame-Width"), res.Header.Get("X-Frame-Height"))
	}
	if len(data) != 20*32*24 {
		t.Fatalf("%d bytes, want %d (five float32s for each of 32x24 pixels)", len(data), 20*32*24)
	}
	px := func(x, y, k int) float32 {
		return math.Float32frombits(binary.LittleEndian.Uint32(data[4*(5*(y*32+x)+k):]))
	}
	if shape := px(16, 12, 3); shape != 1 {
		t.Errorf("the centre pixel hit return %g, want 1", shape)
	}
	if depth := px(16, 12, 4); depth < 44 || depth > 46 {
		t.Errorf("the centre pixel's depth is %g, want about 45", depth)
	}
	if r, g, b := px(16, 12, 0), px(16, 12, 1), px(16, 12, 2); !(r > 0 && g > 0 && b > 0) {
		t.Errorf("the centre pixel is unlit: %g %g %g", r, g, b)
	}
	if shape := px(0, 0, 3); shape != 0 {
		t.Errorf("the corner pixel hit return %g, want none", shape)
	}
}

func TestRenderLit_ColorModeDraws(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	res, data := postLit(t, ts.URL, litBody(t, func(r map[string]any) { r["mode"] = "color" }))
	if res.StatusCode != http.StatusOK || len(data) != 20*32*24 {
		t.Fatalf("status %d, %d bytes: %.200s", res.StatusCode, len(data), data)
	}
}

func TestRenderLit_ErrorsAreJSON(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	cases := []struct {
		name string
		body []byte
		want string
	}{
		{"not JSON", []byte("{"), "invalid JSON"},
		{"no camera", litBody(t, func(r map[string]any) { delete(r, "camera") }), "needs the view's camera"},
		{"too big", litBody(t, func(r map[string]any) { r["width"], r["height"] = 5000, 5000 }), "out of range"},
		{"unknown mode", litBody(t, func(r map[string]any) { r["mode"] = "glossy" }), `unknown mode "glossy"`},
		{"does not compile", litBody(t, func(r map[string]any) { r["source"] = "return y;" }), "undefined identifier: y"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, data := postLit(t, ts.URL, c.body)
			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("status %d, want 400: %s", res.StatusCode, data)
			}
			if got := res.Header.Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type %q, want application/json", got)
			}
			var e struct{ Error string }
			if err := json.Unmarshal(data, &e); err != nil {
				t.Fatalf("body is not JSON: %s", data)
			}
			if !strings.Contains(e.Error, c.want) {
				t.Errorf("error %q does not say %q", e.Error, c.want)
			}
		})
	}
}

// A frame that is long enough to report reports its share to the client that asked, on the progress stream, as stage
// render in thousandths and never as done (the frame itself says that).
func TestRenderLit_ProgressGoesToTheClientThatAsked(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/api/progress?clientId=client-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	progRes, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/progress: %v", err)
	}
	defer progRes.Body.Close()
	events := make(chan server.ProgressEvent, 64)
	go func() {
		scanner := bufio.NewScanner(progRes.Body)
		for scanner.Scan() {
			var ev server.ProgressEvent
			if json.Unmarshal(scanner.Bytes(), &ev) == nil {
				events <- ev
			}
		}
	}()

	res, data := postLit(t, ts.URL, litBody(t, func(r map[string]any) {
		r["clientId"], r["width"], r["height"] = "client-1", 96, 72
	}))
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d: %s", res.StatusCode, data)
	}

	select {
	case ev := <-events:
		if ev.Type != "progress" || ev.ID != "lit-1" || ev.Stage != "render" || ev.Total != 1000 {
			t.Errorf("unexpected progress event: %+v", ev)
		}
		if ev.Completed >= ev.Total {
			t.Errorf("the render was reported done (%d of %d); the frame says that", ev.Completed, ev.Total)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no progress event arrived")
	}
}
