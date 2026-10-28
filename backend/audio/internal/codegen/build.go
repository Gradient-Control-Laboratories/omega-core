package codegen

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// AudioToJs generates AudioToJs code from a GSL program for audio output
func AudioToJs(
	prog *program.Program,
) (string, int, error) {
	if prog == nil || prog.Primary == nil {
		return "", 0, fmt.Errorf("program or primary package is nil")
	}

	var gslSource string
	if len(prog.Primary.Files) > 0 {
		gslSource = prog.Primary.Files[0].Source
	}

	d := &Delegate{
		gslSource:       gslSource,
		complexityScore: 0,
		varMap:          make(map[string]build.VarInfo),
		builtFns:        make(map[string]bool),
	}

	err := build.Build(d, prog, false)
	err = build.ErrWithPkgSrc(prog, err)
	return d.jsSource, d.complexityScore, err
}

func (d *Delegate) NewImpliedVars(bScope *build.Scope) ([]build.VarInfo, error) {

	pVarInfo := &simpleVarInfo{bScope: bScope, name: program.PVarName, typ: jsVec3Type, expr: &jsExpression{
		code:       "gsl_p",
		operand:    nil,
		precedence: build.PrecedenceHighest,
		typ:        jsVec3Type,
	}}
	d.varMap[pVarInfo.Name()] = pVarInfo

	timeVarInfo := &simpleVarInfo{bScope: bScope, name: program.TimeVarName, typ: jsNumType, expr: newExpression("gsl_time")}
	d.varMap[timeVarInfo.Name()] = timeVarInfo

	return []build.VarInfo{pVarInfo, timeVarInfo}, nil
}

// BuildPackage implements [build.Delegate].
func (d *Delegate) BuildPackage(pkg *program.Package) error {
	d.rootBScope = d.GetScope().Parent()
	return nil
}

// BuildPackageReturns implements [build.Delegate].
func (d *Delegate) BuildPackageReturns(ops []*program.Operand, expressions []build.Expression) error {
	// Report an error if expressions does not contain one (mono) or two (stereo) expressions

	if len(expressions) != 1 && len(expressions) != 2 {
		return fmt.Errorf("expected one (mono) or two (stereo) expressions, got %d", len(expressions))
	}

	leftExpr := expressions[0]
	var rightExpr build.Expression
	if len(expressions) == 2 {
		rightExpr = expressions[1]
	} else {
		rightExpr = leftExpr
	}

	var sb strings.Builder

	// Add helper functions
	sb.WriteString(JSRuntimeHelpers())
	sb.WriteString("\n")

	// Generate the function that computes audio
	sb.WriteString("function computeAudio(gsl_time) {\n")

	// Add built-in constants inside the function so user-defined functions can access them
	sb.WriteString("const gsl_p = [0, 0, 0];\n")
	sb.WriteString("const gsl_pi = Math.PI;\n")
	sb.WriteString("const gsl_e = Math.E;\n\n")

	// Add top-level declarations/assignments from every scope (primary + any
	// imports). They were accumulated in d.topDecls as each was visited.
	if src := strings.TrimSpace(d.topDecls.String()); src != "" {
		sb.WriteString(src)
		sb.WriteString("\n")
	}

	// Add user-defined function definitions inside computeAudio so they can access
	// module-level constants like tau
	if fnDefs := strings.TrimSpace(d.fnDefs.String()); fnDefs != "" {
		sb.WriteString(fnDefs)
		sb.WriteString("\n\n")
	}

	// Generate the return statement
	leftCode := getCode(leftExpr)
	rightCode := getCode(rightExpr)
	sb.WriteString(fmt.Sprintf("    return [%s, %s];\n", leftCode, rightCode))

	sb.WriteString("}\n")

	d.jsSource = sb.String()
	return nil
}
