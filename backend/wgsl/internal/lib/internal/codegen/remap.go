package codegen

import (
	"fmt"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) Remap(remapOp *program.Operand, targetOp *program.Operand, expr build.Expression, targetExpr build.Expression) (build.Expression, error) {

	current := d.currentScope()

	// Create a new "p" variable
	pExpr := expr.(wgsl.Expression)
	pWgslName := d.GetScope().NewVarName("gsl_p")
	pVar := wgsl.NewPVar(d.currBScope, pWgslName, pExpr)

	// Encapsulate the target in a helper that re-derives it under the remapped p,
	// UNLESS the target is a function call that threads the remapped p through
	// its arguments on the direct path below. That holds only when the call
	// operand actually lowered to a real (non-inlined) call expression, which
	// passes the remapped p as its first argument.
	//
	// An inlined body does NOT qualify, even when its own inner calls take p
	// first: whether those calls received the remapped point or the enclosing
	// one was decided when the body was built, and the built tree cannot say
	// which — an inlined body is opaque here, a closure over expressions this
	// code cannot inspect. Assuming it threaded silently dropped the coordinate
	// for every target that reads p directly: a bare length(p), an operator
	// expression such as an inlined plane()'s dot(p - c, n), and a user
	// function whose body is a remap called outside one (the Aerospace
	// Bracket's mirrored side draft). Those all evaluated at the original p.
	// The helper is correct in every case; it costs a function.
	threadsP := false
	if targetOp.GetFnCall() != nil {
		switch t := targetExpr.(type) {
		case *wgsl.FuncCall:
			// A call only threads the remapped p when it actually takes p as its
			// first argument (a coordinate function like sphere(p, radius), whose
			// body re-derives under the remapped p via remapCtx). A combinator like
			// unionRound(list, k) does NOT take p: its p-dependence is baked into
			// already-built arguments (e.g. inlined plane() distances) that remapCtx
			// never reaches, so it needs the re-derivation helper below — otherwise
			// the remap is silently dropped for those nested args.
			//
			// cropToBox, which a declared box applies, is both: it takes p first
			// for its box, but its shape is a field argument built against the
			// enclosing point, so it needs the helper as a combinator does, or
			// the box moves and the shape it crops stays behind.
			threadsP = t.PVarIsFirstArg && t.WGSLName != cropToBoxWGSLName
		}
	}
	if !threadsP {
		fnName := d.uniqueTopLevelName("gsl_" + d.prefixWithScopeNames(d.currBScope, "remap"))

		params := d.GetScope().ParamVars()
		refs := make([]*wgsl.VarInfo, 0, len(current.globalRefs)+len(params))
		for _, ref := range current.globalRefs {
			// A primary-scope field var (e.g. a top-level `side = ...`) must
			// re-derive inside the helper under the remapped p, so it is NOT
			// captured as a parameter. Non-primary refs (import consts,
			// enclosing-function params/locals) are captured and passed at the
			// call site, as are p/time and other non-re-derivable primary vars
			// (they carry no definition to re-emit). At top level
			// current.globalRefs holds no re-derivable primary-scope field vars
			// (they're locals, never recorded), so this is a no-op there.
			if ref.Scope().InPrimaryScope() && ref.CanRederive() {
				continue
			}
			refs = append(refs, ref)
		}
		if !d.GetScope().IsInlinedScope() {
			for _, param := range params {
				refs = append(refs, param.(*wgsl.VarInfo))
			}
		}

		returnExpr := targetExpr
		returnType := targetExpr.(wgsl.Expression).WGSLType()
		if targetExpr.Type().IsFunction() {
			return nil, fmt.Errorf("cannot pass a function as an argument to remap, perhaps missing () so that it is a function call?")
		}

		// Build the helper as if in the primary build scope, so the target and
		// the full transitive closure of field vars it references re-derive
		// under the remapped p — the same self-contained codegen a top-level
		// remap already produces. ShouldGenerate then re-emits referenced field
		// vars (primary-scope) as `let`s, and recordRef skips them
		// (primary-locals) so they don't also become params.
		//
		// tempScope is copied from the ENCLOSING scope (prevScope) so its parent
		// and import-prefix cache are preserved — the helper is emitted in the
		// same place as before, so emission order (and existing goldens) is
		// unchanged. Only its build scope is overridden to primary, which is
		// what drives re-derivation. Only the agent-scope and current
		// build-scope pointers are swapped; SetScope is avoided (it has
		// importingScope-cycle side effects). At top level all of this is a
		// no-op (prevScope is already the primary agent scope).
		prevScope := d.currentScope()
		tempScope := prevScope.copy(fnName)
		tempScope.bScope = d.primaryBScope
		if prevScope.bScope.IsFileScope() {
			tempScope.parent = d.rootScope()
		} else {
			tempScope.parent = prevScope.parent
			if tempScope.parent.bScope.IsFileScope() {
				tempScope.parent = tempScope.parent.parent
			}
		}

		var mergedRefs []*wgsl.VarInfo
		var err error
		d.withScopedAgent(tempScope, func() {
			// Params arrive through the signature, so suppress re-emission of their
			// definitions in the body — otherwise, now that the build scope is
			// primary, an import const would re-derive and collide with its
			// parameter ("redeclaration"). Field vars are not in refs (filtered
			// above), so they still re-derive.
			for _, ref := range refs {
				tempScope.markGenerated(ref.WGSLName())
			}

			var actualFnName string
			mergedRefs, actualFnName, err = d.buildFunc(fnName, fnName, refs, nil, nil, returnType, []build.Expression{returnExpr})
			if err == nil {
				fnName = actualFnName
			}
		})
		if err != nil {
			return nil, err
		}

		targetExpr, err = d.buildFuncCall(remapOp, fnName, mergedRefs, nil, returnType)
		if err != nil {
			return nil, err
		}
	}

	var reads []string
	seen := make(map[string]struct{})
	for _, r := range pExpr.Reads() {
		if _, ok := seen[r]; !ok {
			seen[r] = struct{}{}
			reads = append(reads, r)
		}
	}
	for _, r := range targetExpr.(wgsl.Expression).Reads() {
		if r == program.PVarName {
			continue
		}
		if _, ok := seen[r]; !ok {
			seen[r] = struct{}{}
			reads = append(reads, r)
		}
	}

	return wgsl.NewWithPExpr(targetOp, targetExpr.(wgsl.Expression), pVar, reads), nil
}
