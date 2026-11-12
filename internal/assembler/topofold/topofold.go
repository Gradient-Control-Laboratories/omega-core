package topofold

import (
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/assembler/meshslots"
	"gradientcontrol.com/omega/core/assembler/topofold"
	"gradientcontrol.com/omega/core/program"
)

var topoPointBuiltins = map[string]bool{
	"facePoint":    true,
	"edgePoint":    true,
	"vertexPoint":  true,
	"circleCenter": true,
}

var topoNormalBuiltins = map[string]bool{
	"facePointNormal":    true,
	"edgePointNormal":    true,
	"vertexPointNormal":  true,
	"circleCenterNormal": true,
}

var meshQueryBuiltins = map[string]bool{
	"meshVolume":       true,
	"meshArea":         true,
	"meshMaxThickness": true,
}

var meshPointQueryBuiltins = map[string]bool{
	"meshThickestPoint": true,
}

var meshListQueryBuiltins = map[string]bool{
	"meshThickestPoints": true,
}

func isTopoCall(fc *program.FnCall) bool {
	if fc == nil || fc.Reference == nil || fc.Reference.Prefix != "" {
		return false
	}
	name := fc.Reference.Name
	return topoPointBuiltins[name] || topoNormalBuiltins[name] ||
		meshQueryBuiltins[name] || meshPointQueryBuiltins[name] ||
		meshListQueryBuiltins[name]
}

// HasTopoCall reports whether any package in prog contains topology or mesh-query builtin calls.
func HasTopoCall(prog *program.Program) bool {
	// Invariant: prog, prog.Primary, and prog.Dependencies entries are non-nil.
	if hasTopoCall(prog.Primary) {
		return true
	}
	for _, dep := range prog.Dependencies {
		if hasTopoCall(dep) {
			return true
		}
	}
	return false
}

// hasTopoCall returns true if the package contains any topology-builtin calls.
func hasTopoCall(pkg *program.Package) bool {
	// Invariant: pkg and its AST slice elements (Files, Assignments, FnTypes, Groups, Functions, Operands) are non-nil.
	found := false
	var walkOp func(op *program.Operand)
	walkOp = func(op *program.Operand) {
		if found {
			return
		}
		switch t := op.T.(type) {
		case *program.Operand_FnCall:
			if isTopoCall(t.FnCall) {
				found = true
				return
			}
			for _, arg := range t.FnCall.Arguments {
				walkOp(arg)
			}
		case *program.Operand_Expression:
			for _, sub := range t.Expression.Operands {
				walkOp(sub)
			}
		}
	}

	var walkAssignment func(a *program.Assignment)
	var walkFnType func(ft *program.FnType)

	walkAssignment = func(a *program.Assignment) {
		if found {
			return
		}
		walkOp(a.Expression)
		for _, v := range a.Variables {
			for _, p := range v.Properties {
				walkOp(p.Expression)
			}
		}
	}

	walkFnType = func(ft *program.FnType) {
		if found {
			return
		}
		for _, g := range ft.Groups {
			for _, fn := range g.Functions {
				for _, a := range fn.Assignments {
					walkAssignment(a)
				}
				for _, subFt := range fn.FnTypes {
					walkFnType(subFt)
				}
				for _, r := range fn.Returns {
					walkOp(r)
				}
			}
		}
	}

	for _, file := range pkg.Files {
		for _, a := range file.Assignments {
			walkAssignment(a)
		}
		for _, ft := range file.FnTypes {
			walkFnType(ft)
		}
		for _, r := range file.Returns {
			walkOp(r)
		}
		for _, r := range file.ColorReturns {
			walkOp(r)
		}
		for _, r := range file.MaterialReturns {
			walkOp(r)
		}
	}
	return found
}

// FoldProgram folds all topology and mesh-property builtin calls across prog in place
// using pre-folded constant values and the provided GeometryResolver.
func FoldProgram(prog *program.Program, importSlots []meshslots.Import, resolver topofold.GeometryResolver) error {
	// Invariant: prog, prog.Primary, prog.Dependencies entries, and resolver are non-nil,
	// and upstream constant folding has already populated ConstValues on constant expressions.
	if hasTopoCall(prog.Primary) {
		if err := foldPackage(prog.Primary, "", true, importSlots, resolver); err != nil {
			return err
		}
	}

	for _, dep := range prog.Dependencies {
		if hasTopoCall(dep) {
			if err := foldPackage(dep, dep.Path, false, importSlots, resolver); err != nil {
				return err
			}
		}
	}

	return nil
}

// foldPackage folds all topology references within a single package in place.
func foldPackage(
	pkg *program.Package,
	pkgPath string,
	isPrimary bool,
	slotTable []meshslots.Import,
	resolver topofold.GeometryResolver,
) error {
	placements, placementCounts := extractPlacements(pkg)

	folder := &topoFolder{
		bodySlots:       bodySlotsFor(pkg, pkgPath, isPrimary, slotTable),
		placements:      placements,
		placementCounts: placementCounts,
		resolver:        resolver,
		cachedMeshes:    make(map[int]topofold.AnnotatedMesh),
	}

	// Invariant: pkg.Files and its AST slice elements are non-nil.
	for _, file := range pkg.Files {
		consts, err := folder.foldAssignments(&file.Assignments, nil)
		if err != nil {
			return err
		}
		for _, ft := range file.FnTypes {
			if err := folder.foldFnType(ft, consts); err != nil {
				return err
			}
		}
		if err := folder.foldReturns(&file.Returns, consts); err != nil {
			return err
		}
		if err := folder.foldReturns(&file.ColorReturns, consts); err != nil {
			return err
		}
		if err := folder.foldReturns(&file.MaterialReturns, consts); err != nil {
			return err
		}
	}

	return nil
}

// bodySlotsFor lists a package's STEP-import slots in declaration order.
func bodySlotsFor(p *program.Package, pkgPath string, isPrimary bool, slotTable []meshslots.Import) []int {
	var out []int
	for _, f := range p.Files {
		for _, si := range f.StepImports {
			out = append(out, int(si.Index))
		}
	}
	for _, imp := range slotTable {
		if imp.Kind == "step" && (imp.PkgPath == pkgPath || (isPrimary && imp.PkgPath == p.Path)) {
			out = append(out, imp.Slot)
		}
	}
	return out
}

type topoFolder struct {
	bodySlots       []int
	placements      map[int][]float64
	placementCounts map[int]int
	resolver        topofold.GeometryResolver
	cachedMeshes    map[int]topofold.AnnotatedMesh
}

func (f *topoFolder) errorAt(loc *program.Location, msg string) error {
	if loc != nil {
		return fmt.Errorf("%s (line %d, col %d)", msg, loc.Line, loc.Column)
	}
	return fmt.Errorf("%s", msg)
}

func copyConsts(src map[string][]float64) map[string][]float64 {
	dst := make(map[string][]float64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (f *topoFolder) foldAssignments(assignments *[]*program.Assignment, outerConsts map[string][]float64) (map[string][]float64, error) {
	consts := copyConsts(outerConsts)
	out := make([]*program.Assignment, 0, len(*assignments))
	// Invariant: *assignments slice elements are non-nil.
	for _, a := range *assignments {
		if fc := topoCallOf(a.Expression); fc != nil && topoNormalBuiltins[fc.Reference.Name] {
			if len(a.Variables) != 2 {
				return nil, f.errorAt(a.Expression.GetStart(), fmt.Sprintf(
					"%s returns two values (point, normal); assign both: pt, n = %s(...)", fc.Reference.Name, fc.Reference.Name))
			}
			point, normal, err := f.evalNormalCall(fc, a.Expression, consts)
			if err != nil {
				return nil, err
			}
			a1 := constAssignment(a, a.Variables[0].Name, a.Expression, point)
			a2 := constAssignment(a, a.Variables[1].Name, a.Expression, normal)
			out = append(out, a1, a2)
			consts[a1.Variables[0].Name] = a1.Expression.ConstValues
			consts[a2.Variables[0].Name] = a2.Expression.ConstValues
			continue
		}

		if err := f.foldOperand(a.Expression, consts); err != nil {
			return nil, err
		}
		for _, v := range a.Variables {
			for _, p := range v.Properties {
				if err := f.foldOperand(p.Expression, consts); err != nil {
					return nil, err
				}
			}
		}
		if len(a.Variables) == 1 {
			if vals, ok := constValues(a.Expression, consts); ok {
				consts[a.Variables[0].Name] = vals
			} else {
				delete(consts, a.Variables[0].Name)
			}
		}
		out = append(out, a)
	}
	*assignments = out
	return consts, nil
}

func (f *topoFolder) foldFnType(ft *program.FnType, outerConsts map[string][]float64) error {
	// Invariant: ft, ft.Groups, and g.Functions elements are non-nil.
	for _, g := range ft.Groups {
		for _, fn := range g.Functions {
			paramShadowed := copyConsts(outerConsts)
			for _, p := range fn.Parameters {
				delete(paramShadowed, p.Name)
			}
			innerConsts, err := f.foldAssignments(&fn.Assignments, paramShadowed)
			if err != nil {
				return err
			}
			for _, subFt := range fn.FnTypes {
				if err := f.foldFnType(subFt, innerConsts); err != nil {
					return err
				}
			}
			if err := f.foldReturns(&fn.Returns, innerConsts); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *topoFolder) foldReturns(returns *[]*program.Operand, consts map[string][]float64) error {
	out := make([]*program.Operand, 0, len(*returns))
	for _, op := range *returns {
		if fc := topoCallOf(op); fc != nil && topoNormalBuiltins[fc.Reference.Name] {
			point, normal, err := f.evalNormalCall(fc, op, consts)
			if err != nil {
				return err
			}
			out = append(out, constVec(op, point), constVec(op, normal))
			continue
		}
		if err := f.foldOperand(op, consts); err != nil {
			return err
		}
		out = append(out, op)
	}
	*returns = out
	return nil
}

func (f *topoFolder) foldOperand(op *program.Operand, consts map[string][]float64) error {
	// Invariant: op is non-nil.
	switch t := op.T.(type) {
	case *program.Operand_FnCall:
		fc := t.FnCall
		if isTopoCall(fc) {
			if topoNormalBuiltins[fc.Reference.Name] {
				return f.errorAt(op.GetStart(), fmt.Sprintf(
					"%s returns two values (point, normal) and cannot be used inside an expression; assign them first: pt, n = %s(...)", fc.Reference.Name, fc.Reference.Name))
			}
			if meshQueryBuiltins[fc.Reference.Name] {
				v, err := f.evalMeshQueryCall(fc, op, consts)
				if err != nil {
					return err
				}
				folded := constScalar(op, v)
				op.T = folded.T
				op.ConstValues = folded.ConstValues
				return nil
			}
			if meshPointQueryBuiltins[fc.Reference.Name] {
				point, err := f.evalMeshPointQueryCall(fc, op, consts)
				if err != nil {
					return err
				}
				folded := constVec(op, point)
				op.T = folded.T
				op.ConstValues = folded.ConstValues
				return nil
			}
			if meshListQueryBuiltins[fc.Reference.Name] {
				points, err := f.evalMeshPointsQueryCall(fc, op, consts)
				if err != nil {
					return err
				}
				folded := constVecList(op, points)
				op.T = folded.T
				op.ConstValues = folded.ConstValues
				return nil
			}
			point, err := f.evalPointCall(fc, op, consts)
			if err != nil {
				return err
			}
			folded := constVec(op, point)
			op.T = folded.T
			op.ConstValues = folded.ConstValues
			return nil
		}
		for _, arg := range fc.Arguments {
			if err := f.foldOperand(arg, consts); err != nil {
				return err
			}
		}
	case *program.Operand_Expression:
		for _, sub := range t.Expression.Operands {
			if err := f.foldOperand(sub, consts); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *topoFolder) checkPlacementUnambiguous(name string, body int, xf []float64, op *program.Operand) error {
	if xf != nil || f.placementCounts[body] <= 1 {
		return nil
	}
	return f.errorAt(op.GetStart(), fmt.Sprintf(
		"%s: mesh id %d is placed by %d mesh() declarations, so the implicit placement is ambiguous — pass the instance's transform as a trailing argument",
		name, body, f.placementCounts[body]))
}

func (f *topoFolder) evalPointCall(fc *program.FnCall, op *program.Operand, consts map[string][]float64) ([3]float64, error) {
	body, mesh, args, xf, err := f.evalCommon(fc, op, consts)
	if err != nil {
		return [3]float64{}, err
	}
	if err := f.checkPlacementUnambiguous(fc.Reference.Name, body, xf, op); err != nil {
		return [3]float64{}, err
	}
	var p [3]float64
	switch fc.Reference.Name {
	case "facePoint":
		var uv []float64
		if uv, err = argVec(args, 1, 2); err == nil {
			p, err = mesh.FacePoint(argInt(args, 0), uv[0], uv[1])
		}
	case "edgePoint":
		switch len(args) {
		case 2:
			if err = wantScalars(args, 0, 1); err == nil {
				p, err = mesh.EdgePoint(argInt(args, 0), argScalar(args, 1))
			}
		case 3:
			if err = wantScalars(args, 0, 1, 2); err == nil {
				p, err = mesh.EdgePointComp(argInt(args, 0), argInt(args, 1), argScalar(args, 2))
			}
		default:
			err = fmt.Errorf("expected edgePoint(body, edge, t) or edgePoint(body, edge, component, t)")
		}
	case "vertexPoint":
		if err = argCount(args, 1); err == nil {
			p, err = mesh.VertexPoint(argInt(args, 0))
		}
	case "circleCenter":
		if err = argCount(args, 1); err == nil {
			p, err = mesh.CircleCenter(argInt(args, 0))
		}
	}
	if err != nil {
		return [3]float64{}, f.errorAt(op.GetStart(), fmt.Sprintf("%s: %v", fc.Reference.Name, err))
	}
	return f.placePointX(body, xf, p), nil
}

func (f *topoFolder) evalNormalCall(fc *program.FnCall, op *program.Operand, consts map[string][]float64) (point, normal [3]float64, err error) {
	body, mesh, args, xf, err := f.evalCommon(fc, op, consts)
	if err != nil {
		return
	}
	if err = f.checkPlacementUnambiguous(fc.Reference.Name, body, xf, op); err != nil {
		return
	}
	switch fc.Reference.Name {
	case "facePointNormal":
		var uv []float64
		if uv, err = argVec(args, 1, 2); err == nil {
			point, normal, err = mesh.FacePointNormal(argInt(args, 0), uv[0], uv[1])
		}
	case "edgePointNormal":
		switch len(args) {
		case 2:
			if err = wantScalars(args, 0, 1); err == nil {
				point, normal, err = mesh.EdgePointNormal(argInt(args, 0), argScalar(args, 1))
			}
		case 3:
			if err = wantScalars(args, 0, 1, 2); err == nil {
				point, normal, err = mesh.EdgePointNormalComp(argInt(args, 0), argInt(args, 1), argScalar(args, 2))
			}
		default:
			err = fmt.Errorf("expected edgePointNormal(body, edge, t) or edgePointNormal(body, edge, component, t)")
		}
	case "vertexPointNormal":
		if err = argCount(args, 1); err == nil {
			point, normal, err = mesh.VertexPointNormal(argInt(args, 0))
		}
	case "circleCenterNormal":
		if err = argCount(args, 1); err == nil {
			point, normal, err = mesh.CircleCenterNormal(argInt(args, 0))
		}
	}
	if err != nil {
		err = f.errorAt(op.GetStart(), fmt.Sprintf("%s: %v", fc.Reference.Name, err))
		return
	}
	point = f.placePointX(body, xf, point)
	normal = f.placeNormalX(body, xf, normal)
	return
}

// constValues extracts pre-folded constant values or literal AST constants from op.
// Upstream constant evaluation (eval.Fold in assembler) populates ConstValues on
// compile-time constant expressions before topofold executes; literal values,
// COLLECT lists, UI markers, and bindings folded earlier in the pass are read
// directly from the AST without invoking eval.
func constValues(op *program.Operand, consts map[string][]float64) ([]float64, bool) {
	// Invariant: op is non-nil.
	if len(op.ConstValues) > 0 {
		return op.ConstValues, true
	}
	switch t := op.T.(type) {
	case *program.Operand_Value:
		return []float64{t.Value}, true
	case *program.Operand_Reference:
		if t.Reference != nil && t.Reference.Prefix == "" && consts != nil {
			vals, ok := consts[t.Reference.Name]
			return vals, ok && len(vals) > 0
		}
	case *program.Operand_Expression:
		if t.Expression != nil && t.Expression.Operation == program.Operation_COLLECT {
			var out []float64
			for _, elem := range t.Expression.Operands {
				vals, ok := constValues(elem, consts)
				if !ok {
					return nil, false
				}
				out = append(out, vals...)
			}
			return out, len(out) > 0
		}
		if t.Expression != nil && t.Expression.Operation == program.Operation_NEG && len(t.Expression.Operands) == 1 {
			vals, ok := constValues(t.Expression.Operands[0], consts)
			if !ok {
				return nil, false
			}
			neg := make([]float64, len(vals))
			for i, v := range vals {
				neg[i] = -v
			}
			return neg, true
		}
	case *program.Operand_FnCall:
		fc := t.FnCall
		if fc != nil && fc.Reference != nil && fc.Reference.Prefix == "" && len(fc.Arguments) == 1 {
			switch fc.Reference.Name {
			case "uiPoint", "uiTransform", "uiColor", "uiDirection":
				return constValues(fc.Arguments[0], consts)
			}
		}
	}
	return nil, false
}

func (f *topoFolder) resolveMeshSlot(name string, argOp *program.Operand, consts map[string][]float64) (int, error) {
	vals, ok := constValues(argOp, consts)
	if !ok {
		return 0, f.errorAt(argOp.GetStart(), fmt.Sprintf("%s: the mesh id must be a constant", name))
	}
	if len(vals) != 1 {
		return 0, f.errorAt(argOp.GetStart(), fmt.Sprintf("%s: the mesh id must be a scalar", name))
	}
	return int(math.Round(vals[0])), nil
}

func (f *topoFolder) evalMeshQueryCall(fc *program.FnCall, op *program.Operand, consts map[string][]float64) (float64, error) {
	if len(fc.Arguments) != 1 {
		return 0, f.errorAt(op.GetStart(), fmt.Sprintf("%s: expected a single mesh id", fc.Reference.Name))
	}
	slot, err := f.resolveMeshSlot(fc.Reference.Name, fc.Arguments[0], consts)
	if err != nil {
		return 0, err
	}
	switch fc.Reference.Name {
	case "meshVolume", "meshArea":
		geom, ok := f.resolver.GetMesh(slot)
		if !ok || geom == nil {
			return 0, f.errorAt(op.GetStart(), fmt.Sprintf(
				"%s: id %d is not an imported mesh (missing importStep/importMesh?)", fc.Reference.Name, slot))
		}
		if fc.Reference.Name == "meshArea" {
			return geom.Area(), nil
		}
		return geom.Volume(), nil
	case "meshMaxThickness":
		am, err := f.annotatedMesh(slot, fc, op)
		if err != nil {
			return 0, err
		}
		v, _, err := am.MedialMax()
		if err != nil {
			return 0, f.errorAt(op.GetStart(), fmt.Sprintf("%s: %v", fc.Reference.Name, err))
		}
		return v, nil
	}
	return 0, f.errorAt(op.GetStart(), fmt.Sprintf("unhandled mesh query %s", fc.Reference.Name))
}

func (f *topoFolder) evalMeshPointQueryCall(fc *program.FnCall, op *program.Operand, consts map[string][]float64) ([3]float64, error) {
	if len(fc.Arguments) < 1 || len(fc.Arguments) > 2 {
		return [3]float64{}, f.errorAt(op.GetStart(), fmt.Sprintf("%s: expected (id [, transform])", fc.Reference.Name))
	}
	slot, err := f.resolveMeshSlot(fc.Reference.Name, fc.Arguments[0], consts)
	if err != nil {
		return [3]float64{}, err
	}
	var xf []float64
	if len(fc.Arguments) == 2 {
		vals, err := f.constArg(fc.Reference.Name, 2, fc.Arguments[1], consts)
		if err != nil {
			return [3]float64{}, err
		}
		if len(vals) != 16 {
			return [3]float64{}, f.errorAt(fc.Arguments[1].GetStart(), fmt.Sprintf(
				"%s: the transform must be a 16-value column-major matrix", fc.Reference.Name))
		}
		xf = vals
	}
	if err := f.checkPlacementUnambiguous(fc.Reference.Name, slot, xf, op); err != nil {
		return [3]float64{}, err
	}
	am, err := f.annotatedMesh(slot, fc, op)
	if err != nil {
		return [3]float64{}, err
	}
	_, at, err := am.MedialMax()
	if err != nil {
		return [3]float64{}, f.errorAt(op.GetStart(), fmt.Sprintf("%s: %v", fc.Reference.Name, err))
	}
	return f.placePointX(slot, xf, at), nil
}

func (f *topoFolder) constArg(name string, pos int, argOp *program.Operand, consts map[string][]float64) ([]float64, error) {
	vals, ok := constValues(argOp, consts)
	if !ok {
		return nil, f.errorAt(argOp.GetStart(), fmt.Sprintf("%s: argument %d must be a constant", name, pos))
	}
	return vals, nil
}

func (f *topoFolder) evalMeshPointsQueryCall(fc *program.FnCall, op *program.Operand, consts map[string][]float64) ([][3]float64, error) {
	if len(fc.Arguments) < 1 || len(fc.Arguments) > 5 {
		return nil, f.errorAt(op.GetStart(), fmt.Sprintf(
			"%s: expected (id [, transform] [, count]) or (id [, transform], axis, maxAngleDeg [, count])", fc.Reference.Name))
	}
	slot, err := f.resolveMeshSlot(fc.Reference.Name, fc.Arguments[0], consts)
	if err != nil {
		return nil, err
	}
	rest := make([][]float64, len(fc.Arguments)-1)
	for i, argOp := range fc.Arguments[1:] {
		vals, err := f.constArg(fc.Reference.Name, i+2, argOp, consts)
		if err != nil {
			return nil, err
		}
		rest[i] = vals
	}
	var xf []float64
	base := 1
	if len(rest) > 0 && len(rest[0]) == 16 {
		xf = rest[0]
		rest = rest[1:]
		base = 2
	}
	if err := f.checkPlacementUnambiguous(fc.Reference.Name, slot, xf, op); err != nil {
		return nil, err
	}
	keep := func(_, _ [3]float64) bool { return true }
	count := 4
	countArg := -1
	switch len(rest) {
	case 0:
	case 1:
		countArg = 0
	case 2, 3:
		if len(rest[0]) != 3 {
			return nil, f.errorAt(fc.Arguments[base].GetStart(), fmt.Sprintf("%s: axis must be an [x, y, z] direction", fc.Reference.Name))
		}
		ax, ay, az := rest[0][0], rest[0][1], rest[0][2]
		alen := math.Sqrt(ax*ax + ay*ay + az*az)
		if alen == 0 {
			return nil, f.errorAt(fc.Arguments[base].GetStart(), fmt.Sprintf("%s: axis must be nonzero", fc.Reference.Name))
		}
		ax, ay, az = ax/alen, ay/alen, az/alen
		if len(rest[1]) != 1 {
			return nil, f.errorAt(fc.Arguments[base+1].GetStart(), fmt.Sprintf("%s: maxAngleDeg must be a scalar", fc.Reference.Name))
		}
		threshold := -math.Cos(rest[1][0] * math.Pi / 180)
		keep = func(_, n [3]float64) bool {
			n = f.placeNormalX(slot, xf, n)
			d := n[0]*ax + n[1]*ay + n[2]*az
			return d < threshold*math.Sqrt(n[0]*n[0]+n[1]*n[1]+n[2]*n[2])
		}
		if len(rest) == 3 {
			countArg = 2
		}
	default:
		return nil, f.errorAt(op.GetStart(), fmt.Sprintf(
			"%s: expected (id [, transform] [, count]) or (id [, transform], axis, maxAngleDeg [, count])", fc.Reference.Name))
	}
	if countArg >= 0 {
		if len(rest[countArg]) != 1 || rest[countArg][0] < 1 {
			return nil, f.errorAt(fc.Arguments[base+countArg].GetStart(), fmt.Sprintf("%s: count must be a positive scalar", fc.Reference.Name))
		}
		count = int(math.Round(rest[countArg][0]))
	}
	am, err := f.annotatedMesh(slot, fc, op)
	if err != nil {
		return nil, err
	}
	pts, err := am.MedialHotSpots(keep, count)
	if err != nil {
		return nil, f.errorAt(op.GetStart(), fmt.Sprintf("%s: %v", fc.Reference.Name, err))
	}
	for i := range pts {
		pts[i] = f.placePointX(slot, xf, pts[i])
	}
	return pts, nil
}

func (f *topoFolder) evalCommon(fc *program.FnCall, op *program.Operand, consts map[string][]float64) (int, topofold.AnnotatedMesh, [][]float64, []float64, error) {
	if len(fc.Arguments) < 2 {
		return 0, nil, nil, nil, f.errorAt(op.GetStart(), fmt.Sprintf("%s: expected at least a body and an entity id", fc.Reference.Name))
	}
	args := make([][]float64, len(fc.Arguments))
	for i, argOp := range fc.Arguments {
		vals, err := f.constArg(fc.Reference.Name, i+1, argOp, consts)
		if err != nil {
			return 0, nil, nil, nil, err
		}
		args[i] = vals
	}
	if len(args[0]) != 1 {
		return 0, nil, nil, nil, f.errorAt(fc.Arguments[0].GetStart(), fmt.Sprintf("%s: body must be a scalar index", fc.Reference.Name))
	}
	var xf []float64
	if last := len(args) - 1; last >= 1 && len(args[last]) == 16 {
		xf = args[last]
		args = args[:last]
	}
	body := int(math.Round(args[0][0]))
	slot := body
	if len(f.bodySlots) > 0 {
		if body < 0 || body >= len(f.bodySlots) {
			return 0, nil, nil, nil, f.errorAt(fc.Arguments[0].GetStart(), fmt.Sprintf(
				"%s: body %d out of range: this document declares %d STEP import(s)", fc.Reference.Name, body, len(f.bodySlots)))
		}
		slot = f.bodySlots[body]
	}
	mesh, err := f.annotatedMesh(slot, fc, op)
	if err != nil {
		return 0, nil, nil, nil, err
	}
	return slot, mesh, args[1:], xf, nil
}

func (f *topoFolder) annotatedMesh(body int, fc *program.FnCall, op *program.Operand) (topofold.AnnotatedMesh, error) {
	if m, ok := f.cachedMeshes[body]; ok {
		return m, nil
	}
	am, ok, err := f.resolver.GetAnnotatedMesh(body)
	if err != nil {
		return nil, f.errorAt(op.GetStart(), fmt.Sprintf("%s: body %d annotated mesh resolution failed: %v", fc.Reference.Name, body, err))
	}
	if !ok || am == nil {
		return nil, f.errorAt(op.GetStart(), fmt.Sprintf("%s: body %d is not an imported mesh (missing importStep?)", fc.Reference.Name, body))
	}
	f.cachedMeshes[body] = am
	return am, nil
}

func (f *topoFolder) placePoint(body int, p [3]float64) [3]float64 {
	m := f.placements[body]
	if len(m) != 16 {
		return p
	}
	return applyMat4Col(m, p)
}

func (f *topoFolder) placePointX(body int, xf []float64, p [3]float64) [3]float64 {
	if xf != nil {
		return applyMat4Col(xf, p)
	}
	return f.placePoint(body, p)
}

func (f *topoFolder) placeNormal(body int, n [3]float64) [3]float64 {
	m := f.placements[body]
	if len(m) != 16 {
		return n
	}
	return applyNormalMat(m, n)
}

func (f *topoFolder) placeNormalX(body int, xf []float64, n [3]float64) [3]float64 {
	if xf != nil {
		return applyNormalMat(xf, n)
	}
	return f.placeNormal(body, n)
}

func applyNormalMat(m []float64, n [3]float64) [3]float64 {
	a00, a01, a02 := m[0], m[4], m[8]
	a10, a11, a12 := m[1], m[5], m[9]
	a20, a21, a22 := m[2], m[6], m[10]

	c00 := a11*a22 - a12*a21
	c01 := a12*a20 - a10*a22
	c02 := a10*a21 - a11*a20
	c10 := a02*a21 - a01*a22
	c11 := a00*a22 - a02*a20
	c12 := a01*a20 - a00*a21
	c20 := a01*a12 - a02*a11
	c21 := a02*a10 - a00*a12
	c22 := a00*a11 - a01*a10
	det := a00*c00 + a01*c01 + a02*c02
	if det == 0 {
		return n
	}
	out := [3]float64{
		c00*n[0] + c01*n[1] + c02*n[2],
		c10*n[0] + c11*n[1] + c12*n[2],
		c20*n[0] + c21*n[1] + c22*n[2],
	}
	len2 := out[0]*out[0] + out[1]*out[1] + out[2]*out[2]
	if len2 == 0 {
		return n
	}
	inv := 1 / math.Sqrt(len2)
	return [3]float64{out[0] * inv, out[1] * inv, out[2] * inv}
}

func applyMat4Col(m []float64, p [3]float64) [3]float64 {
	return [3]float64{
		m[0]*p[0] + m[4]*p[1] + m[8]*p[2] + m[12],
		m[1]*p[0] + m[5]*p[1] + m[9]*p[2] + m[13],
		m[2]*p[0] + m[6]*p[1] + m[10]*p[2] + m[14],
	}
}

func topoCallOf(op *program.Operand) *program.FnCall {
	// Invariant: op is non-nil.
	if t, ok := op.T.(*program.Operand_FnCall); ok {
		if isTopoCall(t.FnCall) {
			return t.FnCall
		}
	}
	return nil
}

func constAssignment(from *program.Assignment, name string, callOp *program.Operand, v [3]float64) *program.Assignment {
	vec := constVec(callOp, v)
	return &program.Assignment{
		Start:      from.GetStart(),
		End:        from.GetEnd(),
		Variables:  []*program.Variable{{Name: name, ConstValues: vec.ConstValues}},
		Expression: vec,
	}
}

func constScalar(from *program.Operand, v float64) *program.Operand {
	return &program.Operand{
		Start:       from.GetStart(),
		End:         from.GetEnd(),
		ConstValues: []float64{v},
		T:           &program.Operand_Value{Value: v},
	}
}

func constVec(from *program.Operand, v [3]float64) *program.Operand {
	ops := make([]*program.Operand, 3)
	for i, val := range v {
		ops[i] = &program.Operand{
			Start:       from.GetStart(),
			End:         from.GetEnd(),
			ConstValues: []float64{val},
			T:           &program.Operand_Value{Value: val},
		}
	}
	return &program.Operand{
		Start:       from.GetStart(),
		End:         from.GetEnd(),
		ConstValues: []float64{v[0], v[1], v[2]},
		T: &program.Operand_Expression{Expression: &program.Expression{
			Operation: program.Operation_COLLECT,
			Operands:  ops,
		}},
	}
}

func constVecList(from *program.Operand, pts [][3]float64) *program.Operand {
	ops := make([]*program.Operand, len(pts))
	flat := make([]float64, 0, len(pts)*3)
	for i, p := range pts {
		ops[i] = constVec(from, p)
		flat = append(flat, p[0], p[1], p[2])
	}
	return &program.Operand{
		Start:       from.GetStart(),
		End:         from.GetEnd(),
		ConstValues: flat,
		T: &program.Operand_Expression{Expression: &program.Expression{
			Operation: program.Operation_COLLECT,
			Operands:  ops,
		}},
	}
}

func argInt(args [][]float64, i int) int {
	return int(math.Round(args[i][0]))
}

func argScalar(args [][]float64, i int) float64 {
	return args[i][0]
}

func wantScalars(args [][]float64, idx ...int) error {
	for _, i := range idx {
		if len(args[i]) != 1 {
			return fmt.Errorf("argument %d must be a scalar", i+2)
		}
	}
	return nil
}

func argVec(args [][]float64, i, wantLen int) ([]float64, error) {
	if len(args) != i+1 {
		return nil, fmt.Errorf("expected %d arguments", i+2)
	}
	for j := 0; j < i; j++ {
		if len(args[j]) != 1 {
			return nil, fmt.Errorf("argument %d must be a scalar", j+2)
		}
	}
	if len(args[i]) != wantLen {
		return nil, fmt.Errorf("argument %d must be a vector of length %d", i+2, wantLen)
	}
	return args[i], nil
}

func argCount(args [][]float64, want int) error {
	if len(args) != want {
		return fmt.Errorf("expected %d arguments", want+1)
	}
	if err := wantScalars(args, 0); err != nil {
		return err
	}
	return nil
}

func extractPlacements(pkg *program.Package) (map[int][]float64, map[int]int) {
	placements := make(map[int][]float64)
	counts := make(map[int]int)

	var walkOperand func(op *program.Operand, consts map[string][]float64)
	walkOperand = func(op *program.Operand, consts map[string][]float64) {
		switch t := op.T.(type) {
		case *program.Operand_FnCall:
			fc := t.FnCall
			if fc != nil && fc.Reference != nil && fc.Reference.Prefix == "" && fc.Reference.Name == "mesh" {
				index, mat, err := parseMeshCallMatrix(fc, consts)
				if err == nil {
					placements[index] = mat
					counts[index]++
				}
			}
			for _, arg := range fc.Arguments {
				walkOperand(arg, consts)
			}
		case *program.Operand_Expression:
			for _, sub := range t.Expression.Operands {
				walkOperand(sub, consts)
			}
		}
	}

	var walkFnType func(ft *program.FnType, outerConsts map[string][]float64)
	walkFnType = func(ft *program.FnType, outerConsts map[string][]float64) {
		for _, g := range ft.Groups {
			for _, fn := range g.Functions {
				fnConsts := copyConsts(outerConsts)
				for _, p := range fn.Parameters {
					delete(fnConsts, p.Name)
				}
				for _, a := range fn.Assignments {
					walkOperand(a.Expression, fnConsts)
					if len(a.Variables) == 1 {
						if vals, ok := constValues(a.Expression, fnConsts); ok {
							fnConsts[a.Variables[0].Name] = vals
						} else {
							delete(fnConsts, a.Variables[0].Name)
						}
					}
				}
				for _, nested := range fn.FnTypes {
					walkFnType(nested, fnConsts)
				}
				for _, r := range fn.Returns {
					walkOperand(r, fnConsts)
				}
			}
		}
	}

	fileConsts := make(map[string][]float64)
	for _, file := range pkg.Files {
		for _, a := range file.Assignments {
			walkOperand(a.Expression, fileConsts)
			if len(a.Variables) == 1 {
				if vals, ok := constValues(a.Expression, fileConsts); ok {
					fileConsts[a.Variables[0].Name] = vals
				}
			}
		}
	}
	for _, file := range pkg.Files {
		for _, r := range file.Returns {
			walkOperand(r, fileConsts)
		}
		for _, ft := range file.FnTypes {
			walkFnType(ft, fileConsts)
		}
	}

	return placements, counts
}

func parseMeshCallMatrix(fc *program.FnCall, consts map[string][]float64) (int, []float64, error) {
	args := fc.Arguments
	if len(args) == 0 {
		return 0, nil, fmt.Errorf("mesh() requires at least one argument")
	}

	idx, ok := constValues(args[0], consts)
	if !ok || len(idx) != 1 {
		return 0, nil, fmt.Errorf("mesh() first argument is not a scalar constant")
	}
	index := int(math.Round(idx[0]))

	if len(args) <= 1 {
		return index, []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}, nil
	}

	v, ok := constValues(args[1], consts)
	if !ok {
		return index, nil, fmt.Errorf("mesh() second argument is not a constant")
	}

	switch {
	case len(v) == 16:
		return index, append([]float64(nil), v...), nil
	case len(v) == 3 && len(args) == 2:
		return index, []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, v[0], v[1], v[2], 1}, nil
	case len(v) == 3 && len(args) == 3:
		r, ok := constValues(args[2], consts)
		if !ok {
			return index, nil, fmt.Errorf("mesh() third argument is not a constant")
		}
		if len(r) == 16 {
			mat := append([]float64(nil), r...)
			mat[12] = v[0]
			mat[13] = v[1]
			mat[14] = v[2]
			return index, mat, nil
		}
		return index, []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, v[0], v[1], v[2], 1}, nil
	}

	return index, []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}, nil
}
