package server

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"gradientcontrol.com/omega/core/assembler/meshplan"
	"gradientcontrol.com/omega/core/internal/meshimport"
	"gradientcontrol.com/omega/core/program/meshasset"
)

// maxUploadBytes bounds one uploaded STL.
const maxUploadBytes = 256 << 20

// maxUploads is how many uploaded files the server keeps; the oldest goes
// first.
const maxUploads = 32

// uploads holds the STL files the viewer's Imports panel sent, by filename,
// so a document that imports one (importMesh N "part.stl") finds it before
// any file of that name on disk. They live as long as the server.
type uploads struct {
	mu    sync.Mutex
	files map[string][]byte
	order []string
}

func (u *uploads) put(name string, data []byte) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.files == nil {
		u.files = map[string][]byte{}
	}
	if _, ok := u.files[name]; ok {
		for i, n := range u.order {
			if n == name {
				u.order = append(u.order[:i], u.order[i+1:]...)
				break
			}
		}
	}
	u.files[name] = data
	u.order = append(u.order, name)
	for len(u.order) > maxUploads {
		delete(u.files, u.order[0])
		u.order = u.order[1:]
	}
}

// find is a meshimport.Finder over the uploads: the primary document's
// imports only, since the Imports panel writes the import into it.
func (u *uploads) find(pkgPath, filename string) ([]byte, error) {
	if pkgPath == "" {
		u.mu.Lock()
		data, ok := u.files[filename]
		u.mu.Unlock()
		if ok {
			return data, nil
		}
	}
	return nil, fmt.Errorf("%s was not uploaded: %w", filename, fs.ErrNotExist)
}

// meshFinder finds an importMesh file: uploaded from the viewer first, then
// beside the examples (the viewer's documents), then in examples/assets, then
// under a source path.
func (s *Server) meshFinder() meshimport.Finder {
	dirs := []string{s.cfg.ExamplesDir, filepath.Join(s.cfg.ExamplesDir, "assets")}
	for _, p := range s.cfg.SourcePaths {
		if filepath.Clean(p) != filepath.Clean(s.cfg.ExamplesDir) {
			dirs = append(dirs, p)
		}
	}
	return meshimport.Chain(s.uploads.find, meshimport.Dirs(dirs...))
}

// handleAsset is POST /api/asset?name=part.stl with the STL's bytes as the
// body: the Imports panel sends each mesh it imports, so bvhSdf and vdbSdf can
// read it.
func (s *Server) handleAsset(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if !isValidMeshFilename(name) {
		writeJSONError(w, http.StatusBadRequest, "name must be an .stl filename, with no folder")
		return
	}
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxUploadBytes))
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeJSONError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("an STL may be at most %d MB", maxUploadBytes>>20))
			return
		}
		writeJSONError(w, http.StatusBadRequest, "reading the upload: "+err.Error())
		return
	}
	mesh, _, err := meshasset.ParseSTL(data)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, name+": "+err.Error())
		return
	}
	s.uploads.put(name, data)
	log.Printf("asset: %s uploaded, %d triangles", name, len(mesh.Triangles))
	writeJSON(w, http.StatusOK, map[string]any{"name": name, "triangles": len(mesh.Triangles)})
}

// usesMeshes reports whether a document mentions an imported mesh, so its
// compiles are logged.
func usesMeshes(source string) bool {
	for _, call := range []string{"importMesh", "bvhSdf", "vdbSdf", "mesh("} {
		if strings.Contains(source, call) {
			return true
		}
	}
	return false
}

// meshDeclCount is how many placed meshes a compile response lists.
func meshDeclCount(decls any) int {
	if d, ok := decls.([]meshplan.MeshDecl); ok {
		return len(d)
	}
	return 0
}

// isValidMeshFilename is the parser's rule for an importMesh path: an .stl
// filename beside the document, with no folder and no leading dot.
func isValidMeshFilename(name string) bool {
	if len(name) <= len(".stl") || !strings.HasSuffix(strings.ToLower(name), ".stl") {
		return false
	}
	if strings.ContainsAny(name, `/\`) || strings.HasPrefix(name, ".") {
		return false
	}
	return filepath.Base(name) == name
}
