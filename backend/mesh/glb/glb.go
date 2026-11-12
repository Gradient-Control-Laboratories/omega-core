// Package glb writes binary glTF 2.0 (.glb) files: one node and one mesh per
// body, each with a material in the body's color.
//
// Vertices are written as the meshers make them, in millimetres with Z up, the
// same numbers an STL or 3MF of the same bodies holds. glTF is metres with Y
// up, so the bodies hang under one root node that scales by 0.001 and turns Z
// up to Y up (a -90° turn about X, as the viewer turns its Z-up world). A
// glTF reader then shows the part at its size and the right way up.
//
// The file carries positions and indices only, no normals: glTF readers then
// shade each triangle flat, as an STL shows. A material is its color as
// baseColorFactor (linear, decoded from the "#rrggbb" sRGB the viewer shows),
// not metallic, as the spec's default metallicFactor of 1 would make it.
package glb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
)

const (
	magic       = 0x46546C67 // "glTF"
	version     = 2
	chunkJSON   = 0x4E4F534A // "JSON"
	chunkBIN    = 0x004E4942 // "BIN\0"
	headerBytes = 12
	chunkHeader = 8

	componentFloat = 5126 // FLOAT
	componentUint  = 5125 // UNSIGNED_INT
	targetArray    = 34962
	targetElements = 34963
	modeTriangles  = 4
)

// Body is one glTF node: its name, its color as "#rrggbb" (anything else is
// written as #808080), and its indexed triangles: Positions holds xyz per
// vertex in millimetres, Z up, and Indices three vertex indices per triangle.
type Body struct {
	Name      string
	Color     string
	Positions []float32
	Indices   []uint32
}

// RootRotation is the root node's rotation, as glTF's [x, y, z, w]
// quaternion: -90° about X, which takes Z up to Y up.
var RootRotation = [4]float64{-math.Sqrt2 / 2, 0, 0, math.Sqrt2 / 2}

// RootScale is the root node's scale: millimetres to metres.
const RootScale = 0.001

type document struct {
	Asset       asset        `json:"asset"`
	Scene       int          `json:"scene"`
	Scenes      []scene      `json:"scenes"`
	Nodes       []node       `json:"nodes"`
	Meshes      []mesh       `json:"meshes,omitempty"`
	Materials   []material   `json:"materials,omitempty"`
	Accessors   []accessor   `json:"accessors,omitempty"`
	BufferViews []bufferView `json:"bufferViews,omitempty"`
	Buffers     []buffer     `json:"buffers,omitempty"`
}

type asset struct {
	Version   string `json:"version"`
	Generator string `json:"generator"`
}

type scene struct {
	Nodes []int `json:"nodes"`
}

type node struct {
	Name     string      `json:"name,omitempty"`
	Mesh     *int        `json:"mesh,omitempty"`
	Children []int       `json:"children,omitempty"`
	Rotation *[4]float64 `json:"rotation,omitempty"`
	Scale    *[3]float64 `json:"scale,omitempty"`
}

type mesh struct {
	Name       string      `json:"name,omitempty"`
	Primitives []primitive `json:"primitives"`
}

type primitive struct {
	Attributes map[string]int `json:"attributes"`
	Indices    int            `json:"indices"`
	Material   int            `json:"material"`
	Mode       int            `json:"mode"`
}

type material struct {
	Name string `json:"name,omitempty"`
	PBR  pbr    `json:"pbrMetallicRoughness"`
}

type pbr struct {
	BaseColorFactor [4]float64 `json:"baseColorFactor"`
	MetallicFactor  float64    `json:"metallicFactor"`
}

type accessor struct {
	BufferView    int       `json:"bufferView"`
	ComponentType int       `json:"componentType"`
	Count         int       `json:"count"`
	Type          string    `json:"type"`
	Min           []float64 `json:"min,omitempty"`
	Max           []float64 `json:"max,omitempty"`
}

type bufferView struct {
	Buffer     int `json:"buffer"`
	ByteOffset int `json:"byteOffset"`
	ByteLength int `json:"byteLength"`
	Target     int `json:"target"`
}

type buffer struct {
	ByteLength int `json:"byteLength"`
}

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// LinearColor is "#rrggbb" (sRGB) as glTF's linear baseColorFactor, alpha 1;
// anything else is #808080.
func LinearColor(hex string) [4]float64 {
	if !hexColor.MatchString(hex) {
		hex = "#808080"
	}
	var c [4]float64
	for i := 0; i < 3; i++ {
		v, _ := strconv.ParseUint(hex[1+2*i:3+2*i], 16, 8)
		c[i] = srgbDecode(float64(v) / 255)
	}
	c[3] = 1
	return c
}

// srgbDecode is the sRGB transfer function's inverse, the counterpart of the
// encoding the viewer's palette goes through (threemf.PaletteColor).
func srgbDecode(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// Write writes bodies as one .glb: a root node (RootScale, RootRotation)
// over one child node per body, in order. A body with no triangles is a node
// with no mesh, since a glTF accessor cannot be empty.
func Write(out io.Writer, bodies []Body) error {
	if len(bodies) == 0 {
		return errors.New("nothing to write: no bodies")
	}
	scale := [3]float64{RootScale, RootScale, RootScale}
	rotation := RootRotation
	doc := document{
		Asset:  asset{Version: "2.0", Generator: "Omega"},
		Scenes: []scene{{Nodes: []int{0}}},
		Nodes:  []node{{Name: "omega", Rotation: &rotation, Scale: &scale}},
	}
	var bin bytes.Buffer
	for _, b := range bodies {
		if len(b.Positions)%3 != 0 || len(b.Indices)%3 != 0 {
			return fmt.Errorf("body %s: %d position floats and %d indices are not whole vertices and triangles",
				b.Name, len(b.Positions), len(b.Indices))
		}
		vertices := uint32(len(b.Positions) / 3)
		for _, ix := range b.Indices {
			if ix >= vertices {
				return fmt.Errorf("body %s: index %d is past its %d vertices", b.Name, ix, vertices)
			}
		}
		nodeIndex := len(doc.Nodes)
		doc.Nodes[0].Children = append(doc.Nodes[0].Children, nodeIndex)
		n := node{Name: b.Name}
		if len(b.Indices) > 0 {
			meshIndex := len(doc.Meshes)
			n.Mesh = &meshIndex
			positions := addView(&doc, &bin, float32Bytes(b.Positions), targetArray)
			minXYZ, maxXYZ := bounds(b.Positions)
			doc.Accessors = append(doc.Accessors, accessor{
				BufferView: positions, ComponentType: componentFloat, Count: int(vertices), Type: "VEC3",
				Min: minXYZ, Max: maxXYZ,
			})
			indices := addView(&doc, &bin, uint32Bytes(b.Indices), targetElements)
			doc.Accessors = append(doc.Accessors, accessor{
				BufferView: indices, ComponentType: componentUint, Count: len(b.Indices), Type: "SCALAR",
			})
			doc.Materials = append(doc.Materials, material{
				Name: b.Name,
				PBR:  pbr{BaseColorFactor: LinearColor(b.Color), MetallicFactor: 0},
			})
			doc.Meshes = append(doc.Meshes, mesh{
				Name: b.Name,
				Primitives: []primitive{{
					Attributes: map[string]int{"POSITION": len(doc.Accessors) - 2},
					Indices:    len(doc.Accessors) - 1,
					Material:   len(doc.Materials) - 1,
					Mode:       modeTriangles,
				}},
			})
		}
		doc.Nodes = append(doc.Nodes, n)
	}
	if bin.Len() > 0 {
		doc.Buffers = []buffer{{ByteLength: bin.Len()}}
	}

	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("glb: encoding the glTF JSON: %w", err)
	}
	jsonChunk := pad(jsonBytes, ' ')
	binChunk := pad(bin.Bytes(), 0)
	total := headerBytes + chunkHeader + len(jsonChunk)
	if len(binChunk) > 0 {
		total += chunkHeader + len(binChunk)
	}
	if uint64(total) > math.MaxUint32 {
		return fmt.Errorf("glb: %d bytes is past the format's 4 GiB limit", total)
	}

	var head bytes.Buffer
	for _, v := range []uint32{magic, version, uint32(total), uint32(len(jsonChunk)), chunkJSON} {
		_ = binary.Write(&head, binary.LittleEndian, v)
	}
	if _, err := out.Write(head.Bytes()); err != nil {
		return err
	}
	if _, err := out.Write(jsonChunk); err != nil {
		return err
	}
	if len(binChunk) == 0 {
		return nil
	}
	var binHead [chunkHeader]byte
	binary.LittleEndian.PutUint32(binHead[0:4], uint32(len(binChunk)))
	binary.LittleEndian.PutUint32(binHead[4:8], chunkBIN)
	if _, err := out.Write(binHead[:]); err != nil {
		return err
	}
	_, err = out.Write(binChunk)
	return err
}

// addView appends data to the binary buffer as a buffer view and returns its
// index. Every view holds 4-byte components, so each starts 4-byte aligned.
func addView(doc *document, bin *bytes.Buffer, data []byte, target int) int {
	doc.BufferViews = append(doc.BufferViews, bufferView{
		Buffer: 0, ByteOffset: bin.Len(), ByteLength: len(data), Target: target,
	})
	bin.Write(data)
	return len(doc.BufferViews) - 1
}

// bounds is the per-axis min and max of xyz positions, which glTF requires
// on a POSITION accessor.
func bounds(positions []float32) ([]float64, []float64) {
	lo := []float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	hi := []float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for i, v := range positions {
		k := i % 3
		lo[k] = math.Min(lo[k], float64(v))
		hi[k] = math.Max(hi[k], float64(v))
	}
	return lo, hi
}

func float32Bytes(vs []float32) []byte {
	out := make([]byte, 4*len(vs))
	for i, v := range vs {
		binary.LittleEndian.PutUint32(out[4*i:], math.Float32bits(v))
	}
	return out
}

func uint32Bytes(vs []uint32) []byte {
	out := make([]byte, 4*len(vs))
	for i, v := range vs {
		binary.LittleEndian.PutUint32(out[4*i:], v)
	}
	return out
}

// pad is b padded to a multiple of 4 bytes with fill, as each GLB chunk must
// be: spaces after the JSON, zeros after the binary.
func pad(b []byte, fill byte) []byte {
	for len(b)%4 != 0 {
		b = append(b, fill)
	}
	return b
}
