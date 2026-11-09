package codegen

import (
	"gradientcontrol.com/omega/core/program"
)

var jsNumType = program.NewScalarType()
var jsVec3Type = program.NewListType(3, program.NewScalarType())
