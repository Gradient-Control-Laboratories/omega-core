package meshplan

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/assembler/variation"
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
	// times. variation.Program rebuilds the program that evaluates it.
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
		b, err := proto.Marshal(variation.Program(prog, sub.Return))
		if err != nil {
			return nil, fmt.Errorf("marshal subtree %s: %w", sub.ID, err)
		}
		leaves = append(leaves, LeafProgram{ID: sub.ID, Bytes: b})
	}
	return leaves, nil
}

