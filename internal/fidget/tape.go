//go:build cgo

package fidget

import (
	"gradientcontrol.com/omega/core/internal/fidget/internal/mesh"
	"gradientcontrol.com/omega/core/program"
)

// GenerateTape serializes prog to an omega bytecode evaluation tape.
func GenerateTape(prog *program.Program) ([]byte, []string, error) {
	return mesh.GenerateTape(prog)
}
