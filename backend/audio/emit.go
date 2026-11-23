package audio

import (
	"github.com/Gradient-Control-Laboratories/omega-core/backend/audio/internal/codegen"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// AudioToJs translates an assembled GSL program into Web Audio JavaScript.
func AudioToJs(prog *program.Program) (string, int, error) {
	return codegen.AudioToJs(prog)
}
