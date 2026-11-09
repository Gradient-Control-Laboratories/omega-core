package audio

import (
	"gradientcontrol.com/omega/core/backend/audio/internal/lib"
	"gradientcontrol.com/omega/core/program"
)

// AudioToJs translates an assembled GSL program into Web Audio JavaScript.
func AudioToJs(prog *program.Program) (string, int, error) {
	return lib.AudioToJs(prog)
}

// ExtractAudio splits audio return statements from visual geometry return statements.
func ExtractAudio(gslSource string) (audioSource, otherSource string, err error) {
	return lib.ExtractAudio(gslSource)
}
