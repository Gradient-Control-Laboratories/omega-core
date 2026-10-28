package build

import (
	"gradientcontrol.com/omega/core/program"
)

type Delegate interface {

	// Return the current scope.
	GetScope() *Scope

	// Sets the current scope for upcoming delegate method calls
	// and returns the previous scope.
	SetScope(scope *Scope) *Scope

	// Called when starting to build the primary package.
	// Return an error if the package cannot be built.
	//
	// The delegate should cache any built-in types and functions
	// before returning (e.g "p" as the position variable)
	BuildPackage(primaryPkg *program.Package) error

	// Build the primary package's return expressions.
	// Return an error if the package cannot be built.
	//
	// This is only called for the primary package, and is the last
	// delegate method called for the primary package.
	BuildPackageReturns(ops []*program.Operand, expressions []Expression) error

	// -----------------------------------------------------------------------

	// Start building the given import and return the import info associated with it.
	BuildImport(imp *program.Import, pkg *program.Package) (ImportInfo, error)

	// Finish building the given import.
	EndBuildImport(info ImportInfo) error

	// -----------------------------------------------------------------------

	// Build the remap operation where the target expression is the expression being remapped.
	// remapOp is the original remap operation
	// targetOp is the target of the remap operation
	// pExpr is the "p" variable expression
	// targetExpr is the expression being remapped
	Remap(remapOp *program.Operand, targetOp *program.Operand, pExpr Expression, targetExpr Expression) (Expression, error)

	// -----------------------------------------------------------------------

	// Returns a new VarInfo for the given function name and type.
	NewFnVar(name string, decl *program.FnType, enclosingScope *Scope) VarInfo

	// Returns a new VarInfo representing a parameter of an inlined function call.
	NewParamVar(scope *Scope, name string, expr Expression, typ *program.Type) VarInfo

	// Start building the given function and return the info associated with it.
	// Returns func info with an array of parameter [VarInfo], in order
	BuildFunc(fnDecl *program.FnType, fn *program.Function,
		pkgName string, args []Expression) (FnInfo, []VarInfo, error)

	// Finish building the given function and return the info associated with it.
	BuildFuncReturn(fnInfo FnInfo, returnOps []*program.Operand, returnExprs []Expression) (FnInfo, error)

	// Build the given function call and return the expression associated with it.
	BuildFuncCall(fnCallOp *program.Operand, fnCall *program.FnCall,
		fnInfo FnInfo, args []Expression) (Expression, error)

	// Build the given builtin call and return the expression associated with it.
	BuildBuiltinCall(op *program.Operand, fnCall *program.FnCall,
		builtin *program.BuiltinType, args []Expression) (Expression, error)

	// -----------------------------------------------------------------------

	// Return an array of implied top level varInfo such as "p" and "time"
	NewImpliedVars(scope *Scope) ([]VarInfo, error)

	// Build the given reference and return the expression associated with it.
	BuildVarRef(op *program.Operand, varInfo VarInfo) (Expression, error)

	// Build the given single assignment and return the variable info associated with it.
	BuildSingleAssignment(decl *program.Assignment, op *program.Operand,
		name string, expr Expression) (VarInfo, error)

	// Build the given multi assignment and return an error if the variable info object cannot be created.
	BuildMultiAssignment(decl *program.Assignment, op *program.Operand,
		fnCall *program.FnCall, names []string, expr Expression) ([]VarInfo, error)

	// -----------------------------------------------------------------------

	// Build the given value and return the expression associated with it.
	BuildValue(op *program.Operand, value float64) (ConstExpression, error)

	// Build the given access and return the expression associated with it.
	BuildMemberAccess(op *program.Operand, expr *program.Expression,
		target Expression, member string) (Expression, error)

	// Build the given index and return the expression associated with it.
	BuildIndexAccess(op *program.Operand, expr *program.Expression,
		targetOp *program.Operand, indexOp *program.Operand,
		target Expression, index int) (Expression, error)

	// Build the given operation and return the expression associated with it.
	BuildOperation(op *program.Operand, expr *program.Expression, symbol string,
		operands []*program.Operand, expressions []Expression,
		resultType *program.Type, precedence Precedence) (Expression, error)

	// Build the given operation call and return the expression associated with it.
	BuildOpCall(op *program.Operand, expr *program.Expression,
		operands []*program.Operand, expressions []Expression, resultType *program.Type) (Expression, error)

	// Build the logical not and return the expression associated with it.
	BuildNot(op *program.Operand, expr *program.Expression,
		targetOp *program.Operand, target Expression) (Expression, error)

	// Build the given negation and return the expression associated with it.
	BuildNeg(op *program.Operand, expr *program.Expression,
		targetOp *program.Operand, target Expression) (Expression, error)

	// Build the given comparison and return the expression associated with it.
	BuildComparison(op *program.Operand, expr *program.Expression, symbol string, operands []*program.Operand,
		expressions []Expression, resultType *program.Type, precedence Precedence) (Expression, error)

	// Build the given compare and return the expression associated with it.
	BuildCompare(op *program.Operand, expr *program.Expression, operands []*program.Operand,
		expr1 Expression, expr2 Expression, resultType *program.Type) (Expression, error)

	// Build the given select operation (condition ? trueResult : falseResult) and return the expression associated with it.
	BuildSelect2(op *program.Operand, expr *program.Expression, operands []*program.Operand,
		conditionExpr Expression,
		trueResultExpr Expression, falseResultExpr Expression, resultType *program.Type) (Expression, error)

	// Build the given select operation (value ? ltZeroResult : eqZeroResult : gtZeroResult) and return the expression associated with it.
	BuildSelect3(op *program.Operand, expr *program.Expression, operands []*program.Operand,
		valueExpr Expression, ltZeroResultExpr Expression, eqZeroResultExpr Expression, gtZeroResultExpr Expression,
		resultType *program.Type) (Expression, error)

	// Build the given collect operation and return the expression associated with it.
	BuildCollect(op *program.Operand, expr *program.Expression, operands []*program.Operand,
		expressions []Expression) (Expression, error)

	// Build the given constant collect operation and return the expression associated with it.
	BuildConstCollect(op *program.Operand, expr *program.Expression, operands []*program.Operand,
		expressions []Expression, constValues []float64) (ConstExpression, error)
}

// ColorReturnsBuilder is an optional Delegate extension. A delegate that renders
// per-return color implements it to receive each return's lowered color field
// (nil where that return authors no color). Called just before
// BuildPackageReturns; delegates that don't implement it render no color.
type ColorReturnsBuilder interface {
	BuildPackageColorReturns(colorExpressions []Expression) error
}

// MaterialReturnsBuilder is the material counterpart of ColorReturnsBuilder: a
// delegate that renders materials receives each return's lowered material
// field (a list of vec4 rows; nil where that return authors no material),
// after the color hook and before BuildPackageReturns. Delegates that don't
// implement it (meshing, the CPU evaluator) never see materials.
type MaterialReturnsBuilder interface {
	BuildPackageMaterialReturns(materialExpressions []Expression) error
}
