package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gradientcontrol.com/omega/core/backend/mesh"
	"gradientcontrol.com/omega/core/cmd/serve/internal/compiler"
)

func (s *Server) handleExamples(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(s.cfg.ExamplesDir)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to read examples directory: "+err.Error())
		return
	}

	items := make([]ExampleItem, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".gsl") {
			continue
		}
		label := strings.TrimSuffix(name, ".gsl")
		items = append(items, ExampleItem{
			Name:  name,
			Label: label,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"examples": items,
	})
}

func (s *Server) handleExample(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if !isValidGSLFilename(name) {
		writeJSONError(w, http.StatusBadRequest, "invalid example filename")
		return
	}

	fullPath := filepath.Join(s.cfg.ExamplesDir, name)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSONError(w, http.StatusNotFound, "example not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to read example: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"name":   name,
		"source": string(data),
	})
}

func (s *Server) handleSave(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string `json:"name"`
		Source string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	if !isValidGSLFilename(req.Name) {
		writeJSONError(w, http.StatusBadRequest, "invalid .gsl filename")
		return
	}

	dstPath := filepath.Join(s.cfg.ExamplesDir, req.Name)
	tmpFile, err := os.CreateTemp(s.cfg.ExamplesDir, ".serve-save-*.tmp")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create temp file: "+err.Error())
		return
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.WriteString(req.Source); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		writeJSONError(w, http.StatusInternalServerError, "failed to write file: "+err.Error())
		return
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		writeJSONError(w, http.StatusInternalServerError, "failed to close file: "+err.Error())
		return
	}
	if err := os.Rename(tmpPath, dstPath); err != nil {
		_ = os.Remove(tmpPath)
		writeJSONError(w, http.StatusInternalServerError, "failed to save file: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"name": req.Name,
	})
}

func (s *Server) handleCompile(w http.ResponseWriter, r *http.Request) {
	var req compiler.CompileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	req.Meshes = s.meshFinder()
	resp := compiler.Compile(req, s.cfg.SourcePaths, func(stage, label string, completed, total uint32) {
		s.emitProgress(req.ClientID, req.ID, stage, label, completed, total)
	})
	if usesMeshes(req.Source) {
		if len(resp.Errors) > 0 {
			log.Printf("compile: %s", resp.Errors[0].Message)
		} else {
			log.Printf("compile: %d mesh grids, %d placed meshes", len(resp.MeshSdfBindings), meshDeclCount(resp.MeshDecls))
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleMesh(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ID          string    `json:"id"`
		ClientID    string    `json:"clientId"`
		Source      string    `json:"source"`
		Subdivision int       `json:"subdivision"`
		Bbox        []float32 `json:"bbox"`
		Algo        string    `json:"algo"`
		Time        float64   `json:"time"`
		// Format is "stl" (the default: one body's STL, or a zip of <name>.stl for
		// several), "3mf" (one file, each body an object in its own color) or
		// "glb" (one binary glTF, each body a node with its own color).
		Format string `json:"format"`
		// Bodies, when set, is the 1-based returns to export (the Bodies
		// panel's visible ones); without it every return is exported.
		Bodies []int `json:"bodies"`
		// Colors gives a 3MF or GLB body's color by its 1-based return index
		// ("#rrggbb", as the Bodies panel shows it); others take the palette.
		Colors map[int]string `json:"colors"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	format := strings.ToLower(strings.TrimSpace(payload.Format))
	if format != "" && format != "stl" && format != "3mf" && format != "glb" {
		writeJSONError(w, http.StatusBadRequest, "unsupported format "+payload.Format+"; must be stl, 3mf or glb")
		return
	}

	meshReq := compiler.MeshRequest{
		ID:          payload.ID,
		ClientID:    payload.ClientID,
		Source:      payload.Source,
		Subdivision: payload.Subdivision,
		Algo:        payload.Algo,
		Time:        payload.Time,
		Meshes:      s.meshFinder(),
		Bodies:      payload.Bodies,
	}
	if len(payload.Bbox) == 6 {
		meshReq.HasBbox = true
		copy(meshReq.Bbox[:], payload.Bbox)
	}

	bodies, err := compiler.MeshBodies(meshReq, s.cfg.SourcePaths, func(stage, label string, completed, total uint32) {
		s.emitProgress(payload.ClientID, payload.ID, stage, label, completed, total)
	})
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Each top-level return is its own body.
	var data []byte
	contentType, filename := "model/stl", "model.stl"
	switch format {
	case "3mf":
		var buf bytes.Buffer
		if err := mesh.Write3MF(&buf, bodies, payload.Colors); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "writing 3MF: "+err.Error())
			return
		}
		data, contentType, filename = buf.Bytes(), "model/3mf", "model.3mf"
	case "glb":
		var buf bytes.Buffer
		if err := mesh.WriteGLB(&buf, bodies, payload.Colors); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "writing GLB: "+err.Error())
			return
		}
		data, contentType, filename = buf.Bytes(), "model/gltf-binary", "model.glb"
	default:
		packed, zipped, err := mesh.PackSTL(bodies)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "zipping bodies: "+err.Error())
			return
		}
		data = packed
		if zipped {
			contentType, filename = "application/zip", "model.zip"
		}
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func isValidGSLFilename(name string) bool {
	if len(name) <= 4 || !strings.HasSuffix(name, ".gsl") {
		return false
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return false
	}
	return filepath.Base(name) == name
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
