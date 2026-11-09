package program

import "strings"

// StaticElements returns the elements a collection operand denotes, when that
// can be decided from AST shape alone: a literal collect(...) expression, or a
// name bound to one in a dependency package, an enclosing function, or file
// scope.
//
// It is deliberately weaker than lower.extractElements, which also unrolls a
// sized operand through index accesses and treats a scalar as a singleton.
// Both of those read op.Type, which is only populated by types.Analyze, so a
// caller that runs before type analysis — the closure pass — has nothing but
// these two structural paths.
//
// Returns nil when the collection cannot be resolved.
func StaticElements(prog *Program, file *PackageFile, fnStack []*Function, op *Operand) []*Operand {
	if op == nil {
		return nil
	}

	if expr, ok := op.T.(*Operand_Expression); ok && expr.Expression != nil &&
		expr.Expression.Operation == Operation_COLLECT {
		return expr.Expression.Operands
	}

	ref, ok := op.T.(*Operand_Reference)
	if !ok || ref.Reference == nil {
		return nil
	}
	name := ref.Reference.Name

	// A cross-package variable (prefix.name): search the dependency whose path
	// the prefix names for an assignment holding a collect(...).
	if ref.Reference.Prefix != "" && prog != nil {
		for _, dep := range prog.Dependencies {
			if dep == nil {
				continue
			}
			simpleName := dep.Path[strings.LastIndex(dep.Path, "/")+1:]
			if dep.Path != ref.Reference.Prefix && simpleName != ref.Reference.Prefix &&
				!strings.HasSuffix(dep.Path, "/"+ref.Reference.Prefix) {
				continue
			}
			for _, f := range dep.Files {
				if els := collectAssignedTo(f.Assignments, name); els != nil {
					return els
				}
			}
		}
	}

	// Enclosing functions, innermost first, then file scope.
	for i := len(fnStack) - 1; i >= 0; i-- {
		if fn := fnStack[i]; fn != nil {
			if els := collectAssignedTo(fn.Assignments, name); els != nil {
				return els
			}
		}
	}
	if file != nil {
		if els := collectAssignedTo(file.Assignments, name); els != nil {
			return els
		}
	}

	return nil
}

// collectAssignedTo returns the operands of a collect(...) assigned to name in
// assigns, or nil if name is not assigned one there.
func collectAssignedTo(assigns []*Assignment, name string) []*Operand {
	for _, assign := range assigns {
		if assign == nil || assign.Expression == nil {
			continue
		}
		for _, v := range assign.Variables {
			if v.Name != name {
				continue
			}
			if sub, ok := assign.Expression.T.(*Operand_Expression); ok && sub.Expression != nil &&
				sub.Expression.Operation == Operation_COLLECT {
				return sub.Expression.Operands
			}
		}
	}
	return nil
}
