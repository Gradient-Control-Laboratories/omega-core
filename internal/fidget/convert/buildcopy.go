package convert

import (
	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
)

// BuildCopy returns the copy of prog that one build folds. build.Build
// constant-folds the program it is handed in place, and a plain build treats
// every slider as a constant, so on a shared program (a cached query, or a
// caller that evaluates and then differentiates) it would bake the sliders
// into every later build: a parameter build would then read r * k as the
// number it folded to. Each build folds its own copy instead.
func BuildCopy(prog *program.Program) *program.Program {
	return proto.Clone(prog).(*program.Program)
}
