package codegen

import (
	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/wgsl"
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
