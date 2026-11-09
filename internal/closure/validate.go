package closure

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/program"
)

func (t *transformer) validateReturns() error {
	if t.prog.Primary != nil {
		for _, file := range t.prog.Primary.Files {
			for _, ret := range file.Returns {
				if t.isFunctionValue(ret, file, nil) {
					return fmt.Errorf("cannot return a function; perhaps forgot parentheses () after function name?")
				}
			}
		}
	}

	var allFiles []*program.PackageFile
	if t.prog.Primary != nil {
		allFiles = append(allFiles, t.prog.Primary.Files...)
	}
	for _, dep := range t.prog.Dependencies {
		allFiles = append(allFiles, dep.Files...)
	}

	for _, file := range allFiles {
		for _, ft := range file.FnTypes {
			for _, fg := range ft.Groups {
				for _, fn := range fg.Functions {
					if len(fn.Returns) > 1 {
						for _, ret := range fn.Returns {
							if t.isFunctionValue(ret, file, fn) {
								return fmt.Errorf("a function that returns a function cannot return anything else")
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (t *transformer) isFunctionValue(op *program.Operand, file *program.PackageFile, fn *program.Function) bool {
	if op == nil {
		return false
	}
	switch val := op.T.(type) {
	case *program.Operand_Reference:
		ref := val.Reference
		if ref == nil {
			return false
		}
		if fn != nil {
			for _, nestedFt := range fn.FnTypes {
				if nestedFt.Name == ref.Name {
					return true
				}
			}
			for _, p := range fn.Parameters {
				if p.Name == ref.Name {
					return false
				}
			}
			// A local shadows a package function of the same name: it is a
			// function value only if what it was assigned is one.
			for _, a := range fn.Assignments {
				for _, v := range a.Variables {
					if v != nil && v.Name == ref.Name {
						if r, ok := a.Expression.T.(*program.Operand_Reference); ok && r.Reference != nil &&
							r.Reference.Prefix == "" && r.Reference.Name == ref.Name {
							return false
						}
						return t.isFunctionValue(a.Expression, file, fn)
					}
				}
			}
		}
		if ref.Prefix == "" {
			if fileMap, ok := t.fileFunctions[file]; ok {
				if _, found := fileMap[ref.Name]; found {
					return true
				}
			}
		} else {
			pkgPath := t.resolvePrefix(file, ref.Prefix)
			if pkgMap, ok := t.pkgFunctions[pkgPath]; ok {
				if _, found := pkgMap[ref.Name]; found {
					return true
				}
			}
		}
	case *program.Operand_FnCall:
		fc := val.FnCall
		if fc != nil && fc.Reference != nil {
			callee := t.lookupFunction(file, fc.Reference.Prefix, fc.Reference.Name, len(fc.Arguments))
			if callee != nil && t.functionReturnsFunction(callee) {
				return true
			}
		}
	case *program.Operand_Expression:
		// A list holding a function is itself a function value. It is not a
		// thing that can be returned or stored: it only means something where
		// it is folded, with every element in view.
		if val.Expression != nil && val.Expression.Operation == program.Operation_COLLECT {
			for _, sub := range val.Expression.Operands {
				if t.isFunctionValue(sub, file, fn) {
					return true
				}
			}
		}
	}
	return false
}

func (t *transformer) functionReturnsFunction(fn *program.Function) bool {
	if fn == nil || len(fn.Returns) == 0 {
		return false
	}
	for _, ret := range fn.Returns {
		if ref, ok := ret.T.(*program.Operand_Reference); ok && ref.Reference != nil {
			if strings.HasPrefix(ref.Reference.Name, program.AnonFnPrefix) {
				return true
			}
			for _, nestedFt := range fn.FnTypes {
				if nestedFt.Name == ref.Reference.Name {
					return true
				}
			}
		}
	}
	return false
}
