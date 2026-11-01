package mesh

import (
	"fmt"
	"log"
	gotime "time"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/convert"
	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
	tapepb "gradientcontrol.com/omega/core/backend/internal/tape"
	"gradientcontrol.com/omega/core/program"
)

// StampSpinodoids adds the spinodoid tables a build bound to a tape, as oracle
// payloads. Every writer of a tape needs this — the agent's export and the
// tapegen dev tool both — so it lives here rather than in either.
//
// There is no mesh for this kind: the payload is the field's arguments, which
// is what lets a consumer rebuild the field rather than trust a table. See
// tape.proto's SpinodoidParams.
//
// A field driver is an implicit tree like any other, so it travels as nodes of
// this tape: its trees are appended to t.Nodes and the payload names them by
// index, the way Appearance names its colour and material trees. Appending
// keeps every existing index (outputs, appearance) valid, and the nodes stay
// in topological order.
func StampSpinodoids(t *tapepb.Tape, oracles []convert.SpinodoidOracle) error {
	for _, s := range oracles {
		params := &tapepb.SpinodoidParams{
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
		t.Oracles = append(t.Oracles, &tapepb.OraclePayload{
			Kind:      tapepb.OracleKind_SPINODOID,
			AssetId:   s.AssetID,
			Spinodoid: params,
		})
	}
	return nil
}

// spliceTape appends the nodes of the serialized tape sub to t, shifting every
// node reference by where they land, and returns sub's outputs as indices into
// t.Nodes. sub's subgraphs are not deduplicated against t's; a driver is small.
func spliceTape(t *tapepb.Tape, sub []byte) ([]uint32, error) {
	var s tapepb.Tape
	if err := proto.Unmarshal(sub, &s); err != nil {
		return nil, err
	}
	base := uint32(len(t.Nodes))
	for i, n := range s.Nodes {
		switch v := n.N.(type) {
		case *tapepb.Node_Unary:
			v.Unary.Arg += base
		case *tapepb.Node_Binary:
			v.Binary.Lhs += base
			v.Binary.Rhs += base
		case *tapepb.Node_Remap:
			v.Remap.Target += base
			v.Remap.X += base
			v.Remap.Y += base
			v.Remap.Z += base
		case *tapepb.Node_Axis, *tapepb.Node_Constant, *tapepb.Node_Input, *tapepb.Node_MeshOracle:
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

// GenerateMesh is an agent-independent func for STL meshing.
// `subdivisionLevel` is the fidget octree depth (3..12, clamped by the
// caller); `bbox` is the world-space [xmin, ymin, zmin, xmax, ymax,
// zmax] region to mesh. The Rust side derives the world_to_model
// transform from the bbox directly — there's no separate center/scale.
// timeAt is the animation clock the program is built at, so an export matches
// the frame the viewer is showing; pass 0 for a document's opening state.
func GenerateMesh(
	prog *program.Program,
	subdivisionLevel int, bbox [6]float32,
	algorithm string,
	timeAt float64,
	progressFn func(completed, total uint32, status string),
) ([]byte, error) {

	var stlBytes []byte
	var err error

	// TODO: we need to pass the GSL source to the mesh function for error reporting
	err = convert.ToTreesAtTime(prog, timeAt, func(treeIndices []fidget.TreeIndex, _ []string) error {
		stlBytes, err = GenerateSTL(treeIndices, subdivisionLevel, bbox, algorithm, progressFn)
		return err
	})

	return stlBytes, err
}

// GenerateTapeWithSpinodoids is an agent-independent func that serializes a
// parsed program into an omega tape (tape::Tape protobuf) — the post-lowering fidget-tree DAG
// that omega_tape replays. It reuses convert.ToTrees so the tape is exactly the
// field the mesher would consume; serialization happens inside the callback
// while the trees are live and the fidget lock is held.
// It also returns the spinodoid tables the build bound: their leaf ids are
// per-bake slot numbers, so the tape has to carry the field's arguments for a
// consumer to rebuild it (see convert.SpinodoidOracle).
func GenerateTapeWithSpinodoids(
	prog *program.Program,
) ([]byte, []string, []convert.SpinodoidOracle, error) {

	var tapeBytes []byte
	var outputNames []string
	var spinodoids []convert.SpinodoidOracle
	var err error

	err = convert.ToTapeTrees(prog, func(treeIndices []fidget.TreeIndex, names []string,
		oracles []convert.SpinodoidOracle) error {
		tapeBytes, err = fidget.SerializeTape(treeIndices)
		outputNames = names
		spinodoids = oracles
		return err
	})

	return tapeBytes, outputNames, spinodoids, err
}

// GenerateTapeWithAppearanceAndSpinodoids is GenerateTapeWithSpinodoids with
// each return's colour and material fields in the tape's `appearance`, lowered
// the way the viewer's colour and material properties are
// (convert.ToColorShapes).
func GenerateTapeWithAppearanceAndSpinodoids(
	prog *program.Program,
) ([]byte, []string, []convert.SpinodoidOracle, error) {

	var tapeBytes []byte
	var outputNames []string
	var spinodoids []convert.SpinodoidOracle
	var err error

	err = convert.ToColorShapesWithOracles(prog, 0, func(shapes []convert.ColorShape,
		oracles []convert.SpinodoidOracle) error {
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

// GenerateTape serializes prog to an omega tape. A tape export that has to be
// readable elsewhere wants GenerateTapeWithSpinodoids, which also returns the
// spinodoid tables to stamp beside it.
func GenerateTape(prog *program.Program) ([]byte, []string, error) {
	tape, names, _, err := GenerateTapeWithSpinodoids(prog)
	return tape, names, err
}

// GenerateTapeWithAppearance is GenerateTape with each return's colour and
// material fields in the tape's `appearance`.
func GenerateTapeWithAppearance(prog *program.Program) ([]byte, []string, error) {
	tape, names, _, err := GenerateTapeWithAppearanceAndSpinodoids(prog)
	return tape, names, err
}

func GenerateSTL(
	treeIndices []fidget.TreeIndex,
	subdivisionLevel int,
	bbox [6]float32,
	algorithm string,
	progressFn func(completed, total uint32, status string),
) ([]byte, error) {
	if len(treeIndices) == 0 {
		return nil, fmt.Errorf("no implicit shapes to mesh")
	}

	treeIndex := treeIndices[0]
	for i := 1; i < len(treeIndices); i++ {
		treeIndex = fidget.Min(treeIndex, treeIndices[i])
	}

	canvasIndex := fidget.BuildCanvas(treeIndex)

	var stlBytes []byte
	var err error
	if algorithm == "marching-cubes" {
		marchingFn := func(completed, total uint32) {
			if progressFn != nil {
				progressFn(completed, total, "")
			}
		}
		log.Printf("marching cubes: depth=%d, bbox=%v", subdivisionLevel, bbox)
		stlBytes, err = fidget.MeshStlMarchingCubes(canvasIndex, subdivisionLevel, bbox, marchingFn)
		if err != nil {
			return nil, fmt.Errorf("marching cubes failed to generate stl: %v", err)
		}
	} else {
		// MDC sends phase signals: (phase, 0) where total=0 means phase indicator.
		phaseNames := map[uint32]string{
			1: "Building octree",
			2: "Walking dual",
			3: "Writing STL",
		}
		phaseStart := gotime.Now()
		dualFn := func(completed, total uint32) {
			if total != 0 {
				if progressFn != nil {
					progressFn(completed, total, "")
				}
				return
			}
			elapsed := gotime.Since(phaseStart)
			if name, ok := phaseNames[completed]; ok {
				if completed > 1 {
					// Log previous phase duration
					log.Printf("mdc phase %d done in %s", completed-1, elapsed.Round(gotime.Millisecond))
				}
				if progressFn != nil {
					progressFn(completed, total, fmt.Sprintf("%s...", name))
				}
				phaseStart = gotime.Now()
			} else if completed == 4 {
				log.Printf("mdc phase 3 done in %s", elapsed.Round(gotime.Millisecond))
				if progressFn != nil {
					progressFn(completed, total, "Done")
				}
			}
		}
		stlBytes, err = fidget.MeshStl(canvasIndex, subdivisionLevel, bbox, dualFn)
		if err != nil {
			return nil, fmt.Errorf("manifold dual-contouring failed to generate stl: %v", err)
		}
	}
	return stlBytes, nil
}
