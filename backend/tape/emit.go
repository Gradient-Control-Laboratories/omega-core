package tape

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/internal/fidget/convert"
	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/meshasset"
	"gradientcontrol.com/omega/core/program/noise/spinodoid"
	"gradientcontrol.com/omega/core/program/progress"
)

// MeshSdfOracle holds the triangle soup for a bvhSdf(AssetID) leaf.
//
// Deprecated: Use [meshasset.MeshSdfOracle] instead; will be removed in 0.2.0.
type MeshSdfOracle = meshasset.MeshSdfOracle

// VdbOracle holds the triangle soup and narrow-band level-set parameters for a vdbSdf(AssetID) leaf.
//
// Deprecated: Use [meshasset.VdbOracle] instead; will be removed in 0.2.0.
type VdbOracle = meshasset.VdbOracle

// SpinodoidOracle describes one spinodoid table bound during compilation.
//
// Deprecated: Use [spinodoid.Oracle] instead; will be removed in 0.2.0.
type SpinodoidOracle = spinodoid.Oracle

// Option configures bytecode tape compilation.
type Option func(*options)

type options struct {
	bbox        [6]float32
	meshOracles []meshasset.MeshSdfOracle
	vdbOracles  []meshasset.VdbOracle
	appearance  bool
	levelSets   bool
	progressFn  progress.Callback
}

// Progress supplies a progress callback to report tape compilation progress.
func Progress(fn progress.Callback) Option {
	return func(o *options) {
		o.progressFn = fn
	}
}

// WithProgress supplies a progress callback to report tape compilation progress.
func WithProgress(fn progress.Callback) Option {
	return Progress(fn)
}

// Bbox stamps the world-space scene bounding box [xmin, ymin, zmin, xmax, ymax, zmax] into the tape.
func Bbox(bbox [6]float32) Option {
	return func(o *options) {
		o.bbox = bbox
	}
}

// MeshOracles registers BVH mesh SDF oracles before compilation and embeds them in the tape.
func MeshOracles(oracles []meshasset.MeshSdfOracle) Option {
	return func(o *options) {
		o.meshOracles = oracles
	}
}

// VdbOracles registers OpenVDB level-set oracles before compilation and embeds them in the tape.
func VdbOracles(oracles []meshasset.VdbOracle) Option {
	return func(o *options) {
		o.vdbOracles = oracles
	}
}

// WithAppearance configures whether to serialize each return's colour and material fields into the tape.
func WithAppearance(enable bool) Option {
	return func(o *options) {
		o.appearance = enable
	}
}

// LevelSets configures whether to bake and embed each vdbSdf's narrow-band level set into the tape.
func LevelSets(enable bool) Option {
	return func(o *options) {
		o.levelSets = enable
	}
}

// Compile translates an assembled program into a self-contained bytecode evaluation tape.
func Compile(prog *program.Program, opts ...Option) ([]byte, error) {
	if prog == nil {
		return nil, fmt.Errorf("cannot compile nil program into tape")
	}

	var o options
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	if o.progressFn != nil {
		o.progressFn("Compiling bytecode tape", 0, 1)
	}

	if len(o.meshOracles) > 0 || len(o.vdbOracles) > 0 {
		fidget.Lock()
		for _, mo := range o.meshOracles {
			if err := fidget.RegisterMeshOracle(mo.AssetID, mo.Vertices, mo.Triangles); err != nil {
				fidget.Unlock()
				return nil, fmt.Errorf("mesh oracle %d register: %w", mo.AssetID, err)
			}
		}
		for _, vo := range o.vdbOracles {
			if err := fidget.RegisterVdbOracle(vo.AssetID, vo.Vertices, vo.Triangles, vo.VoxelSize, vo.HalfBandVoxels); err != nil {
				fidget.Unlock()
				return nil, fmt.Errorf("vdb oracle %d register: %w", vo.AssetID, err)
			}
		}
		fidget.Unlock()
	}

	generate := GenerateTapeWithSpinodoids
	if o.appearance {
		generate = GenerateTapeWithAppearanceAndSpinodoids
	}
	tapeBytes, outputNames, spinodoids, err := generate(prog)
	if err != nil {
		return nil, fmt.Errorf("tape generation failed: %w", err)
	}
	if tapeBytes == nil {
		return nil, nil
	}

	if stamped, err := StampMetadata(tapeBytes, o.bbox, outputNames, o.meshOracles, o.vdbOracles, spinodoids); err == nil {
		tapeBytes = stamped
	}

	if o.levelSets && len(o.vdbOracles) > 0 {
		withLevelSets, err := EmbedLevelSets(tapeBytes, o.vdbOracles)
		if err != nil {
			return nil, err
		}
		tapeBytes = withLevelSets
	}

	if o.progressFn != nil {
		o.progressFn("Compiling bytecode tape", 1, 1)
	}

	return tapeBytes, nil
}

// GenerateTapeWithSpinodoids serializes a parsed program into an omega tape and returns bound spinodoid oracles.
func GenerateTapeWithSpinodoids(prog *program.Program) ([]byte, []string, []spinodoid.Oracle, error) {
	var tapeBytes []byte
	var outputNames []string
	var spinodoids []spinodoid.Oracle
	var err error

	err = convert.ToTapeTrees(prog, func(treeIndices []fidget.TreeIndex, names []string,
		oracles []spinodoid.Oracle) error {
		tapeBytes, err = fidget.SerializeTape(treeIndices)
		outputNames = names
		spinodoids = oracles
		return err
	})
	return tapeBytes, outputNames, spinodoids, err
}

// GenerateTapeWithAppearanceAndSpinodoids is GenerateTapeWithSpinodoids with each return's colour and material fields.
func GenerateTapeWithAppearanceAndSpinodoids(prog *program.Program) ([]byte, []string, []spinodoid.Oracle, error) {
	var tapeBytes []byte
	var outputNames []string
	var spinodoids []spinodoid.Oracle
	var err error

	err = convert.ToColorShapesWithOracles(prog, 0, func(shapes []convert.ColorShape,
		oracles []spinodoid.Oracle) error {
		spinodoids = oracles
		trees := make([]fidget.TreeIndex, len(shapes))
		appearance := make([]fidget.Appearance, len(shapes))
		for i, s := range shapes {
			trees[i] = s.Tree
			outputNames = append(outputNames, s.Name)
			a := &appearance[i]
			if s.HasColor {
				a.Color = s.Color[:]
			}
			if s.HasMaterial {
				for _, row := range s.Material {
					a.Material = append(a.Material, row[:]...)
				}
				a.Displaced = s.Displaced
			}
		}
		tapeBytes, err = fidget.SerializeTapeWithAppearance(trees, appearance)
		return err
	})
	return tapeBytes, outputNames, spinodoids, err
}

// StampSpinodoids adds the spinodoid tables a build bound to a tape as oracle payloads.
func StampSpinodoids(t *Tape, oracles []spinodoid.Oracle) error {
	for _, s := range oracles {
		params := &SpinodoidParams{
			Fast:              s.Fast,
			Period:            s.Period,
			ThetaRadians:      s.Theta,
			Poles:             s.Poles,
			PacketWidth:       s.PacketWidth,
			Impulses:          s.Impulses,
			Pool:              s.Pool,
			Seed:              s.Seed,
			Bias:              s.Bias,
			LocalNormalize:    s.LocalNormalize,
			Scale:             s.Scale,
			Orientation:       s.Orientation,
			DriverFingerprint: s.DriverFingerprint,
			Periodic:          s.Periodic,
			Domain:            s.Domain,
		}
		if len(s.DriverTape) > 0 {
			roots, err := spliceTape(t, s.DriverTape)
			if err != nil {
				return fmt.Errorf("spinodoid %#x drivers: %w", s.AssetID, err)
			}
			if s.Scale == nil {
				if len(roots) < 1 {
					return fmt.Errorf("spinodoid %#x: driver tape has no scale output", s.AssetID)
				}
				params.ScaleNodes = roots[:1]
				roots = roots[1:]
			}
			if s.Orientation == nil {
				if len(roots) != 9 {
					return fmt.Errorf("spinodoid %#x: driver tape has %d orientation outputs, want 9",
						s.AssetID, len(roots))
				}
				// The trees arrive column-major (element (i, j) is tree 3j+i);
				// the payload is row-major, like `orientation`.
				params.OrientationNodes = make([]uint32, 9)
				for i := range 3 {
					for j := range 3 {
						params.OrientationNodes[3*i+j] = roots[3*j+i]
					}
				}
			}
		}
		t.Oracles = append(t.Oracles, &OraclePayload{
			Kind:      OracleKind_SPINODOID,
			AssetId:   s.AssetID,
			Spinodoid: params,
		})
	}
	return nil
}

// spliceTape appends the nodes of the serialized tape sub to t, shifting every
// node reference by where they land, and returns sub's outputs as indices into
// t.Nodes. sub's subgraphs are not deduplicated against t's; a driver is small.
func spliceTape(t *Tape, sub []byte) ([]uint32, error) {
	var s Tape
	if err := proto.Unmarshal(sub, &s); err != nil {
		return nil, err
	}
	base := uint32(len(t.Nodes))
	for i, n := range s.Nodes {
		switch v := n.N.(type) {
		case *Node_Unary:
			v.Unary.Arg += base
		case *Node_Binary:
			v.Binary.Lhs += base
			v.Binary.Rhs += base
		case *Node_Remap:
			v.Remap.Target += base
			v.Remap.X += base
			v.Remap.Y += base
			v.Remap.Z += base
		case *Node_Axis, *Node_Constant, *Node_Input, *Node_MeshOracle:
			// leaves: nothing to shift
		default:
			return nil, fmt.Errorf("node %d: unhandled node kind %T", i, n.N)
		}
		t.Nodes = append(t.Nodes, n)
	}
	roots := make([]uint32, len(s.Outputs))
	for i, o := range s.Outputs {
		roots[i] = o + base
	}
	return roots, nil
}

// StampMetadata decodes a serialized tape, stamps the scene bbox, output names,
// and embedded mesh/VDB/spinodoid oracle payloads, and re-encodes it.
func StampMetadata(
	tapeBytes []byte,
	bbox [6]float32,
	outputNames []string,
	meshOracles []meshasset.MeshSdfOracle,
	vdbOracles []meshasset.VdbOracle,
	spinodoids []spinodoid.Oracle,
) ([]byte, error) {
	var t Tape
	if err := proto.Unmarshal(tapeBytes, &t); err != nil {
		return nil, err
	}

	if bbox != ([6]float32{}) {
		t.Bbox = []float32{bbox[0], bbox[1], bbox[2], bbox[3], bbox[4], bbox[5]}
	}
	if len(outputNames) > 0 {
		t.OutputNames = outputNames
	}

	for _, p := range meshOracles {
		t.Oracles = append(t.Oracles, &OraclePayload{
			Kind:      OracleKind_MESH_BVH,
			AssetId:   p.AssetID,
			Vertices:  p.Vertices,
			Triangles: p.Triangles,
			Bbox:      vertsBbox(p.Vertices),
		})
	}
	for _, p := range vdbOracles {
		t.Oracles = append(t.Oracles, &OraclePayload{
			Kind:           OracleKind_VDB,
			AssetId:        p.AssetID,
			Vertices:       p.Vertices,
			Triangles:      p.Triangles,
			VoxelSize:      p.VoxelSize,
			HalfBandVoxels: p.HalfBandVoxels,
			Bbox:           vertsBbox(p.Vertices),
		})
	}
	if err := StampSpinodoids(&t, spinodoids); err != nil {
		return nil, err
	}

	return proto.Marshal(&t)
}

// EmbedLevelSets bakes each vdbSdf payload's level set and embeds it in the
// tape's VDB oracle payload for that asset.
func EmbedLevelSets(tapeBytes []byte, vdbOracles []meshasset.VdbOracle) ([]byte, error) {
	var t Tape
	if err := proto.Unmarshal(tapeBytes, &t); err != nil {
		return nil, err
	}
	fidget.Lock()
	defer fidget.Unlock()
	for _, vo := range vdbOracles {
		if err := fidget.RegisterVdbOracle(vo.AssetID, vo.Vertices, vo.Triangles, vo.VoxelSize, vo.HalfBandVoxels); err != nil {
			return nil, fmt.Errorf("vdb oracle %d register: %w", vo.AssetID, err)
		}
	}
	for _, o := range t.Oracles {
		if o.Kind != OracleKind_VDB {
			continue
		}
		raw, err := fidget.ExtractVdbLevelSet(o.AssetId)
		if err != nil {
			return nil, fmt.Errorf("vdbSdf(%d): %w", o.AssetId, err)
		}
		var ls VdbLevelSet
		if err := proto.Unmarshal(raw, &ls); err != nil {
			return nil, fmt.Errorf("vdbSdf(%d) level set decode: %w", o.AssetId, err)
		}
		o.LevelSet = &ls
	}
	return proto.Marshal(&t)
}

func vertsBbox(verts []float32) []float32 {
	if len(verts) < 3 {
		return nil
	}
	lo := [3]float32{verts[0], verts[1], verts[2]}
	hi := lo
	for i := 0; i+2 < len(verts); i += 3 {
		for a := 0; a < 3; a++ {
			v := verts[i+a]
			if v < lo[a] {
				lo[a] = v
			}
			if v > hi[a] {
				hi[a] = v
			}
		}
	}
	return []float32{lo[0], lo[1], lo[2], hi[0], hi[1], hi[2]}
}
