package meshplan

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/internal/assembler/deadcode"
	"gradientcontrol.com/omega/core/internal/assembler/textures"
	"gradientcontrol.com/omega/core/program"
)

// SdfSubtree is an analytical leaf of the plan: the SDF expression the slicer
// carved out of the program's returns, over the program's own assignments.
type SdfSubtree struct {
	// ID is the unique correlation key matching MeshPlanNode.SubtreeID in the export plan.
	// Asynchronous meshing responses from Fidget use this key to route resulting STL URLs
	// back to the matching leaf node in the plan tree.
	ID string

	// Return is the leaf's expression. It references the sliced program's
	// assignments by name, so it is bytes where the program is megabytes, and
	// a response naming N leaves carries the program once rather than N
	// times. SubtreeProgram rebuilds the program that evaluates it.
	Return *program.Operand
}

type sdfSubtreeJSON struct {
	ID     string `json:"id"`
	Return []byte `json:"return,omitempty"`
}

// MarshalJSON encodes SdfSubtree with Return serialized as protobuf bytes so it
// can travel over JSON event buses without exposing protobuf oneof interfaces.
func (s SdfSubtree) MarshalJSON() ([]byte, error) {
	var retBytes []byte
	if s.Return != nil {
		var err error
		retBytes, err = proto.Marshal(s.Return)
		if err != nil {
			return nil, fmt.Errorf("marshal subtree %s return: %w", s.ID, err)
		}
	}
	return json.Marshal(sdfSubtreeJSON{
		ID:     s.ID,
		Return: retBytes,
	})
}

// UnmarshalJSON decodes SdfSubtree from its JSON wire representation, restoring
// the protobuf Return operand.
func (s *SdfSubtree) UnmarshalJSON(data []byte) error {
	var wire sdfSubtreeJSON
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	s.ID = wire.ID
	if len(wire.Return) > 0 {
		ret := &program.Operand{}
		if err := proto.Unmarshal(wire.Return, ret); err != nil {
			return fmt.Errorf("decode subtree %s: %w", wire.ID, err)
		}
		s.Return = ret
	} else {
		s.Return = nil
	}
	return nil
}

// LeafProgram is one plan leaf's self-contained program in its protobuf wire form.
type LeafProgram struct {
	ID    string
	Bytes []byte
}

// LeafPrograms builds each subtree's self-contained program over prog,
// serialized as protobuf bytes for Fidget and WGSL requests.
func LeafPrograms(prog *program.Program, subtrees []SdfSubtree) ([]LeafProgram, error) {
	leaves := make([]LeafProgram, 0, len(subtrees))
	for _, sub := range subtrees {
		if sub.Return == nil {
			return nil, fmt.Errorf("subtree %s has no return", sub.ID)
		}
		b, err := proto.Marshal(SubtreeProgram(prog, sub.Return))
		if err != nil {
			return nil, fmt.Errorf("marshal subtree %s: %w", sub.ID, err)
		}
		leaves = append(leaves, LeafProgram{ID: sub.ID, Bytes: b})
	}
	return leaves, nil
}

// SubtreeProgram builds the self-contained program that evaluates the given
// returns over prog: prog's primary package with its mesh-typed and unreachable
// declarations pruned and the returns in place of its own, with reachable
// dependencies and texture slots extracted. One return makes a plan leaf's
// program for meshing; several make the live preview's.
func SubtreeProgram(prog *program.Program, returns ...*program.Operand) *program.Program {
	var extraRoots []*program.Operand
	if len(returns) == 0 {
		extraRoots = previewControlRoots(prog.Primary)
	}
	return subtreeProgram(prog, returns, extraRoots)
}

func subtreeProgram(prog *program.Program, returns, extraRoots []*program.Operand) *program.Program {
	module := prog.Primary
	deadcodeReturns := returns
	if len(extraRoots) > 0 {
		deadcodeReturns = make([]*program.Operand, 0, len(returns)+len(extraRoots))
		deadcodeReturns = append(deadcodeReturns, returns...)
		deadcodeReturns = append(deadcodeReturns, extraRoots...)
	}
	var files []*program.PackageFile
	for i, f := range module.Files {
		var fReturns []*program.Operand
		if i == len(module.Files)-1 {
			fReturns = deadcodeReturns
		}
		files = append(files, &program.PackageFile{
			Path:        f.Path,
			Imports:     f.Imports,
			Assignments: pruneMeshTypedAssignments(f.Assignments),
			FnTypes:     f.FnTypes,
			Returns:     fReturns,
			MeshImports: f.MeshImports,
			StepImports: f.StepImports,
			WarpImports: f.WarpImports,
		})
	}

	out := &program.Program{
		VersionMajor: prog.VersionMajor,
		VersionMinor: prog.VersionMinor,
		Primary: &program.Package{
			VersionMajor: module.VersionMajor,
			VersionMinor: module.VersionMinor,
			Path:         module.Path,
			Files:        files,
			Scene:        program.CloneScene(module.Scene),
		},
	}
	if len(prog.Dependencies) > 0 {
		out.Dependencies = append([]*program.Package(nil), prog.Dependencies...)
	}
	if len(out.Primary.Files) > 0 {
		if pruned, err := deadcode.Remove(out); err == nil {
			out = pruned
		}
		out.Primary.Files[len(out.Primary.Files)-1].Returns = returns
	}
	out.MeshSdfSlots, out.GlobalTextures = textures.Extract(out)
	return out
}
