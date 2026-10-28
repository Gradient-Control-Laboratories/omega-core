package propval

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/program"
)

var validAssignablePropNames = map[string]bool{
	"min":        true,
	"max":        true,
	"period":     true,
	"color":      true,
	"material":   true,
	"box":        true,
	"simple":     true,
	"detail":     true,
	"detailSize": true,
	// A shape's own surface frame: u and v are coordinates on it, authored by
	// the primitives, so a material can read where it sits on the surface
	// rather than being handed a frame that has to be kept in step by hand.
	// Linear coordinates are millimetres, angular ones radians.
	"u": true,
	"v": true,
}

// Validate inspects the AST of prog prior to property expansion, verifying
// property assignment semantics, scope rules, type and value domains,
// cycle-free dependencies, and operand constraints.
func Validate(prog *program.Program) error {
	if prog == nil || prog.Primary == nil {
		return nil
	}

	// Constant properties may name an imported constant (.max = tau), so the
	// evaluator reads imports from the packages collection already loaded.
	getImport := func(path string) (*program.Package, error) {
		for _, dep := range prog.Dependencies {
			if dep != nil && dep.Path == path {
				return dep, nil
			}
		}
		return nil, nil
	}

	// 1. Validate property assignments and values across primary and dependencies
	for _, pkg := range allPackages(prog) {
		if err := validatePackage(pkg, getImport); err != nil {
			return err
		}
	}

	// 2. Validate DAG dependency cycles on properties
	if err := CheckCycles(prog); err != nil {
		return err
	}

	return nil
}

func allPackages(prog *program.Program) []*program.Package {
	pkgs := []*program.Package{prog.Primary}
	pkgs = append(pkgs, prog.Dependencies...)
	return pkgs
}

func validatePackage(pkg *program.Package, getImport program.GetImportFunc) error {
	if pkg == nil {
		return nil
	}

	for _, f := range pkg.Files {
		if f == nil {
			continue
		}

		// Top-level assignments
		for _, a := range f.Assignments {
			if err := validateAssignment(a, true, pkg, f, getImport); err != nil {
				return err
			}
		}

		// Function-level assignments
		for _, ft := range f.FnTypes {
			if err := validateFnType(ft, pkg, f, getImport); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateFnType(ft *program.FnType, pkg *program.Package, file *program.PackageFile, getImport program.GetImportFunc) error {
	if ft == nil {
		return nil
	}
	for _, g := range ft.Groups {
		if g == nil {
			continue
		}
		for _, fn := range g.Functions {
			if fn == nil {
				continue
			}
			for _, param := range fn.Parameters {
				if param.Range == nil || param.Range.IsSynthetic() {
					continue
				}
				if param.Range.Min == nil || param.Range.Max == nil {
					return fileErr(file, param.Range.Start, fmt.Sprintf("parameter %q @range requires both min and max bounds", param.Name))
				}
				minVals, err := evalConstExpr(param.Range.Min, pkg, getImport)
				if err != nil {
					return fileErr(file, param.Range.Min.Start, fmt.Sprintf("parameter %q range min must evaluate to a compile-time constant: %v", param.Name, err))
				}
				maxVals, err := evalConstExpr(param.Range.Max, pkg, getImport)
				if err != nil {
					return fileErr(file, param.Range.Max.Start, fmt.Sprintf("parameter %q range max must evaluate to a compile-time constant: %v", param.Name, err))
				}
				if len(minVals) != 1 || len(maxVals) != 1 {
					return fileErr(file, param.Range.Min.Start, fmt.Sprintf("parameter %q range min and max must evaluate to scalar values", param.Name))
				}
				if minVals[0] >= maxVals[0] {
					return fileErr(file, param.Range.Min.Start, fmt.Sprintf("parameter %q range min (%v) must be less than max (%v)", param.Name, minVals[0], maxVals[0]))
				}
			}
			for _, a := range fn.Assignments {
				if err := validateAssignment(a, false, pkg, file, getImport); err != nil {
					return err
				}
			}
			for _, nested := range fn.FnTypes {
				if err := validateFnType(nested, pkg, file, getImport); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateAssignment(a *program.Assignment, isTopLevel bool, pkg *program.Package, file *program.PackageFile, getImport program.GetImportFunc) error {
	// Validate explicit @range annotation if present
	if a.Range != nil && !a.Range.IsSynthetic() {
		if !isTopLevel {
			return fileErr(file, a.Range.Start, "slider @range annotations on variables are only permitted on top-level file scope declarations")
		}
		if isCollectionOrShapeTarget(a.Expression) {
			return fileErr(file, a.Range.Start, "range annotation can only be attached to scalar numeric variables")
		}
		if a.Range.Min == nil || a.Range.Max == nil {
			return fileErr(file, a.Range.Start, "range annotation requires both min and max bounds")
		}
		minVals, err := evalConstExpr(a.Range.Min, pkg, getImport)
		if err != nil {
			return fileErr(file, a.Range.Min.Start, fmt.Sprintf("range min must evaluate to a compile-time constant: %v", err))
		}
		maxVals, err := evalConstExpr(a.Range.Max, pkg, getImport)
		if err != nil {
			return fileErr(file, a.Range.Max.Start, fmt.Sprintf("range max must evaluate to a compile-time constant: %v", err))
		}
		if len(minVals) != 1 || len(maxVals) != 1 {
			return fileErr(file, a.Range.Min.Start, "range min and max must evaluate to scalar values")
		}
		if minVals[0] >= maxVals[0] {
			return fileErr(file, a.Range.Min.Start, fmt.Sprintf("range min (%v) must be less than max (%v)", minVals[0], maxVals[0]))
		}
		if a.Range.Period != nil {
			pVals, err := evalConstExpr(a.Range.Period, pkg, getImport)
			if err != nil {
				return fileErr(file, a.Range.Period.Start, fmt.Sprintf("range period must evaluate to a compile-time constant: %v", err))
			}
			if len(pVals) != 1 || pVals[0] <= 0 {
				return fileErr(file, a.Range.Period.Start, "range period must be strictly positive")
			}
		}
		if a.Expression != nil {
			initVals, err := evalConstExpr(a.Expression, pkg, getImport)
			if err == nil && len(initVals) == 1 {
				if initVals[0] < minVals[0] || initVals[0] > maxVals[0] {
					return fileErr(file, a.Expression.Start, fmt.Sprintf("slider default value %v is outside range [%v, %v]", initVals[0], minVals[0], maxVals[0]))
				}
			}
		}
	}

	// Validate expression operands (e.g. check for invalid auto or metadata property reads)
	if a.Expression != nil {
		if err := validateExpression(a.Expression, false, file); err != nil {
			return err
		}
	}

	for _, v := range a.Variables {
		if v == nil || len(v.Properties) == 0 {
			continue
		}

		propMap := make(map[string]*program.Property)
		for _, p := range v.Properties {
			if p == nil {
				continue
			}

			if !validAssignablePropNames[p.Name] {
				return fileErr(file, p.Start, fmt.Sprintf("invalid property %q assigned to variable %q", p.Name, v.Name))
			}

			if _, exists := propMap[p.Name]; exists {
				return fileErr(file, p.Start, fmt.Sprintf("property %s.%s already assigned", v.Name, p.Name))
			}
			propMap[p.Name] = p

			// Validate expression inside property
			isBox := p.Name == "box"
			if p.Expression != nil {
				if err := validateExpression(p.Expression, isBox, file); err != nil {
					return err
				}
			}
		}

		// Validate slider property constraints (.min, .max, .period)
		if propMap["min"] != nil || propMap["max"] != nil || propMap["period"] != nil {
			if !isTopLevel {
				loc := a.Start
				if p := propMap["min"]; p != nil {
					loc = p.Start
				} else if p := propMap["max"]; p != nil {
					loc = p.Start
				} else if p := propMap["period"]; p != nil {
					loc = p.Start
				}
				return fileErr(file, loc, "slider properties (.min, .max, .period) are only permitted on top-level file scope declarations")
			}

			// Must be a scalar assignment
			if isCollectionOrShapeTarget(a.Expression) {
				loc := a.Start
				return fileErr(file, loc, "slider properties can only be assigned to scalar numeric variables")
			}

			// Strict pairing requirement: both .min and .max required
			if propMap["min"] == nil || propMap["max"] == nil {
				loc := a.Start
				if propMap["min"] != nil {
					loc = propMap["min"].Start
				} else if propMap["max"] != nil {
					loc = propMap["max"].Start
				} else if propMap["period"] != nil {
					loc = propMap["period"].Start
				}
				return fileErr(file, loc, fmt.Sprintf("slider variable %q requires both .min and .max bounds", v.Name))
			}

			// Enforce compile-time constant evaluation on .min and .max
			minVals, err := evalConstExpr(propMap["min"].Expression, pkg, getImport)
			if err != nil {
				return fileErr(file, propMap["min"].Start, fmt.Sprintf("property .min on variable %q must evaluate to a compile-time constant: %v", v.Name, err))
			}
			maxVals, err := evalConstExpr(propMap["max"].Expression, pkg, getImport)
			if err != nil {
				return fileErr(file, propMap["max"].Start, fmt.Sprintf("property .max on variable %q must evaluate to a compile-time constant: %v", v.Name, err))
			}

			if len(minVals) != 1 || len(maxVals) != 1 {
				return fileErr(file, propMap["min"].Start, fmt.Sprintf("slider properties .min and .max on variable %q must evaluate to scalar values", v.Name))
			}

			if minVals[0] >= maxVals[0] {
				return fileErr(file, propMap["min"].Start, fmt.Sprintf("invalid slider range: min (%v) must be less than max (%v)", minVals[0], maxVals[0]))
			}

			if periodProp := propMap["period"]; periodProp != nil {
				pVals, err := evalConstExpr(periodProp.Expression, pkg, getImport)
				if err != nil {
					return fileErr(file, periodProp.Start, fmt.Sprintf("property .period on variable %q must evaluate to a compile-time constant: %v", v.Name, err))
				}
				if len(pVals) != 1 || pVals[0] <= 0 {
					return fileErr(file, periodProp.Start, "property .period must be strictly positive")
				}
			}

			if a.Expression != nil {
				initVals, err := evalConstExpr(a.Expression, pkg, getImport)
				if err == nil && len(initVals) == 1 {
					if initVals[0] < minVals[0] || initVals[0] > maxVals[0] {
						return fileErr(file, a.Expression.Start, fmt.Sprintf("slider default value %v is outside range [%v, %v]", initVals[0], minVals[0], maxVals[0]))
					}
				}
			}
		}

		// Validate bounding box property (.box)
		if boxProp := propMap["box"]; boxProp != nil {
			if len(a.Variables) != 1 {
				return fileErr(file, boxProp.Start, fmt.Sprintf("%s.box: a box can only be declared on a single-name assignment", v.Name))
			}

			if isScalarTarget(a.Expression) {
				return fileErr(file, boxProp.Start, fmt.Sprintf("cannot assign .box to scalar variable %q", v.Name))
			}

			if err := validateBoxProperty(boxProp, v.Name, pkg, file, getImport); err != nil {
				return err
			}
		}

		// Validate color property (.color)
		if colorProp := propMap["color"]; colorProp != nil {
			if isScalarTarget(a.Expression) {
				return fileErr(file, colorProp.Start, fmt.Sprintf("cannot assign .color to scalar variable %q", v.Name))
			}
			if err := validateColorProperty(colorProp, v.Name, pkg, file, getImport); err != nil {
				return err
			}
		}

		// Validate material property (.material)
		if matProp := propMap["material"]; matProp != nil {
			if isScalarTarget(a.Expression) {
				return fileErr(file, matProp.Start, fmt.Sprintf("cannot assign .material to scalar variable %q", v.Name))
			}
		}
	}

	return nil
}

func validateBoxProperty(boxProp *program.Property, varName string, pkg *program.Package, file *program.PackageFile, getImport program.GetImportFunc) error {
	op := boxProp.Expression
	if op == nil {
		return fileErr(file, boxProp.Start, fmt.Sprintf("property .box on variable %q has no expression", varName))
	}

	// Must not depend on spatial coordinates p
	if dependsOnP(op) {
		return fileErr(file, boxProp.Start, fmt.Sprintf("property .box on variable %q cannot depend on spatial coordinates p", varName))
	}

	// If standalone auto, valid
	if isAutoOperand(op) {
		return nil
	}

	// If a collection of coordinates: [lo, hi]
	if exprOp, ok := op.T.(*program.Operand_Expression); ok && exprOp.Expression.Operation == program.Operation_COLLECT {
		if len(exprOp.Expression.Operands) != 2 {
			return fileErr(file, boxProp.Start, fmt.Sprintf("bounding box for %q must contain exactly 2 corner coordinates [lo, hi], got %d", varName, len(exprOp.Expression.Operands)))
		}

		loOp := exprOp.Expression.Operands[0]
		hiOp := exprOp.Expression.Operands[1]

		// Check coordinate component length if collections
		if loExpr, ok := loOp.T.(*program.Operand_Expression); ok && loExpr.Expression.Operation == program.Operation_COLLECT {
			if len(loExpr.Expression.Operands) != 3 {
				return fileErr(file, loOp.Start, fmt.Sprintf("bounding box corner lo for %q must be a 3D coordinate [x, y, z], got %d components", varName, len(loExpr.Expression.Operands)))
			}
		}
		if hiExpr, ok := hiOp.T.(*program.Operand_Expression); ok && hiExpr.Expression.Operation == program.Operation_COLLECT {
			if len(hiExpr.Expression.Operands) != 3 {
				return fileErr(file, hiOp.Start, fmt.Sprintf("bounding box corner hi for %q must be a 3D coordinate [x, y, z], got %d components", varName, len(hiExpr.Expression.Operands)))
			}
		}

		// Constant extent evaluation if both fold
		loVals, errLo := evalConstExpr(loOp, pkg, getImport)
		hiVals, errHi := evalConstExpr(hiOp, pkg, getImport)
		if errLo == nil && errHi == nil && len(loVals) == 3 && len(hiVals) == 3 {
			axisNames := []string{"x", "y", "z"}
			for i := 0; i < 3; i++ {
				if loVals[i] >= hiVals[i] {
					return fileErr(file, boxProp.Start, fmt.Sprintf("invalid bounding box: lo.%s (%v) must be strictly less than hi.%s (%v)", axisNames[i], loVals[i], axisNames[i], hiVals[i]))
				}
			}
		}
		return nil
	}

	// Function call returning [lo, hi]
	if _, ok := op.T.(*program.Operand_FnCall); ok {
		return nil
	}

	// Bounding box expression (e.g. s2.box)
	if exprOp, ok := op.T.(*program.Operand_Expression); ok && exprOp.Expression != nil {
		if exprOp.Expression.Operation == program.Operation_BOUNDINGBOX {
			return nil
		}
	}

	// Variable reference or property reference
	if _, ok := op.T.(*program.Operand_Reference); ok {
		return nil
	}
	if _, ok := op.T.(*program.Operand_PropRef); ok {
		return nil
	}

	return fileErr(file, boxProp.Start, fmt.Sprintf("invalid value for .box property on variable %q", varName))
}

func validateColorProperty(colorProp *program.Property, varName string, pkg *program.Package, file *program.PackageFile, getImport program.GetImportFunc) error {
	op := colorProp.Expression
	if op == nil {
		return fileErr(file, colorProp.Start, fmt.Sprintf("property .color on variable %q has no expression", varName))
	}

	// If constant, verify components in [0.0, 1.0]
	vals, err := evalConstExpr(op, pkg, getImport)
	if err == nil {
		if len(vals) != 3 && len(vals) != 4 {
			return fileErr(file, colorProp.Start, fmt.Sprintf("property .color on variable %q must be a 3-element or 4-element vector, got %d components", varName, len(vals)))
		}
		for _, c := range vals {
			if c < 0.0 || c > 1.0 {
				return fileErr(file, colorProp.Start, fmt.Sprintf("color components must be in range [0.0, 1.0], got %v", c))
			}
		}
	}
	return nil
}

func validateExpression(op *program.Operand, inBoxProperty bool, file *program.PackageFile) error {
	if op == nil {
		return nil
	}

	// 1. auto keyword outside .box
	if isAutoOperand(op) && !inBoxProperty {
		return fileErr(file, op.Start, "reserved keyword \"auto\" is only permitted within .box property assignments")
	}

	// 2. Reject reading metadata properties in expressions
	if propRef, ok := op.T.(*program.Operand_PropRef); ok && propRef != nil && propRef.PropRef != nil {
		pName := propRef.PropRef.Property
		if pName != "box" && validAssignablePropNames[pName] {
			return fileErr(file, op.Start, fmt.Sprintf("reading property %q in expressions is not supported; properties are declaration metadata", pName))
		}
	}

	// Traverse children
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression != nil {
			for _, child := range t.Expression.Operands {
				if err := validateExpression(child, inBoxProperty, file); err != nil {
					return err
				}
			}
		}
	case *program.Operand_FnCall:
		if t.FnCall != nil {
			for _, arg := range t.FnCall.Arguments {
				if err := validateExpression(arg, inBoxProperty, file); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func isAutoOperand(op *program.Operand) bool {
	if op == nil {
		return false
	}
	// Check reference to auto or special Auto AST representation
	if ref, ok := op.T.(*program.Operand_Reference); ok && ref.Reference != nil {
		return ref.Reference.Prefix == "" && ref.Reference.Name == "auto"
	}
	return false
}

func dependsOnP(op *program.Operand) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		return t.Reference != nil && t.Reference.Prefix == "" && t.Reference.Name == program.PVarName
	case *program.Operand_Expression:
		if t.Expression != nil {
			for _, child := range t.Expression.Operands {
				if dependsOnP(child) {
					return true
				}
			}
		}
	case *program.Operand_FnCall:
		if t.FnCall != nil {
			for _, arg := range t.FnCall.Arguments {
				if dependsOnP(arg) {
					return true
				}
			}
		}
	}
	return false
}

func isScalarTarget(op *program.Operand) bool {
	if op == nil {
		return false
	}
	if _, ok := op.T.(*program.Operand_Value); ok {
		return true
	}
	return false
}

func isCollectionOrShapeTarget(op *program.Operand) bool {
	if op == nil {
		return false
	}
	if exprOp, ok := op.T.(*program.Operand_Expression); ok && exprOp.Expression != nil {
		if exprOp.Expression.Operation == program.Operation_COLLECT {
			return true
		}
	}
	// Call to shape constructors or similar
	if fnOp, ok := op.T.(*program.Operand_FnCall); ok && fnOp.FnCall != nil && fnOp.FnCall.Reference != nil {
		name := fnOp.FnCall.Reference.Name
		if name == "sphere" || name == "box" || name == "cylinder" || name == "torus" || name == "cone" {
			return true
		}
	}
	return false
}

func evalConstExpr(op *program.Operand, pkg *program.Package, getImport program.GetImportFunc) ([]float64, error) {
	if op == nil {
		return nil, fmt.Errorf("nil operand")
	}
	if valOp, ok := op.T.(*program.Operand_Value); ok && valOp != nil {
		return []float64{valOp.Value}, nil
	}
	return eval.ConstPkg(op, pkg, getImport)
}

func fileErr(file *program.PackageFile, loc *program.Location, msg string) error {
	path := ""
	if file != nil {
		path = file.Path
	}
	line := 0
	col := 0
	if loc != nil {
		line = int(loc.Line)
		col = int(loc.Column)
	}
	return &program.Error{
		FilePath: path,
		LineNum:  line,
		Column:   col,
		Message:  msg,
	}
}
