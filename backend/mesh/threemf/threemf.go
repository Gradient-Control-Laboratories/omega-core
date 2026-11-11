// Package threemf writes 3MF files: one <object> per body, each one solid
// color. It is the Go port of omega's browser exporter (common/web/threemf.js
// in omega-release), kept to the same file layout.
//
// The file is the 3MF core spec plus the Materials extension's color group.
// Each body is an object whose pid/pindex point at its color in the one
// <m:colorgroup>, so no triangle carries a color of its own. That is what a
// slicer reads as "this object is this filament": Bambu Studio turns an
// object-level color into the object's extruder, whereas colors on the
// triangles go through its "Standard 3MF Import color" dialog and come out as
// surface paint. A slicer that ignores Materials still gets the bodies as
// separate objects. Units are millimetres, the scenes' unit.
//
// Write takes []Body and writes the .3mf bytes. Color is "#rrggbb" (sRGB);
// Mesh is indexed. A binary STL becomes one with WeldSTL.
package threemf

import (
	"archive/zip"
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	coreNS     = "http://schemas.microsoft.com/3dmanufacturing/core/2015/02"
	materialNS = "http://schemas.microsoft.com/3dmanufacturing/material/2015/02"
)

const contentTypes = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
 <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
 <Default Extension="model" ContentType="application/vnd.ms-package.3dmanufacturing-3dmodel+xml"/>
</Types>
`

const rels = `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
 <Relationship Target="/3D/3dmodel.model" Id="rel0" Type="http://schemas.microsoft.com/3dmanufacturing/2013/01/3dmodel"/>
</Relationships>
`

// Mesh is an indexed triangle mesh: Positions holds xyz per vertex, Indices
// three vertex indices per triangle.
type Mesh struct {
	Positions []float32
	Indices   []uint32
}

// Body is one 3MF object: its name, its color as "#rrggbb" (anything else is
// written as #808080), and its mesh.
type Body struct {
	Name  string
	Color string
	Mesh  Mesh
}

// WeldSTL turns a binary STL's triangle soup into an indexed mesh, joining
// corners whose coordinates are bit-for-bit equal, as the meshers write a
// shared vertex. 3MF requires a triangle's three vertices to be distinct, so
// a triangle that welding collapses is dropped. -0 and NaN weld to 0.
func WeldSTL(stl []byte) (Mesh, error) {
	if len(stl) < 84 {
		return Mesh{}, fmt.Errorf("binary STL is %d bytes, shorter than its 84-byte header", len(stl))
	}
	count := binary.LittleEndian.Uint32(stl[80:84])
	if need := 84 + 50*uint64(count); uint64(len(stl)) < need {
		return Mesh{}, fmt.Errorf("binary STL says %d triangles (%d bytes) but is %d bytes", count, need, len(stl))
	}
	positions := []float32{}
	indices := make([]uint32, 0, int(count)*3)
	byKey := make(map[[3]uint32]uint32)
	var corner [3]uint32
	for t := 0; t < int(count); t++ {
		base := 84 + t*50 + 12
		for c := 0; c < 3; c++ {
			var xyz [3]float32
			var key [3]uint32
			for k := 0; k < 3; k++ {
				off := base + c*12 + k*4
				v := math.Float32frombits(binary.LittleEndian.Uint32(stl[off : off+4]))
				if v == 0 || v != v { // -0 and NaN weld to 0
					v = 0
				}
				xyz[k] = v
				key[k] = math.Float32bits(v)
			}
			index, ok := byKey[key]
			if !ok {
				index = uint32(len(positions) / 3)
				byKey[key] = index
				positions = append(positions, xyz[0], xyz[1], xyz[2])
			}
			corner[c] = index
		}
		if corner[0] == corner[1] || corner[1] == corner[2] || corner[0] == corner[2] {
			continue
		}
		indices = append(indices, corner[0], corner[1], corner[2])
	}
	return Mesh{Positions: positions, Indices: indices}, nil
}

// coord is a float32 coordinate in the fewest digits that read back as the
// same float32, never in exponent form.
func coord(v float32) string {
	return strconv.FormatFloat(float64(v), 'f', -1, 32)
}

var xmlEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;")

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func colorAttr(hex string) string {
	if hexColor.MatchString(hex) {
		return strings.ToUpper(hex)
	}
	return "#808080"
}

// writeObject writes one body's <object>.
func writeObject(w *bufio.Writer, id, pindex int, name string, m Mesh) {
	fmt.Fprintf(w, "  <object id=\"%d\" type=\"model\" name=\"%s\" pid=\"1\" pindex=\"%d\">\n   <mesh>\n    <vertices>\n", id, xmlEscaper.Replace(name), pindex)
	for i := 0; i+2 < len(m.Positions); i += 3 {
		fmt.Fprintf(w, "     <vertex x=\"%s\" y=\"%s\" z=\"%s\"/>\n", coord(m.Positions[i]), coord(m.Positions[i+1]), coord(m.Positions[i+2]))
	}
	w.WriteString("    </vertices>\n    <triangles>\n")
	for i := 0; i+2 < len(m.Indices); i += 3 {
		fmt.Fprintf(w, "     <triangle v1=\"%d\" v2=\"%d\" v3=\"%d\"/>\n", m.Indices[i], m.Indices[i+1], m.Indices[i+2])
	}
	w.WriteString("    </triangles>\n   </mesh>\n  </object>\n")
}

// WriteModel writes the 3D/3dmodel.model document.
func WriteModel(out io.Writer, bodies []Body) error {
	w := bufio.NewWriterSize(out, 1<<16)
	w.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	fmt.Fprintf(w, "<model unit=\"millimeter\" xml:lang=\"en-US\" xmlns=\"%s\" xmlns:m=\"%s\">\n", coreNS, materialNS)
	w.WriteString(" <metadata name=\"Application\">Omega</metadata>\n")
	w.WriteString(" <resources>\n  <m:colorgroup id=\"1\">\n")
	for _, b := range bodies {
		fmt.Fprintf(w, "   <m:color color=\"%s\"/>\n", colorAttr(b.Color))
	}
	w.WriteString("  </m:colorgroup>\n")
	for i, b := range bodies {
		writeObject(w, i+2, i, b.Name, b.Mesh)
	}
	w.WriteString(" </resources>\n <build>\n")
	for i := range bodies {
		fmt.Fprintf(w, "  <item objectid=\"%d\"/>\n", i+2)
	}
	w.WriteString(" </build>\n</model>\n")
	return w.Flush()
}

// Write writes bodies as a .3mf package.
func Write(out io.Writer, bodies []Body) error {
	if len(bodies) == 0 {
		return errors.New("nothing to write: no bodies")
	}
	zw := zip.NewWriter(out)
	for _, part := range []struct {
		name  string
		write func(io.Writer) error
	}{
		{"[Content_Types].xml", func(w io.Writer) error { _, err := io.WriteString(w, contentTypes); return err }},
		{"_rels/.rels", func(w io.Writer) error { _, err := io.WriteString(w, rels); return err }},
		{"3D/3dmodel.model", func(w io.Writer) error { return WriteModel(w, bodies) }},
	} {
		fw, err := zw.Create(part.name)
		if err != nil {
			return fmt.Errorf("3mf: creating %s: %w", part.name, err)
		}
		if err := part.write(fw); err != nil {
			return fmt.Errorf("3mf: writing %s: %w", part.name, err)
		}
	}
	return zw.Close()
}
