// Package stdpkg manages internal loading, source hashing, and pre-parsed
// protobuf validation for the embedded GSL standard library.
//
// Why this package is internal and dependency-free:
// Keeping stdpkg under internal/ without importing program or parser preserves
// an acyclic package graph (parser -> program -> stdlib -> internal/stdpkg)
// while keeping all .pb bytes, SHA-256 source digests, and hash-indexed lookup
// tables out of omega-core's public API.
package stdpkg

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sync"

	"gopkg.in/yaml.v3"
)

// HashSize is the byte length of the SHA-256 source digest prefixed onto each
// <ImportPath>/package.pb file:
//
//	Bytes 0..31 : SourcesHash(sources)
//	Bytes 32..N : Deterministic marshaled *program.Package protobuf payload
const HashSize = sha256.Size

// Entry describes one registered standard library package in curated order.
type Entry struct {
	ImportPath string
	Domain     string
}

// PackageData holds the loaded source text and optional pre-parsed protobuf
// bytes for one standard library package.
type PackageData struct {
	ImportPath string
	Domain     string
	Sources    []string
	rawProto   []byte
	validProto []byte
}

// Registry indexes embedded standard library sources and pre-parsed protobufs
// lazily on first access so package init() performs zero file reads or hashing.
type Registry struct {
	// Read-only immutable fields set at construction
	fsys    fs.FS
	entries []Entry

	// Synchronization for lazy one-time loading of the fields below
	once   sync.Once
	pkgs   []PackageData
	byPath map[string]*PackageData
	byHash map[[HashSize]byte]string
}

// NewRegistry creates a lazy standard library registry backed by fsys.
func NewRegistry(fsys fs.FS, entries []Entry) *Registry {
	copied := make([]Entry, len(entries))
	copy(copied, entries)
	return &Registry{
		fsys:    fsys,
		entries: copied,
	}
}

func (r *Registry) init() {
	r.once.Do(func() {
		r.pkgs = make([]PackageData, len(r.entries))
		r.byPath = make(map[string]*PackageData, len(r.entries))
		r.byHash = make(map[[HashSize]byte]string, len(r.entries))

		for i, entry := range r.entries {
			srcs, err := LoadSources(r.fsys, entry.ImportPath)
			if err != nil {
				panic(fmt.Sprintf("stdpkg: loading %q: %v", entry.ImportPath, err))
			}
			hash := SourcesHash(srcs)
			pd := PackageData{
				ImportPath: entry.ImportPath,
				Domain:     entry.Domain,
				Sources:    srcs,
			}

			// <ImportPath>/package.pb is optional so `stdlib/internal/gen` can
			// bootstrap newly added .gsl libraries before their .pb exists.
			// Validating the 32-byte SHA-256 header against the current .gsl
			// sources ensures that if a developer edits a .gsl file locally and
			// runs tests before regenerating package.pb, stale AST bytes are
			// never served.
			if raw, err := fs.ReadFile(r.fsys, path.Join(entry.ImportPath, "package.pb")); err == nil {
				pd.rawProto = raw
				if len(raw) >= HashSize && [HashSize]byte(raw[:HashSize]) == hash {
					pd.validProto = raw[HashSize:]
					r.byHash[hash] = entry.ImportPath
				}
			} else if !errors.Is(err, fs.ErrNotExist) {
				panic(fmt.Sprintf("stdpkg: reading %q/package.pb: %v", entry.ImportPath, err))
			}

			r.pkgs[i] = pd
			r.byPath[entry.ImportPath] = &r.pkgs[i]
		}
	})
}

// All returns all registered standard library packages in curated order.
func (r *Registry) All() []PackageData {
	r.init()
	return r.pkgs
}

// Sources returns the ordered .gsl source strings for importPath.
func (r *Registry) Sources(importPath string) ([]string, bool) {
	r.init()
	pd, ok := r.byPath[importPath]
	if !ok {
		return nil, false
	}
	return pd.Sources, true
}

// Proto returns the validated marshaled *program.Package protobuf bytes (with
// the 32-byte SHA-256 header stripped) if and only if package.pb exists and its
// 32-byte header matches SourcesHash(sources).
func (r *Registry) Proto(importPath string) ([]byte, bool) {
	r.init()
	pd, ok := r.byPath[importPath]
	if !ok || pd.validProto == nil {
		return nil, false
	}
	return pd.validProto, true
}

// RawProto returns the raw contents of <importPath>/package.pb (including the
// 32-byte header) when the file exists, regardless of whether the header hash
// matches current sources, so freshness tests can report exact mismatches.
func (r *Registry) RawProto(importPath string) ([]byte, bool) {
	r.init()
	pd, ok := r.byPath[importPath]
	if !ok || pd.rawProto == nil {
		return nil, false
	}
	return pd.rawProto, true
}

// ImportPathByHash resolves a length-prefixed source hash to its standard
// library import path when that package has a valid, up-to-date package.pb.
func (r *Registry) ImportPathByHash(hash [HashSize]byte) (string, bool) {
	r.init()
	imp, ok := r.byHash[hash]
	return imp, ok
}

var defaultRegistry *Registry

// RegisterFS registers the embedded standard library filesystem and curated
// package list from gradientcontrol.com/omega/core/stdlib at init time without
// performing any eager file reads or hashing.
func RegisterFS(fsys fs.FS, entries []Entry) {
	defaultRegistry = NewRegistry(fsys, entries)
}

func mustDefault() *Registry {
	if defaultRegistry == nil {
		panic("stdpkg: standard library FS not registered (import gradientcontrol.com/omega/core/stdlib)")
	}
	return defaultRegistry
}

// All returns all registered standard library packages in curated order.
func All() []PackageData {
	return mustDefault().All()
}

// Sources returns the ordered .gsl source strings for importPath.
func Sources(importPath string) ([]string, bool) {
	return mustDefault().Sources(importPath)
}

// Proto returns the validated marshaled *program.Package protobuf payload for
// importPath when its 32-byte SHA-256 header matches current .gsl sources.
func Proto(importPath string) ([]byte, bool) {
	return mustDefault().Proto(importPath)
}

// RawProto returns the raw <importPath>/package.pb bytes (including the 32-byte
// header) when present in the embedded filesystem.
func RawProto(importPath string) ([]byte, bool) {
	return mustDefault().RawProto(importPath)
}

// ImportPathByHash resolves a SourcesHash digest to its standard library import
// path when that package has a valid, hash-matching package.pb.
func ImportPathByHash(hash [HashSize]byte) (string, bool) {
	return mustDefault().ImportPathByHash(hash)
}

// SourcesHash computes a SHA-256 digest over sources, prefixing each file's
// contents with its 8-byte little-endian length so distinct multi-file splits
// (e.g. ["ab", "c"] vs. ["a", "bc"]) never share a digest.
func SourcesHash(sources []string) [HashSize]byte {
	h := sha256.New()
	var n [8]byte
	for _, src := range sources {
		binary.LittleEndian.PutUint64(n[:], uint64(len(src)))
		h.Write(n[:])
		h.Write([]byte(src))
	}
	var sum [HashSize]byte
	copy(sum[:], h.Sum(nil))
	return sum
}

// LoadSources loads the ordered .gsl source files for importPath from fsys,
// enforcing the same package layout rule as internal/source/local.go:
//   - Either <importPath>/package.yaml exists and lists all .gsl files in parse
//     order under `sources:`,
//   - Or <importPath> contains a single .gsl file named <dirname>.gsl.
func LoadSources(fsys fs.FS, importPath string) ([]string, error) {
	dirEntries, err := fs.ReadDir(fsys, importPath)
	if err != nil {
		return nil, err
	}

	var gslFiles []string
	hasPkgYaml := false
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "package.yaml" {
			hasPkgYaml = true
		} else if path.Ext(name) == ".gsl" {
			gslFiles = append(gslFiles, name)
		}
	}

	if hasPkgYaml {
		yamlBytes, err := fs.ReadFile(fsys, path.Join(importPath, "package.yaml"))
		if err != nil {
			return nil, err
		}
		var cfg struct {
			Sources []string `yaml:"sources"`
		}
		if err := yaml.Unmarshal(yamlBytes, &cfg); err != nil {
			return nil, fmt.Errorf("parsing package.yaml: %w", err)
		}
		if len(cfg.Sources) == 0 {
			return nil, errors.New("package.yaml lists no sources")
		}
		listed := make(map[string]bool, len(cfg.Sources))
		sources := make([]string, 0, len(cfg.Sources))
		for _, name := range cfg.Sources {
			if listed[name] {
				return nil, fmt.Errorf("package.yaml lists duplicate source %q", name)
			}
			listed[name] = true
			data, err := fs.ReadFile(fsys, path.Join(importPath, name))
			if err != nil {
				return nil, fmt.Errorf("reading %q from package.yaml: %w", name, err)
			}
			sources = append(sources, string(data))
		}
		for _, f := range gslFiles {
			if !listed[f] {
				return nil, fmt.Errorf("unlisted .gsl file %q alongside package.yaml", f)
			}
		}
		return sources, nil
	}

	wantName := path.Base(importPath) + ".gsl"
	if len(gslFiles) != 1 || gslFiles[0] != wantName {
		return nil, fmt.Errorf("expected single source file %q (or package.yaml), found %v", wantName, gslFiles)
	}
	data, err := fs.ReadFile(fsys, path.Join(importPath, wantName))
	if err != nil {
		return nil, err
	}
	return []string{string(data)}, nil
}
