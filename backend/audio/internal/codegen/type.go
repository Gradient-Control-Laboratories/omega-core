package codegen

import (
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

var jsNumType = program.NewScalarType()
var jsVec3Type = program.NewListType(3, program.NewScalarType())
