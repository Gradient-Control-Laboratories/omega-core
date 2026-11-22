package property

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
)

// Geometric features: name.featureSize = size marks a value as a feature the
// viewer prunes once it is too small to see, a hole or a gear's teeth, even
// though the part it belongs to is large. A part's box is the wrong measure of
// that: the part is not what disappears. The size is the feature's own, in
// mm, declared where the feature is built; 0 asks for the smallest extent of
// its box, and a box the analysis cannot bound on any axis leaves it so large
// it is never pruned.
//
// The value becomes builtin pruneBelow(value, size), the value everywhere but
// the viewer's WGSL. Pruned, a feature is nothing, which leaves a union
// unchanged and, taken away by a difference, leaves the difference unchanged:
// a bump drops out, a hole closes. Anywhere else nothing would be wrong (an
// intersection with it would vanish whole), so a feature may stand only as an
// operand of a union or as what a difference takes away, and may be returned,
// which leaves its placing to the caller (a part is a union operand of the
// scene; a function's feature, drill.hole's void, is its caller's to take
// away).
const (
	featureSizePropName = "featureSize"
	pruneBelowName      = "pruneBelow"
	featureExtentName   = "featureExtent"
)

// lowerFeatures lowers every featured value in the package, in each file and
// each function at any depth, and checks where each one is read.
func lowerFeatures(pkg *program.Package) error {
	var fnType func(ft *program.FnType) error
	fnType = func(ft *program.FnType) error {
		if ft == nil {
			return nil
		}
		for _, group := range ft.Groups {
			if group == nil {
				continue
			}
			for _, fn := range group.Functions {
				if fn == nil {
					continue
				}
				if err := lowerFeatureScope(fn.Assignments, fn.Returns); err != nil {
					return err
				}
				for _, nested := range fn.FnTypes {
					if err := fnType(nested); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	for _, file := range pkg.Files {
		if file == nil {
			continue
		}
		if err := lowerFeatureScope(file.Assignments, file.Returns); err != nil {
			return err
		}
		for _, ft := range file.FnTypes {
			if err := fnType(ft); err != nil {
				return err
			}
		}
	}
	return nil
}

// lowerFeatureScope rewrites each assignment of one scope that sets
// name.featureSize to pruneBelow(value, size), and checks every read of the
// featured names in the scope.
func lowerFeatureScope(assigns []*program.Assignment, returns []*program.Operand) error {
	featured := make(map[string]bool)
	for _, a := range assigns {
		if a == nil {
			continue
		}
		for _, v := range a.Variables {
			if v == nil {
				continue
			}
			var size *program.Operand
			kept := v.Properties[:0]
			for _, prop := range v.Properties {
				if prop != nil && prop.Name == featureSizePropName {
					size = prop.Expression
					continue
				}
				kept = append(kept, prop)
			}
			if size == nil {
				continue
			}
			if len(a.Variables) != 1 {
				return fmt.Errorf("%s.featureSize: a feature can only be declared on a single-name assignment", v.Name)
			}
			v.Properties = kept
			if isZeroLiteral(size) {
				// The smallest extent of the feature's own box.
				size = call(featureExtentName, nary(program.Operation_BOUNDINGBOX, proto.Clone(a.Expression).(*program.Operand)))
			}
			a.Expression = call(pruneBelowName, a.Expression, size)
			featured[v.Name] = true
		}
	}
	if len(featured) == 0 {
		return nil
	}
	for _, a := range assigns {
		if a == nil {
			continue
		}
		if err := checkFeatureReads(a.Expression, featured, false); err != nil {
			return err
		}
	}
	for _, r := range returns {
		if err := checkFeatureReads(r, featured, true); err != nil {
			return err
		}
	}
	return nil
}

func isZeroLiteral(op *program.Operand) bool {
	if v, ok := op.GetT().(*program.Operand_Value); ok {
		return v.Value == 0
	}
	return false
}

// checkFeatureReads walks op, prunable telling whether op stands where a
// feature may: in a union, taken away by a difference, or returned. A read of
// a featured name anywhere else is an error.
func checkFeatureReads(op *program.Operand, featured map[string]bool, prunable bool) error {
	if op == nil {
		return nil
	}
	switch t := op.GetT().(type) {
	case *program.Operand_Reference:
		if featured[t.Reference.GetName()] && !prunable {
			return featureMisplaced(t.Reference.GetName())
		}
		return nil
	case *program.Operand_Expression:
		e := t.Expression
		switch e.GetOperation() {
		case program.Operation_MIN:
			return checkEach(e.GetOperands(), featured, func(int) bool { return true })
		case program.Operation_MAX:
			return checkTakenAway(e.GetOperands(), featured)
		case program.Operation_REMAP:
			return checkEach(e.GetOperands(), featured, func(i int) bool { return i == 0 && prunable })
		case program.Operation_SELECT:
			return checkEach(e.GetOperands(), featured, func(i int) bool { return i > 0 && prunable })
		}
		return checkEach(e.GetOperands(), featured, func(int) bool { return false })
	case *program.Operand_FnCall:
		name := t.FnCall.GetReference().GetName()
		args := t.FnCall.GetArguments()
		switch {
		case name == "min":
			return checkEach(args, featured, func(int) bool { return true })
		case name == "max":
			return checkTakenAway(args, featured)
		case name == "remap":
			return checkEach(args, featured, func(i int) bool { return i == 0 && prunable })
		case strings.HasPrefix(name, "union"):
			// unionRound(a, b, r), unionChamfer, unionSharp, ...: the fields
			// are its first two arguments.
			return checkEach(args, featured, func(i int) bool { return i < 2 })
		case strings.HasPrefix(name, "difference"):
			// differenceRound(a, b, r), ...: b is taken away from a.
			return checkEach(args, featured, func(i int) bool { return i == 1 })
		}
		return checkEach(args, featured, func(int) bool { return false })
	}
	return nil
}

func checkEach(ops []*program.Operand, featured map[string]bool, prunable func(int) bool) error {
	for i, o := range ops {
		if err := checkFeatureReads(o, featured, prunable(i)); err != nil {
			return err
		}
	}
	return nil
}

// checkTakenAway checks a max's operands: one negated, neg(x) or -x, is taken
// away, so x stands where a feature may; any other is intersected.
func checkTakenAway(ops []*program.Operand, featured map[string]bool) error {
	for _, o := range ops {
		if x, ok := negated(o); ok {
			if err := checkFeatureReads(x, featured, true); err != nil {
				return err
			}
			continue
		}
		if err := checkFeatureReads(o, featured, false); err != nil {
			return err
		}
	}
	return nil
}

func negated(op *program.Operand) (*program.Operand, bool) {
	switch t := op.GetT().(type) {
	case *program.Operand_Expression:
		if t.Expression.GetOperation() == program.Operation_NEG && len(t.Expression.GetOperands()) == 1 {
			return t.Expression.GetOperands()[0], true
		}
	case *program.Operand_FnCall:
		if t.FnCall.GetReference().GetName() == "neg" && len(t.FnCall.GetArguments()) == 1 {
			return t.FnCall.GetArguments()[0], true
		}
	}
	return nil, false
}

func featureMisplaced(name string) error {
	return fmt.Errorf("%s.featureSize: a feature can stand only in a union or as what a difference takes away, where pruning it to nothing leaves the rest unchanged; %s is read elsewhere", name, name)
}

// packageDeclaresFeature reports whether any declaration in the package, at
// any depth, sets name.featureSize.
func packageDeclaresFeature(pkg *program.Package) bool {
	found := false
	_ = eachAssignments(pkg, true, func(assigns []*program.Assignment) error {
		for _, a := range assigns {
			if a == nil {
				continue
			}
			for _, v := range a.Variables {
				if v != nil && hasProp(v, featureSizePropName) {
					found = true
				}
			}
		}
		return nil
	})
	return found
}
