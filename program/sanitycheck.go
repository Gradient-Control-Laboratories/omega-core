package program

import (
	"fmt"
)

func SanityCheck() {
	// Ensure all built-in names are reserved words
	for name := range builtinsMap {
		if !IsReservedWord(name) {
			panic(fmt.Sprintf("builtin %q is not a reserved word", name))
		}
	}
}
