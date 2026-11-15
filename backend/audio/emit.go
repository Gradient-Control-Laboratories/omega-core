package audio

import (
	"gradientcontrol.com/omega/core/backend/audio/internal/codegen"
	"gradientcontrol.com/omega/core/program"
)

// AudioToJs translates an assembled GSL program into Web Audio JavaScript.
func AudioToJs(prog *program.Program) (string, int, error) {
	return codegen.AudioToJs(prog)
}
