package mesh

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"gradientcontrol.com/omega/core/assembler/variation"
	"gradientcontrol.com/omega/core/backend/mesh/glb"
	"gradientcontrol.com/omega/core/backend/mesh/threemf"
	"gradientcontrol.com/omega/core/program"
)

// WithBodies limits MeshBodies to the given 1-based top-level returns, as the
// viewer's Bodies panel shows them; returns not listed are not meshed. Mesh
// ignores it.
func WithBodies(indices ...int) Option {
	return func(o *options) {
		o.bodies = make(map[int]bool, len(indices))
		for _, i := range indices {
			o.bodies[i] = true
		}
	}
}

var errNothingToExport = errors.New("nothing to export: every body is hidden")

// MeshBodies meshes each top-level return of prog as its own solid, so
// bodies that touch or overlap are not welded into one union. Bodies come
// back in return order, named as the viewer's Bodies panel and the mesh plan
// name them (BodyNames). A program with one return (or none) is meshed whole,
// exactly as Mesh meshes it. Every body is meshed with the same options, so
// over the same box and cell grid. WithBodies narrows the returns meshed; it
// is an error when that leaves none.
func MeshBodies(prog *program.Program, opts ...Option) ([]Body, error) {
	var o options
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	kept := func(index int) bool { return o.bodies == nil || o.bodies[index] }
	returns, names := topLevelReturns(prog)
	if len(returns) <= 1 {
		if !kept(1) {
			return nil, errNothingToExport
		}
		stl, err := Mesh(prog, opts...)
		if err != nil {
			return nil, err
		}
		name := "body-1"
		if len(names) == 1 {
			name = names[0]
		}
		return []Body{{Name: name, Index: 1, STL: stl}}, nil
	}
	var bodies []Body
	for i, ret := range returns {
		if !kept(i + 1) {
			continue
		}
		stl, err := Mesh(variation.Program(prog, ret), opts...)
		if err != nil {
			return nil, fmt.Errorf("body %s: %w", names[i], err)
		}
		bodies = append(bodies, Body{Name: names[i], Index: i + 1, STL: stl})
	}
	if len(bodies) == 0 {
		return nil, errNothingToExport
	}
	return bodies, nil
}

// topLevelReturns is prog's top-level returns and each one's body name: the
// name program.OperandName gives it against the last primary file's text (as
// meshplan.SliceDocument names a plan's bodies), a repeat suffixed -2, -3, ...,
// and "body-N" for a return it cannot name.
func topLevelReturns(prog *program.Program) ([]*program.Operand, []string) {
	if prog == nil || prog.Primary == nil {
		return nil, nil
	}
	files := prog.Primary.Files
	var source string
	if len(files) > 0 {
		source = files[len(files)-1].Source
	}
	var returns []*program.Operand
	for _, f := range files {
		returns = append(returns, f.Returns...)
	}
	names := make([]string, len(returns))
	used := make(map[string]bool)
	for i, ret := range returns {
		base := program.OperandName(ret, source)
		if base == "" {
			names[i] = fmt.Sprintf("body-%d", i+1)
			continue
		}
		name := base
		for n := 2; used[name]; n++ {
			name = fmt.Sprintf("%s-%d", base, n)
		}
		used[name] = true
		names[i] = name
	}
	return returns, names
}

// WriteSTLZip writes bodies as a zip holding <name>.stl for each, in order.
// The names are made safe as file names and one that comes out like an
// earlier one is suffixed -2, -3, ..., as the viewer's browser exports name
// zip entries (examples/viewer/panels/export-panel.js zipEntryNames).
func WriteSTLZip(w io.Writer, bodies []Body) error {
	zw := zip.NewWriter(w)
	names := make([]string, len(bodies))
	for i, b := range bodies {
		names[i] = b.Name
	}
	for i, entry := range zipEntryNames(names) {
		fw, err := zw.Create(entry)
		if err != nil {
			return fmt.Errorf("zip %s: %w", bodies[i].Name, err)
		}
		if _, err := fw.Write(bodies[i].STL); err != nil {
			return fmt.Errorf("zip %s: %w", bodies[i].Name, err)
		}
	}
	return zw.Close()
}

// PackSTL is bodies as one file: the STL itself for one body, a zip of
// <name>.stl (WriteSTLZip) for several. zipped reports which.
func PackSTL(bodies []Body) (data []byte, zipped bool, err error) {
	if len(bodies) == 1 {
		return bodies[0].STL, false, nil
	}
	var buf bytes.Buffer
	if err := WriteSTLZip(&buf, bodies); err != nil {
		return nil, true, err
	}
	return buf.Bytes(), true, nil
}

// Write3MF writes bodies as one 3MF, each an object in the viewer's default
// color for its return index (threemf.PaletteColor), or in colors[Index]
// where colors has a "#rrggbb" for it.
func Write3MF(w io.Writer, bodies []Body, colors map[int]string) error {
	entries := make([]threemf.Body, len(bodies))
	for i, b := range bodies {
		m, err := threemf.WeldSTL(b.STL)
		if err != nil {
			return fmt.Errorf("body %s: %w", b.Name, err)
		}
		color := colors[b.Index]
		if color == "" {
			color = threemf.PaletteColor(b.Index)
		}
		entries[i] = threemf.Body{Name: b.Name, Color: color, Mesh: m}
	}
	return threemf.Write(w, entries)
}

// WriteGLB writes bodies as one binary glTF (.glb), each a node with its own
// mesh and material, colored as Write3MF colors it: colors[Index] where colors
// has a "#rrggbb" for it, the viewer's palette otherwise. Vertices are welded
// as for 3MF and stay in millimetres, Z up, under a root node that turns them
// into glTF's metres, Y up (package glb).
func WriteGLB(w io.Writer, bodies []Body, colors map[int]string) error {
	entries := make([]glb.Body, len(bodies))
	for i, b := range bodies {
		m, err := threemf.WeldSTL(b.STL)
		if err != nil {
			return fmt.Errorf("body %s: %w", b.Name, err)
		}
		color := colors[b.Index]
		if color == "" {
			color = threemf.PaletteColor(b.Index)
		}
		entries[i] = glb.Body{Name: b.Name, Color: color, Positions: m.Positions, Indices: m.Indices}
	}
	return glb.Write(w, entries)
}

func zipEntryNames(names []string) []string {
	out := make([]string, len(names))
	used := make(map[string]bool)
	for i, n := range names {
		base := fileSafeName(n)
		name := base
		for k := 2; used[name]; k++ {
			name = fmt.Sprintf("%s-%d", base, k)
		}
		used[name] = true
		out[i] = name + ".stl"
	}
	return out
}

// fileSafeName is a body's name as a file name. A name can be a return's
// source text, so the characters file systems refuse (a path separator would
// make a folder in the zip) become _ and runs of whitespace, newlines among
// them, one space.
func fileSafeName(name string) string {
	var b strings.Builder
	space := false
	for _, r := range name {
		switch {
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			space = true
			continue
		case strings.ContainsRune(`/\:*?"<>|`, r) || r < 0x20 || r == 0x7f:
			r = '_'
		}
		if space && b.Len() > 0 {
			b.WriteByte(' ')
		}
		space = false
		b.WriteRune(r)
	}
	if b.Len() == 0 {
		return "body"
	}
	return b.String()
}
