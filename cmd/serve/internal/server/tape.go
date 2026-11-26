package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/Gradient-Control-Laboratories/omega-core/backend/tape"
	"github.com/Gradient-Control-Laboratories/omega-core/cmd/serve/internal/compiler"
)

// handleTape is POST /api/tape: the viewer's Fidget draw device asks for its document as an evaluation tape (a
// compiler.TapeRequest), and the Export panel's Omega Tape formats ask for it to save. The reply is the tape's bytes, as
// backend/tape writes them, or with "format": "json" the same tape as JSON (tapeJSON). It reports no progress: the viewer
// times the frame it is for.
func (s *Server) handleTape(w http.ResponseWriter, r *http.Request) {
	var req struct {
		compiler.TapeRequest
		// Format is "" or "binary" (the tape's protobuf bytes) or "json".
		Format string `json:"format"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	if req.Format != "" && req.Format != "binary" && req.Format != "json" {
		writeJSONError(w, http.StatusBadRequest, "unsupported tape format "+req.Format+"; must be binary or json")
		return
	}
	data, err := compiler.Tape(req.TapeRequest, s.cfg.SourcePaths)
	if err != nil {
		log.Printf("tape: %v", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(data) == 0 {
		writeJSONError(w, http.StatusBadRequest, "the document has nothing to draw")
		return
	}
	contentType := "application/octet-stream"
	if req.Format == "json" {
		if data, err = tapeJSON(data); err != nil {
			log.Printf("tape: %v", err)
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		contentType = "application/json"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// tapeJSON writes a tape's bytes as JSON: protojson's field names and values (64-bit integers as strings, bytes as
// base64), indented two spaces. Fields at their zero value are written too, so a node index of 0 reads as 0 to a
// reader that knows nothing of protobuf's defaults. protojson varies its own whitespace from one build of the binary to
// the next, so the output is re-indented for the same tape to give the same bytes.
func tapeJSON(data []byte) ([]byte, error) {
	var t tape.Tape
	if err := proto.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("reading the tape back: %w", err)
	}
	raw, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(&t)
	if err != nil {
		return nil, fmt.Errorf("writing the tape as JSON: %w", err)
	}
	var out bytes.Buffer
	if err := json.Indent(&out, raw, "", "  "); err != nil {
		return nil, fmt.Errorf("writing the tape as JSON: %w", err)
	}
	out.WriteByte('\n')
	return out.Bytes(), nil
}
