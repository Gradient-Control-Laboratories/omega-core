package preamble

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

func TestSanityCheck(t *testing.T) {
	// Check that all primitive descriptions are in sync with the model
	for _, op := range program.Operation_value {
		_, ok := primitiveDescriptions[program.Operation(op)]
		if !ok {
			name := program.Operation_name[int32(op)]
			t.Errorf("operator %d: %s has no description", op, name)
		}
	}

	// Check that all built-in descriptions are in sync with the model
	for _, name := range program.Builtins() {
		_, ok := builtinDescriptions[name]
		if !ok {
			t.Errorf("builtin %s has no description", name)
		}
	}
	for name := range builtinDescriptions {
		found := false
		for _, builtin := range program.Builtins() {
			if name == builtin {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("builtin %s has no protobuf definition", name)
		}
	}

	// Check that all primitive and built-in names are reserved words
	for op, desc := range primitiveDescriptions {
		if desc == skipDescription {
			continue
		}
		name := op.Name()
		if !program.IsReservedWord(name) {
			t.Errorf("operator %d: %s is not a reserved word", op, name)
		}
	}
	for name := range builtinDescriptions {
		if !program.IsReservedWord(name) {
			t.Errorf("builtin %s is not a reserved word", name)
		}
	}
}
