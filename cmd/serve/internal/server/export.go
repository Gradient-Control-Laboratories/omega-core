package server

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/cmd/serve/internal/staticpage"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/compile"
)

// exportNoteHeader carries what the page should tell the user about an export that succeeded.
const exportNoteHeader = "X-Export-Note"

// handleExport answers POST /api/export with the document as `gsl export` writes it, zipped: the viewer's
// web files and the document's source and field under one folder, <name>-html, to unpack onto any web
// server. The document is the one sent, so the page exports what its editor shows, saved or not.
func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name   string `json:"name"`
		Source string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	name := payload.Name
	if name == "" {
		name = "model.gsl"
	}
	if !isValidGSLFilename(name) {
		writeJSONError(w, http.StatusBadRequest, "invalid .gsl filename")
		return
	}
	source := staticpage.Source(payload.Source)
	if source == "" {
		writeJSONError(w, http.StatusBadRequest, "the document is empty")
		return
	}
	if err := staticpage.CheckViewer(s.cfg.ViewerDir); err != nil {
		writeJSONError(w, http.StatusNotImplemented, err.Error())
		return
	}

	field := compile.Process(compile.Request{ID: name, Source: source, Meshes: s.meshFinder()}, s.cfg.SourcePaths, nil)
	if len(field.Errors) > 0 {
		e := field.Errors[0]
		writeJSONError(w, http.StatusUnprocessableEntity,
			fmt.Sprintf("the document does not compile (%d:%d: %s)", e.LineNum, e.Column, e.Message))
		return
	}
	if field.WGSL == "" {
		writeJSONError(w, http.StatusUnprocessableEntity, "the document returns no shape to draw")
		return
	}

	// The archive is built whole before any of it is sent, so a failure is still an error reply.
	root := strings.TrimSuffix(name, ".gsl") + "-html"
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	models := []staticpage.Model{{Name: name, Source: source, Field: field}}
	if _, err := staticpage.Write(staticpage.Zip{W: zw, Root: root}, s.cfg.ViewerDir, models); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to write the export: "+err.Error())
		return
	}
	if err := zw.Close(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to write the export: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": root + ".zip"}))
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	if field.MeshDecls != nil {
		w.Header().Set(exportNoteHeader, "It places meshes with mesh(), which the exported page does not carry: it draws the rest.")
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
