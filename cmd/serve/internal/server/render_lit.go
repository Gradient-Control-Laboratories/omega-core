package server

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/Gradient-Control-Laboratories/omega-core/cmd/serve/internal/compiler"
)

// litProgressEvery and litProgressStep thin a CPU frame's progress, which is reported at every march step, to what
// a browser can use: no more often than every litProgressEvery, and only once the share has moved by at least
// litProgressStep thousandths.
const (
	litProgressEvery = 250 * time.Millisecond
	litProgressStep  = 10
)

// handleRenderLit is POST /api/render-lit: the viewer's CPU draw device asks for the lit renderer's frame of its
// document from its camera (a compiler.LitRequest). The reply is the frame as raw bytes: width x height pixels,
// row-major from the top left, five little-endian float32s each, the tone-mapped linear colour (three), the return
// hit (numbered from 1, 0 for none) and the depth along the view direction. The frame's size is in the
// X-Frame-Width and X-Frame-Height headers. While it draws, the share done goes to the client named by the request's
// clientId, on the /api/progress stream, as stage "render" in thousandths.
func (s *Server) handleRenderLit(w http.ResponseWriter, r *http.Request) {
	var req compiler.LitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	req.Meshes = s.meshFinder()

	var sentAt time.Time
	var sent uint32
	start := time.Now()
	lin, err := compiler.RenderLit(req, s.cfg.SourcePaths, func(stage, label string, completed, total uint32) {
		// The last share is not sent: the frame itself says the render is done.
		if completed >= total {
			return
		}
		now := time.Now()
		if !sentAt.IsZero() && (now.Sub(sentAt) < litProgressEvery || completed < sent+litProgressStep) {
			return
		}
		sentAt, sent = now, completed
		s.emitProgress(req.ClientID, req.ID, stage, label, completed, total)
	})
	if err != nil {
		log.Printf("render-lit: %v", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("render-lit %s %dx%d in %v", req.Mode, lin.Width, lin.Height, time.Since(start).Round(time.Millisecond))

	n := lin.Width * lin.Height
	frame := make([]byte, 20*n)
	put := func(i, k int, v float32) {
		binary.LittleEndian.PutUint32(frame[4*(5*i+k):], math.Float32bits(v))
	}
	for i := 0; i < n; i++ {
		put(i, 0, lin.RGB[3*i])
		put(i, 1, lin.RGB[3*i+1])
		put(i, 2, lin.RGB[3*i+2])
		put(i, 3, float32(lin.Shape[i]))
		put(i, 4, lin.Depth[i])
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Frame-Width", strconv.Itoa(lin.Width))
	w.Header().Set("X-Frame-Height", strconv.Itoa(lin.Height))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(frame)
}
