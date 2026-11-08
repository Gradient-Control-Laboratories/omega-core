package codegen

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/ir"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/lower"
	"gradientcontrol.com/omega/core/program"
)

// A reduce over a list of constants, kept by the assembler's LoopFolds
// (lower.Options), is emitted as a loop: its callback becomes one helper
// function, the step, that the loop calls once per element,
//
//	var gsl_fold = <init>;
//	for (var gsl_i = 0; gsl_i < N; gsl_i = gsl_i + 1) {
//	  gsl_fold = step(gsl_fold, gsl_foldK[gsl_i], f32(gsl_i), <captures>);
//	}
//
// rather than a copy of the callback per element. The dynamic work is the
// same; the code is not, and a shader's compile time grows with its code: a
// cell scan unrolled is hundreds of lines at each call site. The elements are
// read from a module-scope constant array, or are the index itself when the
// list counts up from 0. What the callback reads of the function around it
// arrives as the step's parameters, as a remap helper's do (remap.go).
//
// A fold whose accumulator is not a number or a vector of them is unrolled
// as before.

// BuildFoldLoop implements [build.FoldLooper].
func (d *Delegate) BuildFoldLoop(op *program.Operand, init build.Expression, elements []build.ConstExpression,
	step func(acc, element, index build.Expression) (build.Expression, error)) (build.Expression, bool, error) {

	if len(elements) < lower.LoopFoldThreshold {
		return nil, false, nil
	}
	initExpr, ok := init.(wgsl.Expression)
	if !ok || !foldableType(initExpr.WGSLType()) {
		return nil, false, nil
	}
	accType := initExpr.WGSLType()
	elemType, values, ok := foldElements(elements)
	if !ok {
		return nil, false, nil
	}

	bScope := d.GetScope()
	accName := d.uniqueLocalName(bScope.NewVarName("gsl_fold"))
	indexName := d.uniqueLocalName(bScope.NewVarName("gsl_i"))

	acc := wgsl.NewOpaqueExpr(nil, accType, build.PrecedenceHighest, accName, accName)
	index := wgsl.NewOpaqueExpr(nil, wgsl.TypeF32, build.PrecedenceHighest, "f32("+indexName+")", indexName)
	element := index
	if !countsUp(elemType, values) {
		arrayName := d.foldConstArray(elemType, values)
		element = wgsl.NewOpaqueExpr(nil, elemType, build.PrecedenceHighest,
			fmt.Sprintf("%s[%s]", arrayName, indexName), arrayName, indexName)
	}

	pkgScope := d.currentScope()
	for pkgScope.parent != nil && pkgScope.parent != d.rootScope() {
		pkgScope = pkgScope.parent
	}
	origFuncCount := len(pkgScope.funcs)

	// The step reads what the fold's own scope and those around it define as
	// its parameters, which the loop passes in; built, they are not derived
	// again in its body (ShouldGenerate).
	d.foldScopes = append(d.foldScopes, bScope)
	call, err := step(acc, element, index)
	d.foldScopes = d.foldScopes[:len(d.foldScopes)-1]
	if err != nil {
		return nil, false, err
	}
	callExpr, ok := call.(wgsl.Expression)
	if !ok {
		return nil, false, nil
	}
	if !callExpr.WGSLType().Equal(accType) {
		retType := callExpr.WGSLType()
		if accType.IsNumericScalar() && retType.IsNumericVector() && retType.VectorSize() <= 4 {
			promotedInit := promoteScalarToVector(initExpr, retType)
			promotedAcc := wgsl.NewOpaqueExpr(nil, retType, build.PrecedenceHighest, accName, accName)
			for i := origFuncCount; i < len(pkgScope.funcs); i++ {
				delete(d.assignedNames, pkgScope.funcs[i].Name)
			}
			pkgScope.funcs = pkgScope.funcs[:origFuncCount]

			d.foldScopes = append(d.foldScopes, bScope)
			call2, err2 := step(promotedAcc, element, index)
			d.foldScopes = d.foldScopes[:len(d.foldScopes)-1]
			if err2 == nil {
				if call2Expr, ok2 := call2.(wgsl.Expression); ok2 && call2Expr.WGSLType().Equal(retType) {
					callExpr = call2Expr
					accType = retType
					initExpr = promotedInit
				}
			}
		}
		if !callExpr.WGSLType().Equal(accType) {
			return nil, false, nil
		}
	}

	varInfo := wgsl.NewVar(bScope, accName, callExpr, accType)
	varInfo.SetWGSLName(accName)
	if d.foldLoops == nil {
		d.foldLoops = make(map[*wgsl.VarInfo]func(ctx wgsl.CodegenContext))
	}
	// emitVar writes the loop where the fold's value is first read.
	d.foldLoops[varInfo] = func(ctx wgsl.CodegenContext) {
		// Both are written out before the loop, so a let either one needs
		// lands ahead of it, not inside it.
		initSrc := initExpr.Codegen(ctx)
		callSrc := callExpr.Codegen(ctx)
		body := ir.NewBlock()
		body.Add(ir.NewAssignStmt(accName, callSrc, append([]string{accName}, callExpr.Reads()...)))
		d.currentScope().AddStmt(ir.NewVarStmt(accName, "", initSrc, initExpr.Reads()))
		d.currentScope().AddStmt(ir.NewLoopStmt(indexName, len(elements), body))
	}
	if err := bScope.AddLocalVar(varInfo); err != nil {
		return nil, false, err
	}
	ref, err := d.BuildVarRef(op, varInfo)
	if err != nil {
		return nil, false, err
	}
	return ref, true, nil
}

// definedAroundFold reports whether scope is the scope of a fold whose step is
// being built, or one around it.
func (d *Delegate) definedAroundFold(scope *build.Scope) bool {
	for _, s := range d.foldScopes {
		for curr := s; curr != nil; curr = curr.Parent() {
			if curr == scope {
				return true
			}
		}
	}
	return false
}

// foldableType reports whether a loop can carry t: a number or a vector of
// them.
func foldableType(t wgsl.TypeInfo) bool {
	return t.IsNumericScalar() || (t.IsNumericVector() && t.VectorSize() <= 4)
}

// foldElements returns the type the elements share and their values, one row
// per element, or false if they do not share a type a loop can carry.
func foldElements(elements []build.ConstExpression) (wgsl.TypeInfo, [][]float64, bool) {
	var elemType wgsl.TypeInfo
	values := make([][]float64, len(elements))
	for i, el := range elements {
		e, ok := el.(wgsl.Expression)
		if !ok {
			return wgsl.TypeInfo{}, nil, false
		}
		t := e.WGSLType()
		if !foldableType(t) || (i > 0 && !t.Equal(elemType)) {
			return wgsl.TypeInfo{}, nil, false
		}
		elemType = t
		values[i] = el.ConstValues()
		if len(values[i]) != typeWidth(t) {
			return wgsl.TypeInfo{}, nil, false
		}
	}
	return elemType, values, true
}

func typeWidth(t wgsl.TypeInfo) int {
	if t.IsNumericVector() {
		return t.VectorSize()
	}
	return 1
}

// countsUp reports whether the elements are the numbers 0, 1, 2, ..., so the
// loop's index is each one.
func countsUp(elemType wgsl.TypeInfo, values [][]float64) bool {
	if !elemType.IsNumericScalar() {
		return false
	}
	for i, v := range values {
		if v[0] != float64(i) {
			return false
		}
	}
	return true
}

// foldConstArray returns the name of a module-scope constant array of values,
// declaring it the first time any fold asks for those values.
func (d *Delegate) foldConstArray(elemType wgsl.TypeInfo, values [][]float64) string {
	var sb strings.Builder
	for i, v := range values {
		if i > 0 {
			sb.WriteString(", ")
		}
		if elemType.IsNumericScalar() {
			sb.WriteString(wgsl.CodegenValue(v[0]))
			continue
		}
		parts := make([]string, len(v))
		for j, x := range v {
			parts[j] = wgsl.CodegenValue(x)
		}
		fmt.Fprintf(&sb, "%s(%s)", elemType.Name, strings.Join(parts, ", "))
	}
	decl := fmt.Sprintf("array<%s, %d>(%s)", elemType.Name, len(values), sb.String())
	if d.foldConsts == nil {
		d.foldConsts = make(map[string]string)
	}
	if name, ok := d.foldConsts[decl]; ok {
		return name
	}
	name := d.uniqueTopLevelName("gsl_foldK")
	d.foldConsts[decl] = name
	d.foldConstDecls = append(d.foldConstDecls, fmt.Sprintf("const %s = %s;", name, decl))
	return name
}

type promotedVectorExpr struct {
	wgsl.Expression
	inner   wgsl.Expression
	vecType wgsl.TypeInfo
}

func (p *promotedVectorExpr) Codegen(ctx wgsl.CodegenContext) string {
	return fmt.Sprintf("%s(%s)", p.vecType.Name, p.inner.Codegen(ctx))
}

func (p *promotedVectorExpr) WGSLType() wgsl.TypeInfo {
	return p.vecType
}

func (p *promotedVectorExpr) Reads() []string {
	return p.inner.Reads()
}

func promoteScalarToVector(initExpr wgsl.Expression, vecType wgsl.TypeInfo) wgsl.Expression {
	size := vecType.VectorSize()
	if c, ok := initExpr.(build.ConstExpression); ok {
		cVals := c.ConstValues()
		if len(cVals) == 1 {
			parts := make([]string, size)
			vStr := wgsl.CodegenValue(cVals[0])
			for i := range parts {
				parts[i] = vStr
			}
			src := fmt.Sprintf("%s(%s)", vecType.Name, strings.Join(parts, ", "))
			return wgsl.NewOpaqueExpr(nil, vecType, build.PrecedenceHighest, src, initExpr.Reads()...)
		}
	}
	p := &promotedVectorExpr{
		inner:   initExpr,
		vecType: vecType,
	}
	p.Expression = wgsl.NewExprWithReads(nil, vecType, build.PrecedenceHighest, initExpr.Reads()...)
	p.Expression.SetOuter(p)
	return p
}
