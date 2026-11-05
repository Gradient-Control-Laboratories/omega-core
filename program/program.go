package program

import (
	"math"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const VersionMajor = 0
const VersionMinor = 0

const XVarName = "x"
const YVarName = "y"
const ZVarName = "z"
const PVarName = "p"
const TimeVarName = "time"
const DPAccessor = "dp"
const DTimeAccessor = "dtime"
const DAccessor = "d"
const SecondAccessor = "second"
const BoxAccessor = "box"
const UAccessor = "u"
const VAccessor = "v"

const AnonFnPrefix = "gsl_anon"
const InlinedFnPrefix = "gsl_inlined_"

// ParsePackageFunc parses package source code into a program.Package.
type ParsePackageFunc func(sources []string) (*Package, error)

func (i *Import) SimpleName() string {
	return i.Path[strings.LastIndex(i.Path, "/")+1:]
}

func (i *Import) PrefixOrSimpleName() string {
	if i.Prefix != "" {
		return i.Prefix
	}
	return i.SimpleName()
}

func (d *FnType) IsAnonFn() bool {
	return strings.HasPrefix(d.Name, AnonFnPrefix)
}

func (op Operation) Name() string {
	return strings.ToLower(Operation_name[int32(op)])
}

// IsBool reports whether the operation is a comparison (EQ, NE, LT, LE, GT, GE)
// or boolean logic (AND, OR, NOT) operation that yields a boolean result.
func (op Operation) IsBool() bool {
	switch op {
	case Operation_EQ, Operation_NE,
		Operation_LT, Operation_LE,
		Operation_GT, Operation_GE,
		Operation_AND, Operation_OR,
		Operation_NOT:
		return true
	default:
		return false
	}
}

// IsBool reports whether the operand represents a logical or comparison expression.
func (op *Operand) IsBool() bool {
	return op.GetExpression().GetOperation().IsBool()
}

func NewImport(prefix string, path string, start, end *Location) *Import {
	return &Import{Prefix: prefix, Path: path, Start: start, End: end}
}

func NewMeshImport(index int, path string, start, end *Location) *MeshImport {
	return &MeshImport{Index: int32(index), Path: path, Start: start, End: end}
}

func NewStepImport(index int, path string, start, end *Location) *StepImport {
	return &StepImport{Index: int32(index), Path: path, Start: start, End: end}
}

func NewWarpImport(index int, path string, start, end *Location) *WarpImport {
	return &WarpImport{Index: int32(index), Path: path, Start: start, End: end}
}

func (p *PackageFile) AddImport(imp *Import) {
	if p.Imports == nil {
		p.Imports = make([]*Import, 0)
	}
	p.Imports = append(p.Imports, imp)
}

func NewAssignment(names []string, expr *Operand, start, end *Location) *Assignment {
	vars := make([]*Variable, len(names))
	for i, name := range names {
		vars[i] = &Variable{Name: name}
	}
	return &Assignment{Variables: vars, Expression: expr, Start: start, End: end}
}

func (a *Variable) AddPropertyAssignment(propName string, expr *Operand, start, end *Location) {
	a.Properties = append(a.Properties, &Property{
		Name: propName, Expression: expr,
		Start: start, End: end,
	})
}

func AddFunction(
	fnTypes []*FnType, fnName string, paramNames []string,
) ([]*FnType, *Function) {

	// Find or create the function type with the given name
	var fnType *FnType
	for _, ft := range fnTypes {
		if ft.Name == fnName {
			fnType = ft
			break
		}
	}
	if fnType == nil {
		fnType = &FnType{Name: fnName}
		fnTypes = append(fnTypes, fnType)
	}

	// Find or create the function group the the given number of parameters
	paramCount := int32(len(paramNames))
	var fnGroup *FnGroup
	for _, fg := range fnType.Groups {
		if fg.ParamCount == paramCount {
			fnGroup = fg
			break
		}
	}
	if fnGroup == nil {
		fnGroup = &FnGroup{ParamCount: paramCount}
		fnType.Groups = append(fnType.Groups, fnGroup)
	}

	fn := &Function{}
	fn.Parameters = make([]*Parameter, len(paramNames))
	for i, name := range paramNames {
		fn.Parameters[i] = &Parameter{Name: name}
	}
	fnGroup.Functions = append(fnGroup.Functions, fn)

	return fnTypes, fn
}

func NewExpression(operation Operation, operands []*Operand, start, end *Location) *Operand {
	return &Operand{
		Start: start,
		End:   end,
		T: &Operand_Expression{
			Expression: &Expression{
				Operation: operation,
				Operands:  operands,
			}}}
}

func NewFnCall(pkgName, fnName string, arguments []*Operand, start, end *Location) *Operand {
	return &Operand{
		Start: start,
		End:   end,
		T: &Operand_FnCall{
			FnCall: &FnCall{
				Reference: &Reference{Prefix: pkgName, Name: fnName},
				Arguments: arguments,
			},
		}}
}

func NewAccess(target *Operand, accessor string, end *Location) *Operand {
	return &Operand{
		Start: target.Start,
		End:   end,
		T: &Operand_Expression{
			Expression: &Expression{
				Operation: Operation_ACCESS,
				Operands: []*Operand{
					target,
					{
						Start: &Location{
							Line:   end.Line,
							Column: int32(int(end.Column) - len(accessor)),
						},
						End: end,
						T: &Operand_Reference{
							Reference: &Reference{Name: accessor},
						},
					},
				},
			}}}
}

func NewReference(prefix, name string, start, end *Location) *Operand {
	return &Operand{
		Start: start,
		End:   end,
		T: &Operand_Reference{
			Reference: &Reference{
				Prefix: prefix,
				Name:   name,
			},
		}}
}

func NewPropertyRef(varName, propName string, start, end *Location) *Operand {
	return &Operand{
		Start: start,
		End:   end,
		T: &Operand_PropRef{
			PropRef: &PropertyRef{
				Reference: &Reference{
					Name: varName,
				},
				Property: propName,
			},
		}}
}

func NewValue(value float64, start, end *Location) *Operand {
	return &Operand{
		Start: start,
		End:   end,
		T: &Operand_Value{
			Value: value,
		}}
}

func NewLoc(line int32, col int32, offset int32) *Location {
	return &Location{Line: line, Column: col, Offset: offset}
}

// Scene returns the scene framing configuration of the primary package,
// or nil if no primary package or scene is defined.
func (p *Program) Scene() *Scene {
	return p.Primary.Scene
}

// Returns returns the return operands of the package's last file,
// or nil if no files or returns exist.
func (pkg *Package) Returns() []*Operand {
	if len(pkg.Files) == 0 {
		return nil
	}
	return pkg.Files[len(pkg.Files)-1].Returns
}

// SetReturns sets the return operands of the package's last file.
func (pkg *Package) SetReturns(returns []*Operand) {
	if len(pkg.Files) == 0 {
		return
	}
	pkg.Files[len(pkg.Files)-1].Returns = returns
}

// Returns returns the return operands of the primary package's last file,
// or nil if no files or returns exist.
func (p *Program) Returns() []*Operand {
	return p.Primary.Returns()
}

// SetReturns sets the return operands of the primary package's last file.
func (p *Program) SetReturns(returns []*Operand) {
	p.Primary.SetReturns(returns)
}

// ColorReturns returns the lowered color operands of the primary package's
// last file, or nil if no files or color returns exist.
func (p *Program) ColorReturns() []*Operand {
	if len(p.Primary.Files) == 0 {
		return nil
	}
	return p.Primary.Files[len(p.Primary.Files)-1].ColorReturns
}

// MaterialReturns returns the lowered material operands of the primary package's
// last file, or nil if no files or material returns exist.
func (p *Program) MaterialReturns() []*Operand {
	if len(p.Primary.Files) == 0 {
		return nil
	}
	return p.Primary.Files[len(p.Primary.Files)-1].MaterialReturns
}

// ModuleFunction finds a package-local function declaration with the given
// name and parameter count.
func ModuleFunction(pkg *Package, name string, paramCount int) *Function {
	if pkg == nil {
		return nil
	}
	for _, f := range pkg.Files {
		for _, ft := range f.FnTypes {
			if ft.Name != name {
				continue
			}
			for _, g := range ft.Groups {
				if int(g.ParamCount) == paramCount && len(g.Functions) > 0 {
					return g.Functions[0]
				}
			}
		}
	}
	return nil
}

// TopLevelAssignment returns the RHS expression of a single-variable top-level
// assignment with the given variable name across the package's files, or nil.
func TopLevelAssignment(pkg *Package, name string) *Operand {
	if pkg == nil {
		return nil
	}
	for _, file := range pkg.Files {
		if file == nil {
			continue
		}
		for _, a := range file.Assignments {
			if a != nil && len(a.Variables) == 1 && a.Variables[0].GetName() == name {
				return a.Expression
			}
		}
	}
	return nil
}

// ConstValues returns finite constant float64 values from an operand's stamped
// ConstValues slice or from literal numeric/collection AST nodes.
func ConstValues(op *Operand) ([]float64, bool) {
	if op == nil {
		return nil, false
	}
	if len(op.ConstValues) > 0 {
		for _, v := range op.ConstValues {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return nil, false
			}
		}
		return append([]float64(nil), op.ConstValues...), true
	}
	switch t := op.T.(type) {
	case *Operand_Value:
		if math.IsNaN(t.Value) || math.IsInf(t.Value, 0) {
			return nil, false
		}
		return []float64{t.Value}, true
	case *Operand_Expression:
		if t.Expression == nil {
			return nil, false
		}
		switch t.Expression.Operation {
		case Operation_NEG:
			if len(t.Expression.Operands) != 1 {
				return nil, false
			}
			vals, ok := ConstValues(t.Expression.Operands[0])
			if !ok {
				return nil, false
			}
			for i := range vals {
				vals[i] = -vals[i]
			}
			return vals, true
		case Operation_COLLECT:
			if len(t.Expression.Operands) == 0 {
				return nil, false
			}
			var out []float64
			for _, child := range t.Expression.Operands {
				vals, ok := ConstValues(child)
				if !ok {
					return nil, false
				}
				out = append(out, vals...)
			}
			return out, true
		default:
			return nil, false
		}
	default:
		return nil, false
	}
}

// EqualModuloLocations compares two protobuf AST messages for structural
// equality while ignoring source Location and stamped ConstValues fields.
func EqualModuloLocations(a, b proto.Message) bool {
	if a == nil || b == nil || !a.ProtoReflect().IsValid() || !b.ProtoReflect().IsValid() {
		return false
	}
	a, b = proto.Clone(a), proto.Clone(b)
	var clean func(protoreflect.Message)
	clean = func(m protoreflect.Message) {
		m.Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
			if field.Name() == "constValues" {
				m.Clear(field)
				return true
			}
			if field.Kind() != protoreflect.MessageKind {
				return true
			}
			if field.Message().Name() == "Location" {
				m.Clear(field)
			} else if field.IsList() {
				for i := 0; i < value.List().Len(); i++ {
					clean(value.List().Get(i).Message())
				}
			} else if !field.IsMap() {
				clean(value.Message())
			}
			return true
		})
	}
	clean(a.ProtoReflect())
	clean(b.ProtoReflect())
	return proto.Equal(a, b)
}
