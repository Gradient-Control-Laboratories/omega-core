package spinodoid

import (
	"encoding/base64"
	"encoding/binary"
	"math"
)

// Baked is a baked Gabor spinodoid table and the ID of its content: tables with
// the same ContentID are identical, whichever call they serve. Exactly one of
// Fast and Reference is set: the table of the spinodoid the call named.
type Baked struct {
	Fast      *FastGaborTable
	Reference *ReferenceGaborTable
	ContentID string
}

// GPUTable is the table in the GPU layout a preview binds, under key: the
// fast table's sites or the reference table's kernels (each table's
// GPUTexels), with its layout. The shader reads the kind from the call, so the
// two pack the same way.
func (b Baked) GPUTable(key string) BakedTable {
	var texels []float32
	var layout Layout
	var rows int
	if b.Reference != nil {
		texels, layout, rows = b.Reference.GPUTexels(), b.Reference.Layout, b.Reference.Rows()
	} else {
		texels, layout, rows = b.Fast.GPUTexels(), b.Fast.Layout, b.Fast.Rows()
	}
	raw := make([]byte, 4*len(texels))
	for i, v := range texels {
		binary.LittleEndian.PutUint32(raw[4*i:], math.Float32bits(v))
	}
	return BakedTable{
		Key:       key,
		Rows:      rows,
		Origin:    layout.Origin,
		Size:      layout.Size,
		Bounds:    layout.Bounds,
		Periodic:  layout.Periodic,
		ContentID: b.ContentID,
		TexelsB64: base64.StdEncoding.EncodeToString(raw),
	}
}

// Oracle describes one spinodoid table bound during a build: the leaf ID of its
// slot's distance component, and the arguments the field is a function of. A
// tape export stamps these into tape.oracles (tape.proto's SpinodoidParams) so
// a consumer can rebuild the field — the slot number alone means nothing in
// another process.
type Oracle struct {
	// AssetID is the leaf ID of component 0 (distance). The slot's other three
	// components are the next three IDs; a consumer registers the slot once and
	// resolves all four, so a gradient-only call is covered too.
	AssetID uint32
	// Fast is true for the fast spinodoid (fastGaborSpinodoid), false for the
	// reference (gaborSpinodoid): the oracle the document called.
	Fast bool

	Period         float64
	Theta          []float64 // radians
	Poles          []float64 // flat xyz, 3 per theta
	PacketWidth    float64
	Impulses       uint32
	Pool           uint32
	Seed           int64
	Bias           float64 // the level itself, in field standard deviations
	LocalNormalize bool

	// Constant drivers, nil when the driver is a field. Orientation is
	// row-major.
	Scale       []float64
	Orientation []float64
	// DriverFingerprint identifies the field drivers omega used, empty when
	// both are constants.
	DriverFingerprint string
	// DriverTape is the field drivers serialized as a tape of their own: one
	// output for the scale if Scale is nil, then nine for the orientation if
	// Orientation is nil, column-major (the order a GSL matrix's components
	// arrive in).
	DriverTape []byte

	Periodic bool
	Domain   []float64 // 6 entries when bounded, nil otherwise
}
