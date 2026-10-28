package topofold

import (
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/meshfold"
	"gradientcontrol.com/omega/core/internal/meshslots"
	"gradientcontrol.com/omega/core/program"
)

// MeshGeometry provides basic scalar measurements for a mesh.
type MeshGeometry interface {
	Area() float64
	Volume() float64
}

// AnnotatedMesh provides B-rep topological queries and medial axis measurements.
type AnnotatedMesh interface {
	FacePoint(face int, u, v float64) ([3]float64, error)
	EdgePoint(edge int, t float64) ([3]float64, error)
	EdgePointComp(edge, comp int, t float64) ([3]float64, error)
	VertexPoint(vertex int) ([3]float64, error)
	CircleCenter(edge int) ([3]float64, error)
	FacePointNormal(face int, u, v float64) ([3]float64, [3]float64, error)
	EdgePointNormal(edge int, t float64) ([3]float64, [3]float64, error)
	EdgePointNormalComp(edge, comp int, t float64) ([3]float64, [3]float64, error)
	VertexPointNormal(vertex int) ([3]float64, [3]float64, error)
	CircleCenterNormal(edge int) ([3]float64, [3]float64, error)
	MedialMax() (float64, [3]float64, error)
	MedialHotSpots(keep func([3]float64, [3]float64) bool, count int) ([][3]float64, error)
}

// GeometryResolver resolves mesh properties and annotated topology for body slots.
type GeometryResolver interface {
	GetMesh(slot int) (MeshGeometry, bool)
	GetAnnotatedMesh(slot int) (AnnotatedMesh, bool, error)
}

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

// HasTopoCall returns true if the package contains any topology-builtin calls.
func HasTopoCall(pkg *program.Package) bool {
	if pkg == nil {
		return false
	}
	found := false
	var walkOp func(op *program.Operand)
	walkOp = func(op *program.Operand) {
		if op == nil || found {
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
		if a == nil || found {
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
		if ft == nil || found {
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
	}
	return found
}

// FoldProgram folds all topology and mesh-property builtin calls across the entire
// program graph using the provided GeometryResolver.
func FoldProgram(prog *program.Program, resolver GeometryResolver) (*program.Program, error) {
	if prog == nil || resolver == nil {
		return prog, nil
	}

	// Fast path: check if any package has topology calls.
	anyTopo := false
	if prog.Primary != nil && HasTopoCall(prog.Primary) {
		anyTopo = true
	}
	if !anyTopo {
		for _, dep := range prog.Dependencies {
			if dep != nil && HasTopoCall(dep) {
				anyTopo = true
				break
			}
		}
	}
	if !anyTopo {
		return prog, nil
	}

	// Assign mesh slots for the program graph.
	slotTable, slotResolver := meshslots.Assign(prog.Primary, prog.GetImport)

	// Clone & slot-fold the program graph so that mesh slot parameters are constants.
	foldedProg, err := meshfold.FoldProgramMeshIndices(prog, slotResolver)
	if err != nil {
		return nil, fmt.Errorf("topofold: mesh slots folding failed: %w", err)
	}

	gi := foldedProg.GetImport
	foldingGI := meshfold.FoldingGetImport(gi, slotResolver)

	// Fold Primary package if it has topo calls.
	if foldedProg.Primary != nil && HasTopoCall(foldedProg.Primary) {
		if err := foldPackage(foldedProg.Primary, "", slotTable, slotResolver, foldingGI, resolver); err != nil {
			return nil, err
		}
	}

	// Fold dependent packages.
	for _, dep := range foldedProg.Dependencies {
		if dep != nil && HasTopoCall(dep) {
			if err := foldPackage(dep, dep.Path, slotTable, slotResolver, foldingGI, resolver); err != nil {
				return nil, err
			}
		}
	}

	return foldedProg, nil
}

// foldPackage folds all topology references within a single package in place.
func foldPackage(
	pkg *program.Package,
	pkgPath string,
	slotTable []meshslots.Import,
	slotResolver meshslots.Resolver,
	gi program.GetImportFunc,
	resolver GeometryResolver,
) error {
	placements, placementCounts, err := extractPlacements(pkg, gi)
	if err != nil {
		return fmt.Errorf("failed to resolve mesh placements: %w", err)
	}

	folder := &topoFolder{
		pkg:             pkg,
		pkgPath:         pkgPath,
		gi:              gi,
		slots:           slotResolver,
		bodySlots:       bodySlotsFor(pkg, pkgPath, slotTable),
		placements:      placements,
		placementCounts: placementCounts,
		resolver:        resolver,
		cachedMeshes:    make(map[int]AnnotatedMesh),
	}

	for _, file := range pkg.Files {
		if file == nil {
			continue
		}
		if err := folder.foldAssignments(&file.Assignments, nil, nil); err != nil {
			return err
		}
		for _, ft := range file.FnTypes {
			if err := folder.foldFnType(ft, nil, nil); err != nil {
				return err
			}
		}
		if err := folder.foldReturns(&file.Returns, nil, nil); err != nil {
			return err
		}
	}

	return nil
}

// bodySlotsFor lists a package's STEP-import slots in declaration order.
func bodySlotsFor(p *program.Package, pkgPath string, slotTable []meshslots.Import) []int {
	var out []int
	for _, f := range p.Files {
		for _, si := range f.StepImports {
			out = append(out, int(si.Index))
		}
	}
	for _, imp := range slotTable {
		if imp.PkgPath == pkgPath && imp.Kind == "step" {
			out = append(out, imp.Slot)
		}
	}
	return out
}

type topoFolder struct {
	pkg             *program.Package
	pkgPath         string
	gi              program.GetImportFunc
	slots           meshslots.Resolver
	bodySlots       []int
	placements      map[int][]float64
	placementCounts map[int]int
	resolver        GeometryResolver
	cachedMeshes    map[int]AnnotatedMesh
}

func (f *topoFolder) errorAt(loc *program.Location, msg string) error {
	if loc != nil {
		return fmt.Errorf("%s (line %d, col %d)", msg, loc.Line, loc.Column)
	}
	return fmt.Errorf("%s", msg)
}

func (f *topoFolder) foldAssignments(assignments *[]*program.Assignment, outerAssignments []*program.Assignment, outerFnTypes []*program.FnType) error {
	visibleAssignments := append([]*program.Assignment(nil), outerAssignments...)
	visibleFnTypes := outerFnTypes
	out := make([]*program.Assignment, 0, len(*assignments))
	for _, a := range *assignments {
		if a == nil {
			continue
		}
		if fc := topoCallOf(a.Expression); fc != nil && topoNormalBuiltins[fc.Reference.Name] {
			if len(a.Variables) != 2 {
				return f.errorAt(a.Expression.GetStart(), fmt.Sprintf(
					"%s returns two values (point, normal); assign both: pt, n = %s(...)", fc.Reference.Name, fc.Reference.Name))
			}
			point, normal, err := f.evalNormalCall(fc, a.Expression, visibleAssignments, visibleFnTypes)
			if err != nil {
				return err
			}
			a1 := constAssignment(a, a.Variables[0].Name, a.Expression, point)
			a2 := constAssignment(a, a.Variables[1].Name, a.Expression, normal)
			out = append(out, a1, a2)
			visibleAssignments = append(visibleAssignments, a1, a2)
			continue
		}

		if err := f.foldOperand(a.Expression, visibleAssignments, visibleFnTypes); err != nil {
			return err
		}
		for _, v := range a.Variables {
			for _, p := range v.Properties {
				if err := f.foldOperand(p.Expression, visibleAssignments, visibleFnTypes); err != nil {
					return err
				}
			}
		}
		out = append(out, a)
		visibleAssignments = append(visibleAssignments, a)
	}
	*assignments = out
	return nil
}

func (f *topoFolder) foldFnType(ft *program.FnType, visibleAssignments []*program.Assignment, visibleFnTypes []*program.FnType) error {
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
			if err := f.foldAssignments(&fn.Assignments, visibleAssignments, visibleFnTypes); err != nil {
				return err
			}
			innerAssignments := append(append([]*program.Assignment(nil), visibleAssignments...), fn.Assignments...)
			innerFnTypes := append(append([]*program.FnType(nil), visibleFnTypes...), fn.FnTypes...)
			for _, subFt := range fn.FnTypes {
				if err := f.foldFnType(subFt, innerAssignments, innerFnTypes); err != nil {
					return err
				}
			}
			if err := f.foldReturns(&fn.Returns, innerAssignments, innerFnTypes); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *topoFolder) foldReturns(returns *[]*program.Operand, visibleAssignments []*program.Assignment, visibleFnTypes []*program.FnType) error {
	out := make([]*program.Operand, 0, len(*returns))
	for _, op := range *returns {
		if fc := topoCallOf(op); fc != nil && topoNormalBuiltins[fc.Reference.Name] {
			point, normal, err := f.evalNormalCall(fc, op, visibleAssignments, visibleFnTypes)
			if err != nil {
				return err
			}
			out = append(out, constVec(op, point), constVec(op, normal))
			continue
		}
		if err := f.foldOperand(op, visibleAssignments, visibleFnTypes); err != nil {
			return err
		}
		out = append(out, op)
	}
	*returns = out
	return nil
}

func (f *topoFolder) foldOperand(op *program.Operand, visibleAssignments []*program.Assignment, visibleFnTypes []*program.FnType) error {
	if op == nil {
		return nil
	}
	switch t := op.T.(type) {
	case *program.Operand_FnCall:
		fc := t.FnCall
		if isTopoCall(fc) {
			if topoNormalBuiltins[fc.Reference.Name] {
				return f.errorAt(op.GetStart(), fmt.Sprintf(
					"%s returns two values (point, normal) and cannot be used inside an expression; assign them first: pt, n = %s(...)", fc.Reference.Name, fc.Reference.Name))
			}
			if meshQueryBuiltins[fc.Reference.Name] {
				v, err := f.evalMeshQueryCall(fc, op, visibleAssignments, visibleFnTypes)
				if err != nil {
					return err
				}
				op.T = constScalar(op, v).T
				return nil
			}
			if meshPointQueryBuiltins[fc.Reference.Name] {
				point, err := f.evalMeshPointQueryCall(fc, op, visibleAssignments, visibleFnTypes)
				if err != nil {
					return err
				}
				op.T = constVec(op, point).T
				return nil
			}
			if meshListQueryBuiltins[fc.Reference.Name] {
				points, err := f.evalMeshPointsQueryCall(fc, op, visibleAssignments, visibleFnTypes)
				if err != nil {
					return err
				}
				op.T = constVecList(op, points).T
				return nil
			}
			point, err := f.evalPointCall(fc, op, visibleAssignments, visibleFnTypes)
			if err != nil {
				return err
			}
			op.T = constVec(op, point).T
			return nil
		}
		for _, arg := range fc.Arguments {
			if err := f.foldOperand(arg, visibleAssignments, visibleFnTypes); err != nil {
				return err
			}
		}
	case *program.Operand_Expression:
		for _, sub := range t.Expression.Operands {
			if err := f.foldOperand(sub, visibleAssignments, visibleFnTypes); err != nil {
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

func (f *topoFolder) evalPointCall(fc *program.FnCall, op *program.Operand, visibleAssignments []*program.Assignment, visibleFnTypes []*program.FnType) ([3]float64, error) {
	body, mesh, args, xf, err := f.evalCommon(fc, op, visibleAssignments, visibleFnTypes)
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

func (f *topoFolder) evalNormalCall(fc *program.FnCall, op *program.Operand, visibleAssignments []*program.Assignment, visibleFnTypes []*program.FnType) (point, normal [3]float64, err error) {
	body, mesh, args, xf, err := f.evalCommon(fc, op, visibleAssignments, visibleFnTypes)
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

func (f *topoFolder) resolveMeshSlot(name string, argOp *program.Operand, visibleAssignments []*program.Assignment, visibleFnTypes []*program.FnType) (int, error) {
	if len(argOp.ConstValues) == 1 {
		return int(math.Round(argOp.ConstValues[0])), nil
	}
	vals, err := eval.Const(argOp, visibleAssignments, visibleFnTypes)
	if err != nil {
		vals, err = eval.ConstPkgSlotsAt(argOp, f.pkg, f.pkgPath, f.gi, f.slots)
	}
	if err != nil {
		return 0, f.errorAt(argOp.GetStart(), fmt.Sprintf(
			"%s: the mesh id must be a constant: %v", name, err))
	}
	if len(vals) != 1 {
		return 0, f.errorAt(argOp.GetStart(), fmt.Sprintf("%s: the mesh id must be a scalar", name))
	}
	return int(math.Round(vals[0])), nil
}

func (f *topoFolder) evalMeshQueryCall(fc *program.FnCall, op *program.Operand, visibleAssignments []*program.Assignment, visibleFnTypes []*program.FnType) (float64, error) {
	if len(fc.Arguments) != 1 {
		return 0, f.errorAt(op.GetStart(), fmt.Sprintf("%s: expected a single mesh id", fc.Reference.Name))
	}
	slot, err := f.resolveMeshSlot(fc.Reference.Name, fc.Arguments[0], visibleAssignments, visibleFnTypes)
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

func (f *topoFolder) evalMeshPointQueryCall(fc *program.FnCall, op *program.Operand, visibleAssignments []*program.Assignment, visibleFnTypes []*program.FnType) ([3]float64, error) {
	if len(fc.Arguments) < 1 || len(fc.Arguments) > 2 {
		return [3]float64{}, f.errorAt(op.GetStart(), fmt.Sprintf("%s: expected (id [, transform])", fc.Reference.Name))
	}
	slot, err := f.resolveMeshSlot(fc.Reference.Name, fc.Arguments[0], visibleAssignments, visibleFnTypes)
	if err != nil {
		return [3]float64{}, err
	}
	var xf []float64
	if len(fc.Arguments) == 2 {
		vals, err := f.constArg(fc.Reference.Name, 2, fc.Arguments[1], visibleAssignments, visibleFnTypes)
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

func (f *topoFolder) constArg(name string, pos int, argOp *program.Operand, visibleAssignments []*program.Assignment, visibleFnTypes []*program.FnType) ([]float64, error) {
	if len(argOp.ConstValues) > 0 {
		return argOp.ConstValues, nil
	}
	vals, err := eval.Const(argOp, visibleAssignments, visibleFnTypes)
	if err != nil {
		vals, err = eval.ConstPkgSlotsAt(argOp, f.pkg, f.pkgPath, f.gi, f.slots)
	}
	if err != nil {
		return nil, f.errorAt(argOp.GetStart(), fmt.Sprintf(
			"%s: argument %d must be a constant: %v", name, pos, err))
	}
	return vals, nil
}

func (f *topoFolder) evalMeshPointsQueryCall(fc *program.FnCall, op *program.Operand, visibleAssignments []*program.Assignment, visibleFnTypes []*program.FnType) ([][3]float64, error) {
	if len(fc.Arguments) < 1 || len(fc.Arguments) > 5 {
		return nil, f.errorAt(op.GetStart(), fmt.Sprintf(
			"%s: expected (id [, transform] [, count]) or (id [, transform], axis, maxAngleDeg [, count])", fc.Reference.Name))
	}
	slot, err := f.resolveMeshSlot(fc.Reference.Name, fc.Arguments[0], visibleAssignments, visibleFnTypes)
	if err != nil {
		return nil, err
	}
	rest := make([][]float64, len(fc.Arguments)-1)
	for i, argOp := range fc.Arguments[1:] {
		vals, err := f.constArg(fc.Reference.Name, i+2, argOp, visibleAssignments, visibleFnTypes)
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

func (f *topoFolder) evalCommon(fc *program.FnCall, op *program.Operand, visibleAssignments []*program.Assignment, visibleFnTypes []*program.FnType) (int, AnnotatedMesh, [][]float64, []float64, error) {
	if len(fc.Arguments) < 2 {
		return 0, nil, nil, nil, f.errorAt(op.GetStart(), fmt.Sprintf("%s: expected at least a body and an entity id", fc.Reference.Name))
	}
	args := make([][]float64, len(fc.Arguments))
	for i, argOp := range fc.Arguments {
		vals, err := f.constArg(fc.Reference.Name, i+1, argOp, visibleAssignments, visibleFnTypes)
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

func (f *topoFolder) annotatedMesh(body int, fc *program.FnCall, op *program.Operand) (AnnotatedMesh, error) {
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
	if op == nil {
		return nil
	}
	if t, ok := op.T.(*program.Operand_FnCall); ok {
		if isTopoCall(t.FnCall) {
			return t.FnCall
		}
	}
	return nil
}

func constAssignment(from *program.Assignment, name string, callOp *program.Operand, v [3]float64) *program.Assignment {
	return &program.Assignment{
		Start:      from.GetStart(),
		End:        from.GetEnd(),
		Variables:  []*program.Variable{{Name: name}},
		Expression: constVec(callOp, v),
	}
}

func constScalar(from *program.Operand, v float64) *program.Operand {
	return &program.Operand{
		Start: from.GetStart(),
		End:   from.GetEnd(),
		T:     &program.Operand_Value{Value: v},
	}
}

func constVec(from *program.Operand, v [3]float64) *program.Operand {
	ops := make([]*program.Operand, 3)
	for i, val := range v {
		ops[i] = &program.Operand{
			Start: from.GetStart(),
			End:   from.GetEnd(),
			T:     &program.Operand_Value{Value: val},
		}
	}
	return &program.Operand{
		Start: from.GetStart(),
		End:   from.GetEnd(),
		T: &program.Operand_Expression{Expression: &program.Expression{
			Operation: program.Operation_COLLECT,
			Operands:  ops,
		}},
	}
}

func constVecList(from *program.Operand, pts [][3]float64) *program.Operand {
	ops := make([]*program.Operand, len(pts))
	for i, p := range pts {
		ops[i] = constVec(from, p)
	}
	return &program.Operand{
		Start: from.GetStart(),
		End:   from.GetEnd(),
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

func extractPlacements(pkg *program.Package, getImport program.GetImportFunc) (map[int][]float64, map[int]int, error) {
	placements := make(map[int][]float64)
	counts := make(map[int]int)

	var walkOperand func(op *program.Operand)
	walkOperand = func(op *program.Operand) {
		if op == nil {
			return
		}
		switch t := op.T.(type) {
		case *program.Operand_FnCall:
			fc := t.FnCall
			if fc != nil && fc.Reference != nil && fc.Reference.Prefix == "" && fc.Reference.Name == "mesh" {
				index, mat, err := parseMeshCallMatrix(fc, pkg, getImport)
				if err == nil {
					placements[index] = mat
					counts[index]++
				}
			}
			for _, arg := range fc.Arguments {
				walkOperand(arg)
			}
		case *program.Operand_Expression:
			for _, sub := range t.Expression.Operands {
				walkOperand(sub)
			}
		}
	}

	for _, file := range pkg.Files {
		for _, a := range file.Assignments {
			walkOperand(a.Expression)
		}
		for _, r := range file.Returns {
			walkOperand(r)
		}
		for _, ft := range file.FnTypes {
			for _, g := range ft.Groups {
				for _, fn := range g.Functions {
					for _, a := range fn.Assignments {
						walkOperand(a.Expression)
					}
					for _, r := range fn.Returns {
						walkOperand(r)
					}
				}
			}
		}
	}

	return placements, counts, nil
}

func parseMeshCallMatrix(fc *program.FnCall, pkg *program.Package, getImport program.GetImportFunc) (int, []float64, error) {
	args := fc.Arguments
	if len(args) == 0 {
		return 0, nil, fmt.Errorf("mesh() requires at least one argument")
	}

	// Slot index
	var index int
	if v, ok := args[0].T.(*program.Operand_Value); ok {
		index = int(v.Value)
	} else if len(args[0].ConstValues) == 1 {
		index = int(math.Round(args[0].ConstValues[0]))
	} else {
		idx, err := eval.ConstPkg(args[0], pkg, getImport)
		if err != nil || len(idx) != 1 {
			return 0, nil, fmt.Errorf("mesh() first argument not a constant: %v", err)
		}
		index = int(idx[0])
	}

	if len(args) <= 1 {
		return index, []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}, nil
	}

	var v []float64
	if len(args[1].ConstValues) > 0 {
		v = args[1].ConstValues
	} else {
		var err error
		v, err = eval.ConstPkg(args[1], pkg, getImport)
		if err != nil {
			return index, nil, err
		}
	}

	switch {
	case len(v) == 16:
		return index, append([]float64(nil), v...), nil
	case len(v) == 3 && len(args) == 2:
		return index, []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, v[0], v[1], v[2], 1}, nil
	case len(v) == 3 && len(args) == 3:
		var r []float64
		if len(args[2].ConstValues) > 0 {
			r = args[2].ConstValues
		} else {
			var err error
			r, err = eval.ConstPkg(args[2], pkg, getImport)
			if err != nil {
				return index, nil, err
			}
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
