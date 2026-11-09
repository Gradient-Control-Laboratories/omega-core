package build

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// Helper utilities for working with *program.Type directly.

func CommonType(a, b *program.Type) (*program.Type, error) {
	if a == nil {
		return b, nil
	}
	if b == nil {
		return a, nil
	}
	if a.Equal(b) {
		return a, nil
	}
	if a.IsScalar() && (b.IsList() || b.IsMatrix()) {
		return b, nil
	}
	if (a.IsList() || a.IsMatrix()) && b.IsScalar() {
		return a, nil
	}
	return nil, fmt.Errorf("incompatible types: %s and %s", a.TypeString(), b.TypeString())
}
