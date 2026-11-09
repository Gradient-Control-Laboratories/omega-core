package program

import "testing"

func TestSanityCheck(t *testing.T) {
	// Ensure all built-in names are reserved words
	for name := range builtinsMap {
		if !IsReservedWord(name) {
			t.Errorf("builtin %q is not a reserved word", name)
		}
	}
}
