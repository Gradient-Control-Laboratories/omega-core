// Package staticpage writes the viewer as static web pages: the viewer's own web files with host-static.js
// in host.js's place, and beside them each document with the field bundle it compiles to (internal/compile,
// the same one /api/compile answers). The page then draws what `gsl viewer` draws and compiles nothing.
//
// `gsl export` writes it as a folder (Dir) and the viewer's Export panel downloads it as a zip (Zip).
package staticpage

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gradientcontrol.com/omega/core/internal/compile"
)

const (
	// StaticHost is the viewer's host for an exported page, written there as ServedHost.
	StaticHost = "host-static.js"
	ServedHost = "host.js"
	// ModelsDir holds an export's documents, and ModelsIndex lists them.
	ModelsDir   = "models"
	ModelsIndex = "index.json"
)

// Skips are the viewer's files an exported page never asks for, by slash path from the viewer's folder
// (a folder skips all under it): the material catalog's large previews, operator pictures and recoloring
// tables, which only the Materials panel reads, and Manifold, which only meshing does. Neither is offered
// by a page that cannot change or mesh its document. What is left of materials/ draws the Bodies panel's rows.
var Skips = []string{
	"README.md",
	"materials/large",
	"materials/operators",
	"materials/coat-basis.bin",
	"materials/finish-basis.bin",
	"vendor/manifold",
}

// Example is one row of models/index.json: a document's name, its label in the menu, and its file.
type Example struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	File  string `json:"file"`
}

// Index is models/index.json.
type Index struct {
	Examples []Example `json:"examples"`
}

// Model is one document's file: its source and the field bundle the source compiles to.
type Model struct {
	Name   string            `json:"name"`
	Source string            `json:"source"`
	Field  *compile.Response `json:"field"`
}

// Source is a document's text as the page compiles it: without a byte order mark, and trimmed (main.js
// trims what it shows). The field bundle's spans are positions in this text, so it is what gets compiled
// and what gets saved.
func Source(text string) string {
	return strings.TrimSpace(strings.TrimPrefix(text, "\ufeff"))
}

// CheckViewer says why the viewer in dir cannot be exported: only one that has the static host can.
func CheckViewer(dir string) error {
	if info, err := os.Stat(filepath.Join(dir, StaticHost)); err != nil || info.IsDir() {
		return fmt.Errorf("%s has no %s: this viewer cannot be exported", dir, StaticHost)
	}
	return nil
}

// A Sink takes an export's files, each named by its slash path from the export's root, and answers the
// bytes it took.
type Sink interface {
	Put(name string, r io.Reader) (int64, error)
}

// Write puts an export into sink: the viewer's files but for Skips and dotfiles, with the static host in the
// served host's place, then each model and the index that lists them. It returns the bytes written.
func Write(sink Sink, viewerDir string, models []Model) (int64, error) {
	size, err := writeViewer(sink, viewerDir)
	if err != nil {
		return 0, err
	}
	index := Index{Examples: make([]Example, 0, len(models))}
	for _, m := range models {
		label := strings.TrimSuffix(m.Name, ".gsl")
		file := label + ".json"
		n, err := putJSON(sink, path.Join(ModelsDir, file), m)
		if err != nil {
			return 0, err
		}
		size += n
		index.Examples = append(index.Examples, Example{Name: m.Name, Label: label, File: file})
	}
	n, err := putJSON(sink, path.Join(ModelsDir, ModelsIndex), index)
	if err != nil {
		return 0, err
	}
	return size + n, nil
}

func writeViewer(sink Sink, viewerDir string) (int64, error) {
	var size int64
	err := filepath.WalkDir(viewerDir, func(file string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(viewerDir, file)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		slash := filepath.ToSlash(rel)
		if strings.HasPrefix(d.Name(), ".") || skipped(slash) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// Only files are put, so a folder with nothing left to copy is not made.
		if d.IsDir() || !d.Type().IsRegular() || slash == ServedHost {
			return nil
		}
		if slash == StaticHost {
			slash = ServedHost
		}
		in, err := os.Open(file)
		if err != nil {
			return err
		}
		defer in.Close()
		n, err := sink.Put(slash, in)
		size += n
		return err
	})
	return size, err
}

func skipped(slash string) bool {
	for _, skip := range Skips {
		if slash == skip || strings.HasPrefix(slash, skip+"/") {
			return true
		}
	}
	return false
}

// putJSON puts v as JSON, with <, > and & as they are (WGSL is full of them).
func putJSON(sink Sink, name string, v any) (int64, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return 0, err
	}
	return sink.Put(name, &buf)
}

// Dir is a Sink that writes an export as a folder. PrepareDir makes the folder ready first.
type Dir string

// Put writes the file under the folder, making the folders above it.
func (d Dir) Put(name string, r io.Reader) (int64, error) {
	to := filepath.Join(string(d), filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return 0, err
	}
	dst, err := os.Create(to)
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(dst, r)
	if closeErr := dst.Close(); err == nil {
		err = closeErr
	}
	return n, err
}

// PrepareDir makes out ready to write into: a new or empty folder, or an earlier export, whose models are
// removed so the new list replaces them. Any other folder with files in it is refused, so an export never
// lands among someone's files.
func PrepareDir(out string) error {
	info, err := os.Stat(out)
	if errors.Is(err, fs.ErrNotExist) {
		return os.MkdirAll(out, 0o755)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is a file: name a folder to write", out)
	}
	entries, err := os.ReadDir(out)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	indexPath := filepath.Join(out, ModelsDir, ModelsIndex)
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("%s has files in it and is not an earlier export: name a new or empty folder", out)
	}
	var old Index
	if err := json.Unmarshal(data, &old); err != nil {
		return fmt.Errorf("%s: %w", indexPath, err)
	}
	for _, e := range old.Examples {
		// Only a file of the models folder: the index is read back from disk.
		if e.File == "" || e.File != filepath.Base(e.File) {
			continue
		}
		if err := os.Remove(filepath.Join(out, ModelsDir, e.File)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	return nil
}

// Zip is a Sink that adds an export's files to a zip archive, all under the folder Root, so that
// unpacking the archive gives one folder to put on a web server.
type Zip struct {
	W    *zip.Writer
	Root string
}

// Put adds the file to the archive.
func (z Zip) Put(name string, r io.Reader) (int64, error) {
	w, err := z.W.Create(path.Join(z.Root, name))
	if err != nil {
		return 0, err
	}
	return io.Copy(w, r)
}
