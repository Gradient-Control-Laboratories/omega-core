package wgsl

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/program/float16"
	"gradientcontrol.com/omega/core/program/meshasset"
	"gradientcontrol.com/omega/core/program/noise/spinodoid"
)

// SpinodoidBinding is the 2D texture the viewer binds as spinodoidTex: Dims
// (width, height) texels of RGBA float32, little-endian, row by row, base64.
// Hash names its content; TexelsB64 is omitted when the browser reports it
// already holds that Hash (CompileRequest.SpinodoidHash).
type SpinodoidBinding struct {
	Dims      [2]int `json:"dims"`
	Hash      string `json:"hash"`
	TexelsB64 string `json:"texelsB64,omitempty"`
}

// PackSpinodoidBinding stacks the baked spinodoid tables into a single RGBA
// float32 2D texture whose row zero stores 4 metadata texels per slot (origin,
// row offset, size, periodicity, and domain bounds) matching the layout read
// by gsl_spinodoid / gsl_spinodoidGrad in internal/codegen/spinodoid.go.
func PackSpinodoidBinding(
	slots []SpinodoidSlot,
	byKey map[string]spinodoid.BakedTable,
	texelsByID map[string][]byte,
	held string,
) (*SpinodoidBinding, error) {
	const texel = 16 // RGBA float32
	rows := 1        // metadata header
	seen := make(map[int]bool, len(slots))
	h := sha256.New()
	fmt.Fprintf(h, "spinodoid-layout-v2|%d", spinodoid.GPUWidth)
	for _, s := range slots {
		if s.Slot < 0 || s.Slot >= spinodoid.GPUWidth/spinodoid.GPUHeaderTexels || seen[s.Slot] {
			return nil, fmt.Errorf("invalid or duplicate spinodoid texture slot %d", s.Slot)
		}
		seen[s.Slot] = true
		t, ok := byKey[s.Key]
		if !ok {
			return nil, fmt.Errorf("fidget baked no spinodoid table for slot %d", s.Slot)
		}
		if t.Rows < 1 || t.Rows > spinodoid.GPUMaxRows-rows {
			return nil, fmt.Errorf("spinodoid tables exceed the %d-row preview texture limit", spinodoid.GPUMaxRows)
		}
		fmt.Fprintf(h, "|%d:%s:%d:%v:%v:%v:%t", s.Slot, t.ContentID, t.Rows, t.Origin, t.Size, t.Bounds, t.Periodic)
		rows += t.Rows
	}
	dims := [2]int{spinodoid.GPUWidth, rows}
	hash := hex.EncodeToString(h.Sum(nil)[:16])
	if held == hash {
		return &SpinodoidBinding{Dims: dims, Hash: hash}, nil
	}
	out := make([]byte, dims[0]*dims[1]*texel)
	rowOffset := 1
	for _, s := range slots {
		t := byKey[s.Key]
		src := texelsByID[t.ContentID]
		if len(src) != spinodoid.GPUWidth*t.Rows*texel {
			return nil, fmt.Errorf("spinodoid table %d has %d bytes, expected %d", s.Slot, len(src), spinodoid.GPUWidth*t.Rows*texel)
		}
		header := [16]float32{
			float32(t.Origin[0]), float32(t.Origin[1]), float32(t.Origin[2]), float32(rowOffset),
			float32(t.Size[0]), float32(t.Size[1]), float32(t.Size[2]), 0,
			float32(t.Bounds[0]), float32(t.Bounds[1]), float32(t.Bounds[2]), 0,
			float32(t.Bounds[3]), float32(t.Bounds[4]), float32(t.Bounds[5]), 0,
		}
		if t.Periodic {
			header[7] = 1
		}
		for i, value := range header {
			offset := s.Slot*spinodoid.GPUHeaderTexels*texel + i*4
			binary.LittleEndian.PutUint32(out[offset:offset+4], math.Float32bits(value))
		}
		copy(out[rowOffset*spinodoid.GPUWidth*texel:], src)
		rowOffset += t.Rows
	}
	return &SpinodoidBinding{
		Dims:      dims,
		Hash:      hash,
		TexelsB64: base64.StdEncoding.EncodeToString(out),
	}, nil
}

// MeshSdfBinding is the per-slot payload attached to canvas/update-shader for
// every bvhSdf(i) / vdbSdf(i) call referenced by the current program. The
// browser uses it to build one Data3DTexture per slot and bind it as the
// matching sdfTex* / sdfBMin* / sdfBSize* params of the generated sdShape.
type MeshSdfBinding struct {
	Slot       int       `json:"slot"`
	AssetIndex int       `json:"assetIndex"`
	Dims       [3]int    `json:"dims"`
	BboxMin    []float64 `json:"bboxMin"`
	BboxSize   []float64 `json:"bboxSize"`
	// BytesB64 is little-endian r16float (IEEE 754 binary16), row-major with x
	// fastest then y then z (matches meshasset.SdfGrid layout).
	BytesB64 string `json:"bytesB64"`
}

// BuildMeshSdfBindings packages pre-baked SDF grids into the slot-ordered list
// the browser expects, encoding each grid's float32 values as r16float (IEEE
// 754 binary16) base64 bytes.
func BuildMeshSdfBindings(slots []MeshSdfSlot, bakes map[int]*meshasset.SdfGrid) []MeshSdfBinding {
	out := make([]MeshSdfBinding, 0, len(slots))
	for _, s := range slots {
		grid := bakes[s.AssetIndex]
		size := grid.Bbox.Size()
		out = append(out, MeshSdfBinding{
			Slot:       s.Slot,
			AssetIndex: s.AssetIndex,
			Dims:       grid.Dims,
			BboxMin:    []float64{grid.Bbox.Min[0], grid.Bbox.Min[1], grid.Bbox.Min[2]},
			BboxSize:   []float64{size[0], size[1], size[2]},
			BytesB64:   float16.EncodeSliceBase64(grid.Values),
		})
	}
	return out
}
