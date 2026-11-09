package eval

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
)

// Evaluator gives Fold's verdict on single operands of one program, for a
// pass that needs it before the next fold runs. The inliner is one: it
// copies a constant argument to every place the callee reads it, and the fold
// that would reduce each copy to its number only runs after the inlining
// pass is done.
//
// It reads the program as it stands when asked. Values it has worked out for
// top-level declarations are kept for later questions, which holds for
// passes, like inlining, that rewrite operands without changing their values.
type Evaluator struct {
	progScope *scope.ProgScope
	getImport program.GetImportFunc
	checker   ParameterChecker
	cache     map[*program.Operand]value
}

// NewEvaluator returns an Evaluator for prog, whose scope is progScope.
// checker marks the runtime parameters, which are never constant.
func NewEvaluator(prog *program.Program, progScope *scope.ProgScope, checker ParameterChecker) *Evaluator {
	return &Evaluator{
		progScope: progScope,
		getImport: func(path string) (*program.Package, error) {
			if prog.Primary != nil && prog.Primary.Path == path {
				return prog.Primary, nil
			}
			for _, dep := range prog.Dependencies {
				if dep.Path == path {
					return dep, nil
				}
			}
			return nil, fmt.Errorf("package %q not found in program dependencies", path)
		},
		checker: checker,
		cache:   make(map[*program.Operand]value),
	}
}

// Scalar reports op's value when it is one constant number, with op read in
// the package whose scope is pkgScope. An operand that depends on a runtime
// parameter, reads p or time, or that the fold cannot evaluate is not. The
// whole of op is evaluated at once, called bodies included, so the checker
// has to report each parameter's literal, where every evaluation of it ends,
// and not only references to it.
func (e *Evaluator) Scalar(op *program.Operand, pkgScope *scope.PkgScope) (float64, bool) {
	if pkgScope == nil {
		return 0, false
	}
	ctx := &constCtx{
		progScope:  e.progScope,
		cur:        pkgScope,
		getImport:  e.getImport,
		evaluating: make(map[evalKey]bool),
		calls:      make(map[callKey]bool),
		cache:      e.cache,
		checker:    e.checker,
	}
	val, err := ctx.operand(op)
	if err != nil {
		return 0, false
	}
	floats := val.floats()
	if len(floats) != 1 {
		return 0, false
	}
	return floats[0], true
}
