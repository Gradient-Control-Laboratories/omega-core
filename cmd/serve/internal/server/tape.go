package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Gradient-Control-Laboratories/omega-core/cmd/serve/internal/compiler"
)

// handleTape is POST /api/tape: the viewer's Fidget draw device asks for its document as an evaluation tape (a
// compiler.TapeRequest). The reply is the tape's bytes, as backend/tape writes them. It reports no progress: the viewer
// times the frame it is for.
func (s *Server) handleTape(w http.ResponseWriter, r *http.Request) {
	var req compiler.TapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	data, err := compiler.Tape(req, s.cfg.SourcePaths)
	if err != nil {
		log.Printf("tape: %v", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(data) == 0 {
		writeJSONError(w, http.StatusBadRequest, "the document has nothing to draw")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
