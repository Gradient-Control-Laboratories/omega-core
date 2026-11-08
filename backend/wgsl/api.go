package wgsl

import (
	"gradientcontrol.com/omega/core/program/edit"
)

// ShapeLocations contains all source spans that contribute to a shape
type ShapeLocations struct {
	Spans []edit.Span `json:"spans"` // All source locations that define this shape
}

type Result struct {
	WGSL           string
	ShapeCount     int              // Number of shapes (from module.Returns)
	ShapeNames     []string         // Human-readable name for each shape, index 0 is shape 1
	ShapeLocations []ShapeLocations // Source locations for each shape, index 0 is shape 1
	MeshSdfSlots   []MeshSdfSlot    // Densified shader slots for bvhSdf / vdbSdf references
	SpinodoidSlots []SpinodoidSlot  // Spinodoid table bands the shader reads
	Parts          []Part           // The box of each part the viewer guards whose box is constant
}

// Part is a part the viewer guards (a document's named value that only a min
// or a return reads; property.MarkParts), by its box, [Lo, Hi] in world
// coordinates, Z up. The viewer's Detail slider measures a part's size on
// screen by its box's shortest edge, and can draw the box. A part whose box
// depends on a slider or on time is left out.
type Part struct {
	Lo [3]float64 `json:"lo"`
	Hi [3]float64 `json:"hi"`
}

// MeshSdfSlot maps a densified shader binding slot to the asset index it
// references. Slots are assigned in the order unique bvhSdf(i) / vdbSdf(i)
// indices are first encountered while walking the module (sorted by
// first-occurrence), yielding contiguous 0-based binding numbers even when
// user-assigned indices are sparse (e.g. bvhSdf(3) + bvhSdf(7) → slots 0 and 1).
type MeshSdfSlot struct {
	Slot       int `json:"slot"`
	AssetIndex int `json:"assetIndex"`
}

// SpinodoidSlot is one band of the spinodoidTex texture the shader reads: slot
// k owns four metadata texels in header row zero, and Key is the spinodoid.Call key of the table
// the canvas must place there (the fidget agent's spinodoid bake returns tables
// under the same keys).
type SpinodoidSlot struct {
	Slot int    `json:"slot"`
	Key  string `json:"key"`
}

