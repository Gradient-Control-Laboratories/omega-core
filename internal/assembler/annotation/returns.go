package annotation

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// ValidateReturns holds an annotated zero-argument function to the pair it
// promises: a material record and a colour, in that order.
//
// This is the check 6e52b040 deferred "to the manifest, where types are
// resolved", and it runs after types.Analyze rather than beside Validate for
// that reason. It applies only to ZERO-ARGUMENT functions, and that is a
// principled line rather than a convenience: every other function body is
// typed from a call site, since its parameter types come from the arguments,
// so an uncalled one is never typed and cannot be checked. A zero-parameter
// function has no argument types to supply, so types.Analyze can infer it
// standalone. A six-argument process function like machined.turned stays
// unverifiable here, and the thing that could check it — something that knows
// how to call it — is the same thing that could preview it.
//
// A constant carries no returns, so an annotated constant is left to Validate
// and its own rules.
func ValidateReturns(prog *program.Program) error {
	if prog == nil || prog.Primary == nil {
		return nil
	}
	pkgs := append([]*program.Package{prog.Primary}, prog.Dependencies...)
	for _, pkg := range pkgs {
		if pkg == nil {
			continue
		}
		for _, file := range pkg.Files {
			if file == nil {
				continue
			}
			for _, ann := range file.Annotations {
				if ann == nil || ann.Kind != Material {
					continue
				}
				if err := validatePairReturns(pkg, file, ann); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// validatePairReturns checks the one zero-argument overload of an annotated
// function, if it has one. An annotation marks a name rather than an arity, so
// a name with both a zero-argument form and a parameterised one is checked on
// the form that can be.
func validatePairReturns(pkg *program.Package, file *program.PackageFile, ann *program.Annotation) error {
	fn := zeroArgOverload(pkg, ann.Symbol)
	if fn == nil {
		return nil
	}
	if len(fn.Returns) != 2 {
		return fileErr(file, ann.Start, fmt.Sprintf(
			"@%s %q returns %d values, want 2: a material record and its colour",
			ann.Kind, ann.Symbol, len(fn.Returns)))
	}
	if err := wantMaterial(fn.Returns[0].GetType()); err != nil {
		return fileErr(file, ann.Start, fmt.Sprintf(
			"@%s %q: first return %s", ann.Kind, ann.Symbol, err))
	}
	if err := wantColor(fn.Returns[1].GetType()); err != nil {
		return fileErr(file, ann.Start, fmt.Sprintf(
			"@%s %q: second return %s", ann.Kind, ann.Symbol, err))
	}
	return nil
}

// zeroArgOverload is the function declared for name taking no parameters, or
// nil where the name has no such form.
func zeroArgOverload(pkg *program.Package, name string) *program.Function {
	for _, file := range pkg.Files {
		if file == nil {
			continue
		}
		for _, fnType := range file.FnTypes {
			if fnType == nil || fnType.Name != name {
				continue
			}
			for _, group := range fnType.Groups {
				if group == nil || group.ParamCount != 0 {
					continue
				}
				if len(group.Functions) > 0 {
					return group.Functions[0]
				}
			}
		}
	}
	return nil
}

// wantMaterial reports why t is not a material record: rows of four, four of
// them today. Later versions may append rows (see the material library's
// header), so the row count is a floor rather than an equality.
func wantMaterial(t *program.Type) error {
	if t == nil {
		return fmt.Errorf("has no resolved type")
	}
	if !t.IsMatrix() && !t.IsList() {
		return fmt.Errorf("is not a material record (want rows of four, got %s)", describe(t))
	}
	if t.Size < 4 {
		return fmt.Errorf("has %d rows, want at least 4 (%s)", t.Size, describe(t))
	}
	if elem := t.ElemType; elem == nil || elem.Size != 4 {
		return fmt.Errorf("has rows that are not four wide (%s)", describe(t))
	}
	return nil
}

// wantColor reports why t is not a colour: three components, sRGB in [0, 1].
func wantColor(t *program.Type) error {
	if t == nil {
		return fmt.Errorf("has no resolved type")
	}
	if !t.IsList() || t.Size != 3 {
		return fmt.Errorf("is not a colour (want three components, got %s)", describe(t))
	}
	return nil
}

// describe names a type's shape for an error message.
func describe(t *program.Type) string {
	if t == nil {
		return "nothing"
	}
	switch {
	case t.IsScalar():
		return "a scalar"
	case t.IsFunction():
		return "a function"
	case t.ElemType != nil:
		return fmt.Sprintf("%d of %s", t.Size, describe(t.ElemType))
	default:
		return fmt.Sprintf("%d values", t.Size)
	}
}
