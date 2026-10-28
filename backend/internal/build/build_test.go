package build_test

import (
	"testing"

	assembler "gradientcontrol.com/omega/core/assembler/internal/lib"
	"gradientcontrol.com/omega/core/backend/internal/build"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/program"
)

var mockMgr = source.NewLocalManager("testpkg", "../../../../common/source/package")

type mockDelegate struct {
	build.Delegate
	scope *build.Scope

	builtPackage bool
	returnedOps  []*program.Operand
	operands     []*program.Operand
}

func newMockDelegate() *mockDelegate {
	return &mockDelegate{}
}

func (d *mockDelegate) GetScope() *build.Scope {
	return d.scope
}

func (d *mockDelegate) SetScope(s *build.Scope) *build.Scope {
	prev := d.scope
	d.scope = s
	return prev
}

func (d *mockDelegate) BuildPackage(pkg *program.Package) error {
	d.builtPackage = true
	return nil
}

func (d *mockDelegate) BuildPackageReturns(ops []*program.Operand, expressions []build.Expression) error {
	d.returnedOps = ops
	return nil
}

func (d *mockDelegate) NewImpliedVars(s *build.Scope) ([]build.VarInfo, error) {
	return nil, nil
}

func (d *mockDelegate) BuildValue(op *program.Operand, value float64) (build.ConstExpression, error) {
	d.operands = append(d.operands, op)
	return &mockConstExpr{op: op, val: value}, nil
}

type mockImportInfo struct{}

func (d *mockDelegate) BuildImport(imp *program.Import, pkg *program.Package) (build.ImportInfo, error) {
	return &mockImportInfo{}, nil
}

func (d *mockDelegate) EndBuildImport(info build.ImportInfo) error {
	return nil
}

func (d *mockDelegate) BuildSingleAssignment(decl *program.Assignment, op *program.Operand, name string, expr build.Expression) (build.VarInfo, error) {
	return &mockVarInfo{name: name, expr: expr}, nil
}

func (d *mockDelegate) BuildMultiAssignment(decl *program.Assignment, op *program.Operand, fnCall *program.FnCall, names []string, expr build.Expression) ([]build.VarInfo, error) {
	var infos []build.VarInfo
	for _, n := range names {
		infos = append(infos, &mockVarInfo{name: n, expr: expr})
	}
	return infos, nil
}

func (d *mockDelegate) BuildFunc(fnDecl *program.FnType, fn *program.Function, pkgName string, args []build.Expression) (build.FnInfo, []build.VarInfo, error) {
	return &mockFnInfo{}, nil, nil
}

func (d *mockDelegate) BuildFuncReturn(fnInfo build.FnInfo, returnOps []*program.Operand, returnExprs []build.Expression) (build.FnInfo, error) {
	return fnInfo, nil
}

func (d *mockDelegate) BuildFuncCall(fnCallOp *program.Operand, fnCall *program.FnCall, fnInfo build.FnInfo, args []build.Expression) (build.Expression, error) {
	return &mockConstExpr{op: fnCallOp, val: 0}, nil
}

func (d *mockDelegate) BuildBuiltinCall(op *program.Operand, fnCall *program.FnCall, builtin *program.BuiltinType, args []build.Expression) (build.Expression, error) {
	return &mockConstExpr{op: op, val: 0}, nil
}

func (d *mockDelegate) NewFnVar(name string, decl *program.FnType, enclosingScope *build.Scope) build.VarInfo {
	return &mockVarInfo{name: name}
}

func (d *mockDelegate) NewParamVar(scope *build.Scope, name string, expr build.Expression, typ *program.Type) build.VarInfo {
	return &mockVarInfo{name: name, expr: expr}
}

func (d *mockDelegate) Remap(remapOp *program.Operand, targetOp *program.Operand, pExpr build.Expression, targetExpr build.Expression) (build.Expression, error) {
	return targetExpr, nil
}

func (d *mockDelegate) BuildConstCollect(op *program.Operand, expr *program.Expression, elemOps []*program.Operand, exprs []build.Expression, vals []float64) (build.ConstExpression, error) {
	return &mockConstExpr{op: op, val: 0}, nil
}

func (d *mockDelegate) BuildMemberAccess(op *program.Operand, expr *program.Expression, target build.Expression, member string) (build.Expression, error) {
	return &mockConstExpr{op: op, val: 0}, nil
}

func (d *mockDelegate) BuildVarRef(op *program.Operand, v build.VarInfo) (build.Expression, error) {
	return &mockConstExpr{op: op, val: 0}, nil
}

type mockVarInfo struct {
	name string
	expr build.Expression
}

func (m *mockVarInfo) Name() string                  { return m.name }
func (m *mockVarInfo) Expr() build.Expression        { return m.expr }
func (m *mockVarInfo) Type() *program.Type           { return nil }
func (m *mockVarInfo) Scope() *build.Scope           { return nil }
func (m *mockVarInfo) SetExpr(expr build.Expression) { m.expr = expr }

type mockFnInfo struct{}

func (m *mockFnInfo) Name() string                { return "" }
func (m *mockFnInfo) ParamTypes() []*program.Type { return nil }

func (m *mockImportInfo) Import() *program.Import   { return nil }
func (m *mockImportInfo) Scope() *build.Scope       { return nil }
func (m *mockImportInfo) Package() *program.Package { return nil }

type mockConstExpr struct {
	op  *program.Operand
	val float64
}

func (m *mockConstExpr) Operand() *program.Operand    { return m.op }
func (m *mockConstExpr) Start() *program.Location     { return m.op.Start }
func (m *mockConstExpr) Precedence() build.Precedence { return build.PrecedenceHighest }
func (m *mockConstExpr) Type() *program.Type          { return nil }
func (m *mockConstExpr) ConstValues() []float64       { return []float64{m.val} }

func TestBuild_NativeProgramType(t *testing.T) {
	src := "return 42.0;"
	res, err := assembler.Assemble([]string{src}, mockMgr)
	assert.NoError(t, err)

	d := newMockDelegate()
	err = build.Build(d, res.Program, false)
	assert.NoError(t, err)

	assert.Equal(t, true, d.builtPackage)
	assert.Equal(t, 1, len(d.returnedOps))
	assert.Equal(t, int32(1), d.returnedOps[0].Type.Size)
}

type mockColorDelegate struct {
	*mockDelegate
	receivedColors []build.Expression
}

func (m *mockColorDelegate) BuildPackageColorReturns(colorExpressions []build.Expression) error {
	m.receivedColors = colorExpressions
	return nil
}

type mockMaterialDelegate struct {
	*mockDelegate
	receivedMaterials []build.Expression
}

func (m *mockMaterialDelegate) BuildPackageMaterialReturns(materialExpressions []build.Expression) error {
	m.receivedMaterials = materialExpressions
	return nil
}

func TestBuild_ColorReturnsForwarded(t *testing.T) {
	src := `
target = [1.0, 2.0, 3.0];
target.color = [1.0, 0.0, 0.0];
return target;
`
	res, err := assembler.Assemble([]string{src}, mockMgr)
	assert.NoError(t, err)

	d := &mockColorDelegate{mockDelegate: newMockDelegate()}
	err = build.Build(d, res.Program, false)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(d.receivedColors))
	assert.True(t, d.receivedColors[0] != nil)
}

func TestBuild_MaterialReturnsForwarded(t *testing.T) {
	src := `
target = [1.0, 2.0, 3.0];
target.material = [0.2, 0.4, 0.6];
return target;
`
	res, err := assembler.Assemble([]string{src}, mockMgr)
	assert.NoError(t, err)

	d := &mockMaterialDelegate{mockDelegate: newMockDelegate()}
	err = build.Build(d, res.Program, false)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(d.receivedMaterials))
	assert.True(t, d.receivedMaterials[0] != nil)
}

func TestBuild_NeitherColorNorMaterial(t *testing.T) {
	src := "return 42.0;"
	res, err := assembler.Assemble([]string{src}, mockMgr)
	assert.NoError(t, err)

	d := &mockColorDelegate{mockDelegate: newMockDelegate()}
	err = build.Build(d, res.Program, false)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(d.receivedColors))
}
