package program_test

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"gradientcontrol.com/omega/core/program"
)

func TestOperationSecondClosest(t *testing.T) {

	// Serialization round-trips (one operand: the target field).
	expr := &program.Expression{
		Operation: program.Operation_SECONDCLOSEST,
		Operands: []*program.Operand{
			{
				Start: &program.Location{Line: 1, Column: 1},
				End:   &program.Location{Line: 1, Column: 5},
				T:     &program.Operand_Value{Value: 42.0},
			},
		},
	}

	data, err := proto.Marshal(expr)
	if err != nil {
		t.Fatalf("failed to marshal program.Expression with SECONDCLOSEST: %v", err)
	}

	var decoded program.Expression
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal program.Expression with SECONDCLOSEST: %v", err)
	}

	if decoded.Operation != program.Operation_SECONDCLOSEST {
		t.Fatalf("expected decoded operation SECONDCLOSEST (%d), got %d",
			program.Operation_SECONDCLOSEST, decoded.Operation)
	}
}
