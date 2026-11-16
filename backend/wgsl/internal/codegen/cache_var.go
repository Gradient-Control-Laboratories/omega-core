package codegen

import (
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// If this is not a local variable, record it in the global map
func (d *Delegate) recordRef(varInfo build.VarInfo) {
	d.currentScope().recordRef(varInfo)
}

// If this is not a local variable, record it in the global map
func (s *Scope) recordRef(varInfo build.VarInfo) {
	if varInfo.Name() == program.PVarName {
		// always keep p as the first global ref for remapping
		if len(s.globalRefs) == 0 || s.globalRefs[0].Name() != program.PVarName {
			s.globalRefs = append([]*wgsl.VarInfo{s.delegate.rootPVar}, s.globalRefs...)
		}
		if s.importingScope != nil {
			s.importingScope.recordRef(varInfo)
		}
		return
	}
	if s.bScope.IsLocalVar(varInfo.Name()) {
		return
	}
	for _, g := range s.globalRefs {
		if g.Name() == varInfo.Name() {
			if s.importingScope != nil {
				s.importingScope.recordRef(varInfo)
			}
			return
		}
	}
	s.globalRefs = append(s.globalRefs, varInfo.(*wgsl.VarInfo))
	if s.importingScope != nil {
		s.importingScope.recordRef(varInfo)
	}
}

// needSliderParam records the slider parameter a variable is defined by as a
// reference of the function now being generated, so that function takes it as
// a parameter and its caller passes it. The assignment recorded it where the
// slider was declared, the primary scope; a helper that re-derives the
// variable inside another function (a remap in a function's body) is
// generated in a scope that never saw that, and would read a name nothing
// there defines.
func (d *Delegate) needSliderParam(ref *wgsl.VarInfo) {
	param, ok := d.sliderParamVars[ref.Name()]
	if !ok || param == nil {
		return
	}
	if def, ok := ref.Expr().(*wgsl.VarRef); !ok || def.VarInfo != param {
		return
	}
	scope := d.currentScope()
	for _, g := range scope.globalRefs {
		if g.Name() == param.Name() {
			return
		}
	}
	scope.globalRefs = append(scope.globalRefs, param)
}
