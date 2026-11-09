package program

import (
	"sort"

	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// UpdateImports prepares a GSL package by adding the implicit dot-import of "builtin"
// (unless it is the "builtin" package itself) and auto-importing any referenced standard packages.
func UpdateImports(pkg *Package, importPath string) {
	if pkg == nil {
		return
	}
	if importPath == "builtin" || importPath == "omega" {
		return
	}

	c := &importCollector{found: make(map[string]bool), seen: make(map[*Operand]bool)}

	for _, file := range pkg.Files {
		if file == nil {
			continue
		}
		// Add implicit dot-import of "builtin" if not already present
		hasBuiltin := false
		explicit := make(map[string]bool)
		for _, imp := range file.Imports {
			if imp != nil {
				explicit[imp.Path] = true
				if imp.Path == "builtin" && imp.Prefix == "." {
					hasBuiltin = true
				}
			}
		}
		if !hasBuiltin {
			implicitImp := NewImport(".", "builtin", &Location{Line: 1, Column: 1}, &Location{Line: 1, Column: 1})
			file.Imports = append([]*Import{implicitImp}, file.Imports...)
		}

		// Auto-import referenced system packages
		c.collectFromPackageFile(file)

		var toAdd []string
		for p := range c.found {
			if !explicit[p] {
				toAdd = append(toAdd, p)
			}
		}
		sort.Strings(toAdd)
		for _, p := range toAdd {
			autoImp := NewImport("", p, &Location{Line: 1, Column: 1}, &Location{Line: 1, Column: 1})
			file.Imports = append(file.Imports, autoImp)
		}
		// Reset for next file. The seen set resets with it: an operand two
		// files share must still contribute its package to each file's imports.
		c.found = make(map[string]bool)
		c.seen = make(map[*Operand]bool)
	}
}

type importCollector struct {
	found map[string]bool
	// seen records the operands already walked. The optimizer hands us a DAG,
	// not a tree: inlining and common subexpressions leave several parents
	// pointing at one operand. Walking such a node once per path that reaches
	// it costs time exponential in the sharing depth, which on a large
	// document does not finish. A node contributes the same packages however
	// many times it is reached, so visiting it once is equivalent.
	seen map[*Operand]bool
}

// visit walks an operand the first time it is reached and skips it after.
func (c *importCollector) visit(op *Operand) {
	if op == nil {
		return
	}
	if c.seen[op] {
		return
	}
	c.seen[op] = true
	_, _ = VisitOperand(op, c)
}

func (c *importCollector) VisitValue(op *Operand, val float64) (any, error) {
	return nil, nil
}

func (c *importCollector) VisitReference(op *Operand, ref *Reference) (any, error) {
	if ref.Prefix != "" {
		if pkgPath, ok := gsllibs.PathForPrefix(ref.Prefix); ok {
			c.found[pkgPath] = true
		}
	}
	return nil, nil
}

func (c *importCollector) VisitExpression(op *Operand, expr *Expression) (any, error) {
	if expr.Operation == Operation_ACCESS && len(expr.Operands) > 0 {
		if refOp, ok := expr.Operands[0].T.(*Operand_Reference); ok && refOp.Reference != nil {
			if pkgPath, ok := gsllibs.PathForPrefix(refOp.Reference.Name); ok {
				c.found[pkgPath] = true
			}
		}
	}
	for _, operand := range expr.Operands {
		c.visit(operand)
	}
	return nil, nil
}

func (c *importCollector) VisitFnCall(op *Operand, fc *FnCall) (any, error) {
	if fc.Reference != nil && fc.Reference.Prefix != "" {
		if pkgPath, ok := gsllibs.PathForPrefix(fc.Reference.Prefix); ok {
			c.found[pkgPath] = true
		}
	}
	for _, operand := range fc.Arguments {
		c.visit(operand)
	}
	return nil, nil
}

func (c *importCollector) VisitPropRef(op *Operand, ref *PropertyRef) (any, error) {
	if ref.Reference != nil && ref.Reference.Prefix != "" {
		if pkgPath, ok := gsllibs.PathForPrefix(ref.Reference.Prefix); ok {
			c.found[pkgPath] = true
		}
	}
	return nil, nil
}

func (c *importCollector) collectFromAssignment(assign *Assignment) {
	if assign == nil {
		return
	}
	c.visit(assign.Expression)
	for _, v := range assign.Variables {
		if v != nil {
			for _, prop := range v.Properties {
				if prop != nil {
					c.visit(prop.Expression)
				}
			}
		}
	}
}

func (c *importCollector) collectFromFunction(fn *Function) {
	if fn == nil {
		return
	}
	for _, assign := range fn.Assignments {
		c.collectFromAssignment(assign)
	}
	for _, op := range fn.Returns {
		c.visit(op)
	}
	for _, fnType := range fn.FnTypes {
		if fnType != nil {
			for _, group := range fnType.Groups {
				if group != nil {
					for _, childFn := range group.Functions {
						c.collectFromFunction(childFn)
					}
				}
			}
		}
	}
}

func (c *importCollector) collectFromPackageFile(file *PackageFile) {
	if file == nil {
		return
	}
	for _, assign := range file.Assignments {
		c.collectFromAssignment(assign)
	}
	for _, op := range file.Returns {
		c.visit(op)
	}
	for _, op := range file.ColorReturns {
		c.visit(op)
	}
	for _, op := range file.MaterialReturns {
		c.visit(op)
	}
	for _, fnType := range file.FnTypes {
		if fnType != nil {
			for _, group := range fnType.Groups {
				if group != nil {
					for _, fn := range group.Functions {
						c.collectFromFunction(fn)
					}
				}
			}
		}
	}
}
