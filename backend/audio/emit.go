package audio

import (
	"gradientcontrol.com/omega/core/backend/audio/internal/codegen"
	"gradientcontrol.com/omega/core/backend/audio/internal/extract"
	"gradientcontrol.com/omega/core/program"
)

// AudioToJs translates an assembled GSL program into Web Audio JavaScript.
func AudioToJs(prog *program.Program) (string, int, error) {
	return codegen.AudioToJs(prog)
}

// ExtractAudio looks for the `audio(...)` expression in the given gslSource final return statement.
// If found, then it returns one source string with just the audio return with the "audio(" removed
// and a second source string with all of the other returns and no audio return.
// If there are no `audio()` returns, then it returns two empty strings.
// If the return contained only the audio call, then the second string is empty.
//
// An error is reported if the `audio(...)` expression is not the last element in the return statement
// or if the `audio(...)` expression does not contain either 1 or 2 arguments.
//
// The first string returned is the gsl source for generating mono or stereo audio
// while the second string is the gsl source for generating the other return values.
func ExtractAudio(gslSource string) (audioSource, otherSource string, err error) {
	return extract.ExtractAudio(gslSource)
}
