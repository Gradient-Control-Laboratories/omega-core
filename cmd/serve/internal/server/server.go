package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/Gradient-Control-Laboratories/omega-core/cmd/serve/internal/config"
	"github.com/Gradient-Control-Laboratories/omega-core/program/preamble"
)

// ProgressEvent represents a single NDJSON progress update pushed over GET /api/progress.
type ProgressEvent struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Stage     string `json:"stage"`
	Label     string `json:"label"`
	Completed uint32 `json:"completed"`
	Total     uint32 `json:"total"`
}

// ExampleItem represents a discovered .gsl example file in GET /api/examples.
type ExampleItem struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

// Server hosts the static WebGPU viewer and JSON/NDJSON/STL HTTP API endpoints.
type Server struct {
	// Read-only immutable fields set at construction
	cfg *config.Config
	mux *http.ServeMux

	// Lock for accessing the fields below
	mu      sync.Mutex
	clients map[string]chan ProgressEvent

	// uploads are the meshes the viewer sent (POST /api/asset); it locks itself.
	uploads uploads
}

// New constructs and routes a new Server instance for the given Config.
func New(cfg *config.Config) *Server {
	s := &Server{
		cfg:     cfg,
		mux:     http.NewServeMux(),
		clients: make(map[string]chan ProgressEvent),
	}
	s.registerRoutes()
	return s
}

// Handler returns the root HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /api/progress", s.handleProgress)
	s.mux.HandleFunc("GET /api/examples", s.handleExamples)
	s.mux.HandleFunc("GET /api/example", s.handleExample)
	s.mux.HandleFunc("POST /api/save", s.handleSave)
	s.mux.HandleFunc("POST /api/compile", s.handleCompile)
	s.mux.HandleFunc("POST /api/mesh", s.handleMesh)
	s.mux.HandleFunc("POST /api/render-lit", s.handleRenderLit)
	s.mux.HandleFunc("POST /api/tape", s.handleTape)
	s.mux.HandleFunc("POST /api/export", s.handleExport)
	s.mux.HandleFunc("POST /api/asset", s.handleAsset)

	// The Help panel's GSL reference, generated from the standard library
	// this server compiles; ahead of the static files so it cannot go stale.
	s.mux.HandleFunc("GET /docs/gsl_description.md", serveDoc(func() string { return preamble.GSLDescription }))
	s.mux.HandleFunc("GET /docs/gsl_libraries.md", serveDoc(func() string { return preamble.LibrariesDescriptiveText() }))

	// The same reference for a model, as `gsl skill` prints it: the index,
	// then whatever it names, by name or by search.
	s.mux.HandleFunc("GET /docs/skill", handleSkill)
	s.mux.HandleFunc("GET /docs/skill/{name...}", handleSkill)

	assetsDir := filepath.Join(s.cfg.ExamplesDir, "assets")
	s.mux.Handle("GET /assets/", noCacheMiddleware(http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir)))))
	s.mux.Handle("GET /", noCacheMiddleware(http.FileServer(http.Dir(s.cfg.ViewerDir))))
}

// serveDoc answers with a markdown document text generates.
func serveDoc(text func() string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write([]byte(text()))
	}
}

// skillFooter tells a reader of the served index how to read further.
const skillFooter = `
---

## **Reading Further**

The index above names each package and built-in group. Read one in full at ` + "`/docs/skill/NAME`" + `: a
package by import path or prefix (` + "`/docs/skill/material/bulk/wood`" + `, ` + "`/docs/skill/wood`" + `), a built-in group
(` + "`/docs/skill/builtin/mesh`" + `), a built-in, or a member (` + "`/docs/skill/transform.rotate`" + `, ` + "`/docs/skill/unionRound`" + `).
` + "`/docs/skill?member=NAME`" + ` is the same lookup, and ` + "`/docs/skill?grep=TEXT`" + ` finds functions by what their docs say.
`

// handleSkill serves the GSL author skill: the index at /docs/skill, a name's
// documentation at /docs/skill/NAME or ?member=NAME, and a search at
// ?grep=TEXT. An unknown name or a search that matches nothing is a 404 whose
// body says so.
func handleSkill(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if m := r.URL.Query().Get("member"); m != "" {
		name = m
	}
	var text string
	switch grep := r.URL.Query().Get("grep"); {
	case grep != "":
		text = preamble.Grep(grep)
		if text == "" {
			http.Error(w, fmt.Sprintf("no built-in or member mentions %q", grep), http.StatusNotFound)
			return
		}
	case name != "":
		doc, err := preamble.Lookup(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		text = doc
	default:
		text = preamble.GSLSkillIndex() + skillFooter
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write([]byte(text))
}

func noCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleProgress(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("clientId")
	if clientID == "" {
		writeJSONError(w, http.StatusBadRequest, "missing clientId query parameter")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	ch := make(chan ProgressEvent, 64)
	s.mu.Lock()
	s.clients[clientID] = ch
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		if s.clients[clientID] == ch {
			delete(s.clients, clientID)
		}
		s.mu.Unlock()
	}()

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case ev := <-ch:
			line, err := json.Marshal(ev)
			if err != nil {
				continue
			}
			line = append(line, '\n')
			if _, err := w.Write(line); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (s *Server) emitProgress(clientID, reqID, stage, label string, completed, total uint32) {
	if clientID == "" {
		return
	}
	s.mu.Lock()
	ch := s.clients[clientID]
	s.mu.Unlock()
	if ch == nil {
		return
	}
	ev := ProgressEvent{
		Type:      "progress",
		ID:        reqID,
		Stage:     stage,
		Label:     label,
		Completed: completed,
		Total:     total,
	}
	select {
	case ch <- ev:
	default:
	}
}
