package slider

import (
	"gradientcontrol.com/omega/core/assembler/slider"
	"gradientcontrol.com/omega/core/program"
)

// Extract delegates to the public slider.Extract implementation.
func Extract(source string, pkg *program.Package) []slider.SliderInfo {
	return slider.Extract(source, pkg)
}

// Lifted delegates to the public slider.Lifted implementation.
func Lifted(pkg *program.Package) map[*program.Assignment]bool {
	return slider.Lifted(pkg)
}
