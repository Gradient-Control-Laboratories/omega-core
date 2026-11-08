package meshasset

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// Asset is one imported mesh's server-side state: the parsed triangles
// and their bbox. Distance-field baking is no longer done in Go — the
// fidget agent owns the OpenVDB level set per asset_id, and the canvas
// agent caches sampled dense grids for the WGSL preview path. Instances
// are shared between Registry slots so two session indices that
// uploaded identical bytes point at the same *Asset.
type Asset struct {
	ContentHash        ContentHash
	Filename           string
	Mesh               *TriMesh
	Bytes              []byte
	StepFilename       string
	StepBytes          []byte
	StepSourceFilename string
	StepSourceBytes    []byte
	Error              string
}

// Registry holds the set of imported mesh assets for a single webai/webedit
// session. It maps session-assigned indices to Asset pointers, with a
// secondary content-hash cache so re-imports of identical bytes reuse the
// existing Asset (and its baked artifacts) rather than re-baking.
//
// The registry is intentionally per-session: an index is only meaningful
// relative to the session that assigned it. Nothing persists across restarts.
type Registry struct {
	mu       sync.Mutex
	byIndex  map[int]*Asset
	byHash   map[ContentHash]*Asset
	bySource map[sourceKey]*Asset
	uploads  map[string]*uploadEntry
}

type sourceKey struct {
	groupName string
	filename  string
}

type uploadEntry struct {
	asset *Asset
}

// NewRegistry constructs an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		byIndex:  make(map[int]*Asset),
		byHash:   make(map[ContentHash]*Asset),
		bySource: make(map[sourceKey]*Asset),
		uploads:  make(map[string]*uploadEntry),
	}
}

func (r *Registry) availableIndex(index int) int {
	for index < 0 && r.byIndex[index] != nil {
		index--
	}
	return index
}

// Put binds index → triangles. If another index already references the same
// content hash, the existing Asset (and any bakes) is reused and returned
// with hit=true. Otherwise a fresh Asset is installed and hit=false.
func (r *Registry) Put(index int, filename string, mesh *TriMesh, hash ContentHash, bytes []byte) (asset *Asset, hit bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	index = r.availableIndex(index)
	if existing, ok := r.byHash[hash]; ok {
		r.byIndex[index] = existing
		return existing, true
	}
	a := &Asset{
		ContentHash: hash,
		Filename:    filename,
		Mesh:        mesh,
		Bytes:       append([]byte(nil), bytes...),
	}
	r.byIndex[index] = a
	r.byHash[hash] = a
	return a, false
}

func (r *Registry) PutStep(index int, meshFilename, stepFilename, sourceFilename string, mesh *TriMesh, hash ContentHash, meshBytes, stepBytes, sourceBytes []byte) (asset *Asset, hit bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	index = r.availableIndex(index)
	a := &Asset{
		ContentHash:        hash,
		Filename:           meshFilename,
		Mesh:               mesh,
		Bytes:              append([]byte(nil), meshBytes...),
		StepFilename:       stepFilename,
		StepBytes:          append([]byte(nil), stepBytes...),
		StepSourceFilename: sourceFilename,
		StepSourceBytes:    append([]byte(nil), sourceBytes...),
	}
	r.byIndex[index] = a
	return a, false
}

// PutError binds index -> error state for an asset that failed to load.
func (r *Registry) PutError(index int, filename string, errStr string) *Asset {
	r.mu.Lock()
	defer r.mu.Unlock()
	index = r.availableIndex(index)
	a := &Asset{
		Filename:           filename,
		StepFilename:       filename,
		StepSourceFilename: filename,
		Error:              errStr,
	}
	r.byIndex[index] = a
	return a
}

// Get returns the Asset for the given index, or (nil, false) if absent.
func (r *Registry) Get(index int) (*Asset, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.byIndex[index]
	return a, ok
}

// FindByFilename returns the first Asset whose mesh, STEP payload, or STEP
// source filename matches name. It lets a save resolve an import-expression
// asset (whose compiler-assigned slot the browser does not know) by the
// filename written in the source.
func (r *Registry) FindByFilename(name string) (*Asset, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, a := range r.byIndex {
		if matchFilename(a.Filename, name) || matchFilename(a.StepFilename, name) || matchFilename(a.StepSourceFilename, name) {
			return a, true
		}
	}
	return nil, false
}

// BindSource records the storage location that produced an asset.
func (r *Registry) BindSource(groupName string, asset *Asset, filenames ...string) {
	if groupName == "" || asset == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, filename := range filenames {
		if filename != "" {
			r.bySource[sourceKey{groupName, normalizeFilename(filename)}] = asset
		}
	}
}

// FindBySource returns the asset loaded from one storage location.
func (r *Registry) FindBySource(groupName, filename string) (*Asset, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.bySource[sourceKey{groupName, normalizeFilename(filename)}]
	return a, ok
}

// ForgetSource drops the binding for one storage location, so the next lookup
// misses and the file is loaded again (it was replaced). Slots aliased to the
// old asset keep it until they are bound again.
func (r *Registry) ForgetSource(groupName, filename string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.bySource, sourceKey{groupName, normalizeFilename(filename)})
}

// BindUpload records an asset supplied directly by the browser.
func (r *Registry) BindUpload(asset *Asset, filenames ...string) {
	if asset == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	entry := &uploadEntry{asset: asset}
	for _, filename := range filenames {
		if filename != "" {
			r.uploads[normalizeFilename(filename)] = entry
		}
	}
}

// ClaimUpload binds the latest matching browser upload to one document.
func (r *Registry) ClaimUpload(groupName, filename string) (*Asset, bool) {
	if groupName == "" {
		return nil, false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.uploads[normalizeFilename(filename)]
	if !ok {
		return nil, false
	}
	for name, uploaded := range r.uploads {
		if uploaded == entry {
			delete(r.uploads, name)
		}
	}
	for _, name := range []string{filename, entry.asset.Filename, entry.asset.StepFilename, entry.asset.StepSourceFilename} {
		if name != "" {
			r.bySource[sourceKey{groupName, normalizeFilename(name)}] = entry.asset
		}
	}
	return entry.asset, true
}

func normalizeFilename(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".step" || ext == ".stp" {
		return strings.ToLower(name[:len(name)-len(ext)]) + ".step"
	}
	return strings.ToLower(name)
}

func matchFilename(a, b string) bool {
	if strings.EqualFold(a, b) {
		return true
	}
	extA := strings.ToLower(filepath.Ext(a))
	extB := strings.ToLower(filepath.Ext(b))
	if (extA == ".step" || extA == ".stp") && (extB == ".step" || extB == ".stp") {
		baseA := strings.ToLower(a[:len(a)-len(extA)])
		baseB := strings.ToLower(b[:len(b)-len(extB)])
		return baseA == baseB
	}
	return false
}

// Alias binds slot to an existing Asset (e.g. one uploaded under a different,
// browser-assigned slot) so a compiler-assigned import-expression slot can point
// at it. Multiple slots referencing one Asset is already how content-hash dedup
// works, so this is just an explicit form of the same.
func (r *Registry) Alias(slot int, a *Asset) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byIndex[slot] = a
}

// Remove drops the mapping for index. If no other index still references the
// same content hash, the Asset becomes unreachable and may be GC'd.
func (r *Registry) Remove(index int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.byIndex[index]
	if !ok {
		return
	}
	delete(r.byIndex, index)
	// If this was the last index referencing the hash, drop the hash mapping too.
	stillUsed := false
	for _, other := range r.byIndex {
		if other == a {
			stillUsed = true
			break
		}
	}
	if !stillUsed {
		delete(r.byHash, a.ContentHash)
		for key, sourced := range r.bySource {
			if sourced == a {
				delete(r.bySource, key)
			}
		}
		for name, uploaded := range r.uploads {
			if uploaded.asset == a {
				delete(r.uploads, name)
			}
		}
	}
}

// RemoveAsset drops every index, source and upload binding that points at
// asset, so a replacement registered afterwards is the only thing a lookup by
// filename or storage location can find. Used when a load-error placeholder is
// superseded by a successful re-translation of the same file.
func (r *Registry) RemoveAsset(asset *Asset) {
	if asset == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for index, a := range r.byIndex {
		if a == asset {
			delete(r.byIndex, index)
		}
	}
	if a, ok := r.byHash[asset.ContentHash]; ok && a == asset {
		delete(r.byHash, asset.ContentHash)
	}
	for key, sourced := range r.bySource {
		if sourced == asset {
			delete(r.bySource, key)
		}
	}
	for name, uploaded := range r.uploads {
		if uploaded.asset == asset {
			delete(r.uploads, name)
		}
	}
}

// Require returns the Asset for index, or an error if no asset is bound.
func (r *Registry) Require(index int) (*Asset, error) {
	a, ok := r.Get(index)
	if !ok {
		return nil, fmt.Errorf("no mesh asset at index %d", index)
	}
	return a, nil
}
