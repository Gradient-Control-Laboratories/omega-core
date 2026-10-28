package lib

import (
	api "gradientcontrol.com/omega/core/backend/wgsl/internal/lib"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/logic"
	"gradientcontrol.com/omega/core/program"
)

func Render(
	source string,
	prog *program.Program,
) (*api.RenderResult, error) {
	return logic.Render(source, prog)
}
