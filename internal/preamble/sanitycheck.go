package preamble

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// Sanity check that the descriptions are in sync with the model
func SanityCheck() {

	// Check that all primitive descriptions are in sync with the model

	for _, op := range program.Operation_value {
		_, ok := primitiveDescriptions[program.Operation(op)]
		if !ok {
			// This should never happen
			name := program.Operation_name[int32(op)]
			panic(fmt.Sprintf("operator %d: %s has no description", op, name))
		}
	}

	// Check that all built-in descriptions are in sync with the model

	for _, name := range program.Builtins() {
		_, ok := builtinDescriptions[name]
		if !ok {
			panic(fmt.Sprintf("builtin %s has no description", name))
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
			panic(fmt.Sprintf("builtin %s has no protobuf definition", name))
		}
	}

	// Check that all primitive and built-in reserved words

	for op, desc := range primitiveDescriptions {
		if desc == skipDescription {
			continue
		}
		name := op.Name()
		if !program.IsReservedWord(name) {
			panic(fmt.Sprintf("operator %d: %s is not a reserved word", op, name))
		}
	}
	for name := range builtinDescriptions {
		if !program.IsReservedWord(name) {
			panic(fmt.Sprintf("builtin %s is not a reserved word", name))
		}
	}
}
