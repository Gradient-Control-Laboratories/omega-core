package scope

import (
	"strings"

	"gradientcontrol.com/omega/core/program"
)

// Scope defines the read-only query and traversal interface for a GSL program scope.
type Scope interface {
	ContainingScope() Scope
	PkgFileScope() *PkgFileScope
	PkgScope() *PkgScope
	ProgScope() *ProgScope

	// NamedElem retrieves a symbol by name from the current or parent scopes.
	NamedElem(name string) NamedElem

	// Assignments returns all assignments defined in the current scope level.
	Assignments() []*program.Assignment

	// Err returns a structured build error anchored at a specific program location.
	Err(loc *program.Location, err error) error

	// GetAgentScope retrieves target-specific backend compilation state.
	GetAgentScope() AgentScope

	// SetAgentScope attaches target-specific backend compilation state.
	SetAgentScope(agentScope AgentScope)
}

// BuilderScope defines the mutable AST-building and parsing interface.
type BuilderScope interface {
	Scope

	AddAssignment(names []string, expr *program.Operand, start, end *program.Location, rng *program.Range) error
	AddPropertyAssignment(varName, propName string, expr *program.Operand, start, end *program.Location) error
	AddAnnotation(kind, symbol string, start, end *program.Location) error
	AddImport(imp *program.Import) error
	AddMeshImport(meshImport *program.MeshImport) error
	AddStepImport(stepImport *program.StepImport) error
	AddWarpImport(warpImport *program.WarpImport) error

	IncExprDepth() error
	DecExprDepth()

	NewFnScope(fnName string, params []string) (BuilderScope, error)
	SetParam(index int, start *program.Location, rng *program.Range)
	// SetParamKind records a parameter's control annotation: "color",
	// "point", "direction", "transform" or "frame", and for a direction the
	// sibling parameter that anchors it, or "". See Parameter.kind and .anchor.
	SetParamKind(index int, kind, anchor string)
	SetReturnOperands(operands []*program.Operand)
}

// NamedElem represents a named symbol (e.g. variable, function, parameter) that has a containing scope.
type NamedElem interface {
	Name() string
	ContainingScope() Scope
}

// AgentScope allows compiler/codegen backends to attach custom, target-specific data to standard scopes.
type AgentScope interface {
	IsAgentScope()
}

// ErrLoc constructs a standard program.Error with line and column information.
func ErrLoc(pkgPath string, pkgFilePath string, loc *program.Location, err error) error {
	lineNum := 0
	column := 0
	if loc != nil {
		lineNum = int(loc.Line)
		column = int(loc.Column)
	}
	return &program.Error{
		PackagePath: pkgPath,
		FilePath:    pkgFilePath,
		LineNum:     lineNum,
		Column:      column,
		Message:     err.Error(),
	}
}

// ResolvePaths calculates the relative package/reference path for referencing elem from refScope.
func ResolvePaths(refScope Scope, elem NamedElem) (string, string) {
	elemPkg := elem.ContainingScope().PkgScope()
	refPkg := refScope.PkgScope()

	if elemPkg != nil && refPkg != nil && elemPkg.Package.Path == refPkg.Package.Path {
		depthRef := fnDepth(refScope)
		depthElem := fnDepth(elem.ContainingScope())
		parentLevels := depthRef - depthElem
		if parentLevels < 0 {
			parentLevels = 0
		}
		refPath := strings.Repeat("../", parentLevels) + elem.Name()
		return "", refPath
	}

	if elemPkg != nil {
		return elemPkg.Package.Path, elem.Name()
	}
	return "", elem.Name()
}

func fnDepth(s Scope) int {
	depth := 0
	curr := s
	for curr != nil {
		if _, ok := curr.(*FnScope); ok {
			depth++
		}
		curr = curr.ContainingScope()
	}
	return depth
}
