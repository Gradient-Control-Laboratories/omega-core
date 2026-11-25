package property

import (
	"fmt"

	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// GradientNote reports a return whose spatial gradient could not be lowered:
// its gradient is left nil, and a renderer falls back to central differences
// for hits on that return.
type GradientNote struct {
	Return int    // 0-based index into the geometry returns
	Reason string // why the derivative failed
}

// LowerGradientReturns differentiates each of the program's geometry returns
// with respect to p and attaches the gradients to the primary package's last
// file GradientReturns, 1-to-1 with the returns. A return the derivative cannot
// go through (an op or builtin with no derivative rule) gets a nil gradient and
// a note; the others are unaffected.
//
// It runs where LowerReturnProperties does, after Expand, so every property
// operator in the returns is already expanded and the gradient is taken of the
// plain distance expression sdShape evaluates. The derivative is the one a
// .dp written at package scope gets: the same scope (own, dependency, builtin
// and dot-imported declarations and functions; qualified imports inlined) and
// the same sharing, so a gradient's repeated nodes are written once as the
// locals of a function synthesized beside the returns.
func LowerGradientReturns(prog *program.Program) ([]GradientNote, error) {
	if prog == nil || prog.Primary == nil || len(prog.Primary.Files) == 0 {
		return nil, nil
	}
	pkg := prog.Primary
	lastFile := pkg.Files[len(pkg.Files)-1]
	if len(lastFile.Returns) == 0 {
		return nil, nil
	}

	exp := &expansion{
		getImport: prog.GetImport,
		share:     true,
		deps:      prog.Dependencies,
		scopes:    map[*program.Package]*pkgScope{},
	}
	s := exp.scopeFor(pkg)
	names := &shareNames{taken: func(name string) bool {
		_, isDecl := s.decls[name]
		_, isFunc := s.funcsList[name]
		return isDecl || isFunc
	}}
	deriv := &derivative{
		decls:      s.decls,
		multiDecls: s.multiDecls,
		funcs:      s.funcsList,
		imports:    s.imports,
		own:        ownFunctionNames(pkg),
	}

	out := make([]*program.Operand, len(lastFile.Returns))
	var notes []GradientNote
	for i, ret := range lastFile.Returns {
		grad, err := returnGradient(deriv, ret)
		if err != nil {
			notes = append(notes, GradientNote{Return: i, Reason: err.Error()})
			continue
		}
		root, locals := sharedLocals(grad, names)
		if len(locals) > 0 {
			fnName := names.fresh("gradient")
			lastFile.FnTypes = append(lastFile.FnTypes, &program.FnType{
				Name: fnName,
				Groups: []*program.FnGroup{{
					ParamCount: 0,
					Functions:  []*program.Function{{Assignments: locals, Returns: []*program.Operand{root}}},
				}},
			})
			// Taken from here on, so the next return's names do not collide.
			s.funcsList[fnName] = nil
			root = program.NewFnCall("", fnName, nil, nil, nil)
		}
		out[i] = root
	}
	lastFile.GradientReturns = out
	return notes, nil
}

// returnGradient is ∂ret/∂p, recovering from a panic in the engine so one
// exotic return cannot take the whole program down with it.
func returnGradient(deriv *derivative, ret *program.Operand) (grad *program.Operand, err error) {
	defer func() {
		if r := recover(); r != nil {
			grad, err = nil, fmt.Errorf("derivative failed: %v", r)
		}
	}()
	if ret == nil {
		return nil, fmt.Errorf("empty return")
	}
	return deriv.gradient(ret)
}
