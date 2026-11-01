package parser

import (
	"fmt"

	"gradientcontrol.com/omega/core/parser/internal/lib"
	"gradientcontrol.com/omega/core/program"
)

// ParsePackage parses all source files in a package into a program.Package.
func ParsePackage(sources []string) (*program.Package, error) {
	if len(sources) == 0 {
		pkg, _ := lib.ParsePackage(sources)
		return pkg, fmt.Errorf("no sources provided to parse")
	}
	return lib.ParsePackage(sources)
}

var _ program.ParsePackageFunc = ParsePackage
