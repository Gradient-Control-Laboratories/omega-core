package program

import (
	"strings"
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
