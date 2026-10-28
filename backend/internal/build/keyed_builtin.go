package build

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/program"
)

// A keyed builtin reads data baked at compile time from its constant
// arguments: gaborSpinodoid's kernel table is chosen by its period, cone
// angles, seed and the rest. Those constants usually arrive through a library
// wrapper (lattice/spinodoid's gabor(...)), and a backend that keeps the
// wrapper a real function — WGSL emits it as a helper — sees them as runtime
// parameters, unable to name the table. So a function that reaches a keyed
// builtin is always inlined, whatever its arguments: the builtin then sees the
// caller's constants in every backend, and every backend enumerates the same
// tables. text() is inlined for the same reason (a constant string must reach
// its glyph fold); this rule differs only in not requiring the other arguments
// (a spinodoid's scale and orientation fields) to be constant.
func isKeyedBuiltin(name string) bool {
	return program.IsGaborSpinodoid(name)
}

// reachesKeyedBuiltin reports whether fn's body calls a keyed builtin,
// directly or through functions it calls. The receiver is fn's enclosing
// scope, where its free references resolve.
func (s *Scope) reachesKeyedBuiltin(fn *program.Function, visited map[string]bool) bool {
	if fn == nil {
		return false
	}
	localFns := anonFnDecls(fn)
	for _, assign := range fn.Assignments {
		if s.operandReachesKeyedBuiltin(assign.Expression, visited, localFns) {
			return true
		}
	}
	for _, ret := range fn.Returns {
		if s.operandReachesKeyedBuiltin(ret, visited, localFns) {
			return true
		}
	}
	return false
}

func (s *Scope) operandReachesKeyedBuiltin(op *program.Operand, visited map[string]bool, localFns map[string]*program.Function) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		for _, child := range t.Expression.Operands {
			if s.operandReachesKeyedBuiltin(child, visited, localFns) {
				return true
			}
		}
		return false
	case *program.Operand_FnCall:
		fc := t.FnCall
		for _, arg := range fc.Arguments {
			if s.operandReachesKeyedBuiltin(arg, visited, localFns) {
				return true
			}
		}
		if fc.Reference.Prefix == "" && program.BuiltinNamed(fc.Reference.Name) != nil {
			return isKeyedBuiltin(fc.Reference.Name)
		}
		visitedKey := fmt.Sprintf("%s.%s/%d", fc.Reference.Prefix, fc.Reference.Name, len(fc.Arguments))
		if visited[visitedKey] {
			return false
		}
		visited[visitedKey] = true

		var callee *program.Function
		calleeScope := s
		if fc.Reference.Prefix == "" && localFns != nil {
			callee = localFns[fc.Reference.Name]
		}
		if callee == nil {
			fnScope := s
			if fc.Reference.Prefix != "" {
				fnScope = fnScope.GetNamedImportScope(fc.Reference.Prefix)
			}
			if fnScope == nil {
				return false
			}
			_, callee, calleeScope = fnScope.FindFunction(fc.Reference.Name, len(fc.Arguments))
		}
		return calleeScope.reachesKeyedBuiltin(callee, visited)
	}
	return false
}

func (b *Builder) pushInlineSite(fnDecl *program.FnType, op *program.Operand) {
	site := fnDecl.Name
	if op != nil && op.Start != nil {
		site = fmt.Sprintf("%s@%d:%d", fnDecl.Name, op.Start.Line, op.Start.Column)
	}
	b.inlineSites = append(b.inlineSites, site)
}

func (b *Builder) popInlineSite() {
	b.inlineSites = b.inlineSites[:len(b.inlineSites)-1]
}

// InlineCallChain names the chain of inlined calls the builder is inside,
// outermost first. Inlining is decided by the builder, not the backend, and a
// function reaching a keyed builtin is always inlined, so every backend sees
// the same chain at the same keyed call: with the builtin's constant
// arguments it identifies one call site across backends, which is how a
// table one backend bakes is matched to the call another backend emits.
func (s *Scope) InlineCallChain() string {
	if s == nil || s.builder == nil {
		return ""
	}
	return strings.Join(s.builder.inlineSites, "/")
}

// KeyedCallSite identifies a keyed builtin call across backends: the inline
// chain it was built in, then the call's own source position. The position
// matters where the assembler has already inlined a library wrapper: it
// rewrites the call to the builtin in place and keeps the caller's position,
// leaving no chain to tell two such calls apart.
func KeyedCallSite(s *Scope, op *program.Operand) string {
	site := s.InlineCallChain()
	if op != nil && op.Start != nil {
		site += fmt.Sprintf("#%d:%d", op.Start.Line, op.Start.Column)
	}
	return site
}
