package spinodoid

// Baked is a baked Gabor spinodoid table and the ID of its content: tables with
// the same ContentID are identical, whichever call they serve.
type Baked struct {
	Table     *Table
	ContentID string
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
	// Fast marks the lattice field rather than hodge's. Every table omega
	// binds is the lattice, whichever builtin the document called.
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

