package property

import (
	"fmt"

	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// LowerReturnProperties evaluates .color and .material properties across
// the program's primary package and attaches the lowered operand trees
// to the primary package's last file (ColorReturns and MaterialReturns).
//
// Where a shape inside a return sets a material with a height, it also lowers
// the return's displaced distance (DisplacedReturns) and the declarations it
// reads, and switches each layered shape's colour and material by layer: see
// layers.go.
func LowerReturnProperties(prog *program.Program) error {
	if len(prog.Primary.Files) == 0 {
		return nil
	}
	lay := analyzeLayers(prog)
	lastFile := prog.Primary.Files[len(prog.Primary.Files)-1]
	// A query's derivatives were left in its returns
	// (ExpandSharedKeepingReturnDerivatives) so the layers above could carry
	// them into the displaced returns; the plain returns take theirs now, as
	// Expand would have, before their colour and material are read.
	if err := expandReturnDerivs(prog, lastFile.Returns); err != nil {
		return err
	}
	// Each return's working tolerance, for a viewer to draw the part at its
	// most or least material; nothing for a document that declares none.
	if err := lowerToleranceReturns(prog); err != nil {
		return err
	}
	colorFields, err := layeredPropertyFields(prog.Primary, prog.GetImport, colorSpec, lay)
	if err != nil {
		return fmt.Errorf("lower color returns: %w", err)
	}
	matFields, err := layeredPropertyFields(prog.Primary, prog.GetImport, materialSpec, lay)
	if err != nil {
		return fmt.Errorf("lower material returns: %w", err)
	}
	if colorFields != nil {
		lastFile.ColorReturns = colorFields
	}
	if matFields != nil {
		lastFile.MaterialReturns = matFields
	}
	lay.attach(prog)
	// The displaced returns take theirs of the layered distances, with the
	// declarations the layers just added in scope: read at a condition, a
	// derivative is the coated field's.
	return expandReturnDerivs(prog, lastFile.DisplacedReturns)
}

// expandReturnDerivs replaces each DERIV operation in ops (a .d(var), .dp or
// .dtime left there by ExpandSharedKeepingReturnDerivatives) with its
// expansion against the primary package's declarations as they now stand,
// innermost first.
func expandReturnDerivs(prog *program.Program, ops []*program.Operand) error {
	if !anyDeriv(ops) {
		return nil
	}
	pkg := prog.Primary
	exp := &expansion{getImport: prog.GetImport, deps: prog.Dependencies, scopes: map[*program.Package]*pkgScope{}}
	s := exp.scopeFor(pkg)
	d := &derivative{decls: s.decls, multiDecls: s.multiDecls, funcs: s.funcsList, imports: s.imports, own: ownFunctionNames(pkg)}
	// Shared as ExpandShared shares: an expansion's nodes reached more than
	// once become the locals of a function synthesized beside it, or every
	// pass after this copies the graph out as a tree, which grows with each
	// remap the derivative goes through.
	file := pkg.Files[len(pkg.Files)-1]
	share := func(expanded *program.Operand) *program.Operand {
		root, locals := sharedLocals(expanded, s.names)
		if len(locals) == 0 {
			return expanded
		}
		fnName := s.names.fresh("shared")
		file.FnTypes = append(file.FnTypes, &program.FnType{
			Name: fnName,
			Groups: []*program.FnGroup{{
				ParamCount: 0,
				Functions:  []*program.Function{{Assignments: locals, Returns: []*program.Operand{root}}},
			}},
		})
		return program.NewFnCall("", fnName, nil, nil, nil)
	}
	var walk func(op *program.Operand) (*program.Operand, error)
	walk = func(op *program.Operand) (*program.Operand, error) {
		if op == nil {
			return nil, nil
		}
		switch t := op.T.(type) {
		case *program.Operand_Expression:
			for i, sub := range t.Expression.Operands {
				w, err := walk(sub)
				if err != nil {
					return nil, err
				}
				t.Expression.Operands[i] = w
			}
			if t.Expression.Operation == program.Operation_DERIV {
				expanded, err := d.expandDeriv(t.Expression.Operands)
				if err != nil {
					return nil, err
				}
				return share(expanded), nil
			}
		case *program.Operand_FnCall:
			for i, a := range t.FnCall.Arguments {
				w, err := walk(a)
				if err != nil {
					return nil, err
				}
				t.FnCall.Arguments[i] = w
			}
		}
		return op, nil
	}
	for i, op := range ops {
		w, err := walk(op)
		if err != nil {
			return fmt.Errorf("differentiating return %d: %w", i+1, err)
		}
		ops[i] = w
	}
	return nil
}

// anyDeriv reports whether a DERIV operation appears in ops.
func anyDeriv(ops []*program.Operand) bool {
	var found func(op *program.Operand) bool
	found = func(op *program.Operand) bool {
		if op == nil {
			return false
		}
		switch t := op.T.(type) {
		case *program.Operand_Expression:
			if t.Expression.Operation == program.Operation_DERIV {
				return true
			}
			for _, sub := range t.Expression.Operands {
				if found(sub) {
					return true
				}
			}
		case *program.Operand_FnCall:
			for _, a := range t.FnCall.Arguments {
				if found(a) {
					return true
				}
			}
		}
		return false
	}
	for _, op := range ops {
		if found(op) {
			return true
		}
	}
	return false
}

// LowerColorReturns evaluates .color properties across the program's primary
// package and attaches the lowered color operand trees to the primary
// package's last file ColorReturns.
func LowerColorReturns(prog *program.Program) error {
	if len(prog.Primary.Files) == 0 {
		return nil
	}
	colorFields, err := ColorFields(prog.Primary, prog.GetImport)
	if err != nil {
		return fmt.Errorf("lower color returns: %w", err)
	}
	if colorFields != nil {
		lastFile := prog.Primary.Files[len(prog.Primary.Files)-1]
		lastFile.ColorReturns = colorFields
	}
	return nil
}

// LowerMaterialReturns evaluates .material properties across the program's primary
// package and attaches the lowered material operand trees to the primary
// package's last file MaterialReturns.
func LowerMaterialReturns(prog *program.Program) error {
	if len(prog.Primary.Files) == 0 {
		return nil
	}
	matFields, err := MaterialFields(prog.Primary, prog.GetImport)
	if err != nil {
		return fmt.Errorf("lower material returns: %w", err)
	}
	if matFields != nil {
		lastFile := prog.Primary.Files[len(prog.Primary.Files)-1]
		lastFile.MaterialReturns = matFields
	}
	return nil
}
