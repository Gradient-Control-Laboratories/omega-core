package common_test

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/backend/common"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

type dummyDelegate struct {
	build.Delegate
	scope *common.Scope
}

func (d *dummyDelegate) GetScope() *common.Scope {
	return d.scope
}

func (d *dummyDelegate) SetScope(s *common.Scope) *common.Scope {
	prev := d.scope
	d.scope = s
	return prev
}

func (d *dummyDelegate) BuildPackage(pkg *program.Package) error {
	return nil
}

func (d *dummyDelegate) BuildPackageReturns(ops []*program.Operand, exprs []common.Expression) error {
	return nil
}

func (d *dummyDelegate) NewImpliedVars(s *common.Scope) ([]build.VarInfo, error) {
	return nil, nil
}

func (d *dummyDelegate) BuildValue(op *program.Operand, val float64) (build.ConstExpression, error) {
	return &dummyConstExpr{op: op, val: val}, nil
}

func (d *dummyDelegate) BuildSingleAssignment(decl *program.Assignment, op *program.Operand, name string, expr common.Expression) (build.VarInfo, error) {
	return &dummyVarInfo{name: name, expr: expr}, nil
}

func (d *dummyDelegate) BuildImport(imp *program.Import, pkg *program.Package) (build.ImportInfo, error) {
	return &dummyImportInfo{}, nil
}

func (d *dummyDelegate) EndBuildImport(info build.ImportInfo) error {
	return nil
}

func (d *dummyDelegate) BuildVarRef(op *program.Operand, v build.VarInfo) (common.Expression, error) {
	return &dummyConstExpr{op: op, val: 0}, nil
}

type dummyVarInfo struct {
	name string
	expr common.Expression
}

func (m *dummyVarInfo) Name() string                   { return m.name }
func (m *dummyVarInfo) Expr() common.Expression        { return m.expr }
func (m *dummyVarInfo) Type() *program.Type            { return nil }
func (m *dummyVarInfo) Scope() *common.Scope           { return nil }
func (m *dummyVarInfo) SetExpr(expr common.Expression) { m.expr = expr }

type dummyImportInfo struct{}

func (m *dummyImportInfo) Import() *program.Import   { return nil }
func (m *dummyImportInfo) Scope() *common.Scope      { return nil }
func (m *dummyImportInfo) Package() *program.Package { return nil }

type dummyConstExpr struct {
	op  *program.Operand
	val float64
}

func (m *dummyConstExpr) Operand() *program.Operand    { return m.op }
func (m *dummyConstExpr) Start() *program.Location     { return m.op.Start }
func (m *dummyConstExpr) Precedence() build.Precedence { return build.PrecedenceHighest }
func (m *dummyConstExpr) Type() *program.Type          { return nil }
func (m *dummyConstExpr) ConstValues() []float64       { return []float64{m.val} }

func TestBuild_ValidProgram(t *testing.T) {
	res, err := assembler.Assemble([]string{"x = 10;\nreturn x;"})
	if err != nil {
		t.Fatalf("failed to assemble program: %v", err)
	}
	err = common.Build(&dummyDelegate{}, res.Program, false)
	if err != nil {
		t.Fatalf("expected nil error for program, got: %v", err)
	}
}
