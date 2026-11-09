package build

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// VisitReference implements [program.OperandVisitor].
func (b *Builder) VisitReference(op *program.Operand, ref *program.Reference) (Expression, error) {
	if ref.Prefix == "" {
		if expr, ok := b.getInlinedBinding(ref.Name); ok {
			return expr, nil
		}
	}
	if len(op.ConstValues) > 0 && !b.isParameter(op) {
		t := op.Type
		if t == nil && len(op.ConstValues) > 1 {
			// The folder stamps a reference with its variable's values, but a
			// reference the inliner wrote in (a hoisted argument) can carry no
			// type of its own. The variable's gives the values their shape;
			// guessing would read a record as one flat vector.
			t = b.referencedType(ref)
		}
		return b.buildConstExpressionWithType(op, op.ConstValues, t)
	}
	var varInfo VarInfo
	if ref.Prefix != "" {
		pkgScope := b.delegate.GetScope().GetNamedImportScope(ref.Prefix)
		varInfo = pkgScope.GetVar(ref.Name)
	} else {
		varInfo = b.delegate.GetScope().GetVar(ref.Name)
	}
	// A name the type check resolved but the scope has not bound yet is one
	// declared further down: top-level variables build in order. Every backend
	// needs the binding, and two of them dereferenced the nil one.
	if varInfo == nil {
		// TODO: update assembler to catch all unbound/forward-referenced
		// variables and replace this check with an internal error and
		// eventually make it an invariant.
		return nil, ErrLoc(b.delegate.GetScope(), op.Start,
			fmt.Errorf("%s is used before it is declared: top-level variables are evaluated in order, so declare it above this line", ref.Name))
	}

	expr, err := b.delegate.BuildVarRef(op, varInfo)
	if err != nil {
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
	}

	return expr, nil
}

// referencedType is the type of the variable ref names, or nil.
func (b *Builder) referencedType(ref *program.Reference) *program.Type {
	var varInfo VarInfo
	if ref.Prefix != "" {
		if pkgScope := b.delegate.GetScope().GetNamedImportScope(ref.Prefix); pkgScope != nil {
			varInfo = pkgScope.GetVar(ref.Name)
		}
	} else {
		varInfo = b.delegate.GetScope().GetVar(ref.Name)
	}
	if varInfo == nil {
		return nil
	}
	return varInfo.Type()
}

// VisitValue implements [program.OperandVisitor].
func (b *Builder) VisitValue(op *program.Operand, value float64) (Expression, error) {
	return b.delegate.BuildValue(op, value)
}
