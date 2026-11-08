package codegen

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/internal"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// Implement all required Delegate interface methods

func (d *Delegate) Remap(remapOp *program.Operand, targetOp *program.Operand, pExpr build.Expression, targetExpr build.Expression) (build.Expression, error) {
	// Remap is a graphics operation; return 0 in audio context so mixed graphics+audio
	// programs don't fail to compile.
	return newExpression("0"), nil
}

func (d *Delegate) NewFnVar(name string, decl *program.FnType, enclosingScope *build.Scope) build.VarInfo {
	// VarInfo.Name() stays the GSL name (scope lookups key on it). The
	// emitted JS identifier — possibly mangled with the import pkg path —
	// rides in expr so BuildVarRef returns it directly.
	jsName := d.pkgQualifiedName(enclosingScope, name)
	expr := newExpression(jsName)
	return &simpleVarInfo{bScope: enclosingScope, name: name, typ: nil, expr: expr}
}

// jsFnInfo implements build.FnInfo for user-defined JavaScript functions.
// `name` is the GSL name. `jsName` is the emitted JS identifier — possibly
// mangled with a pkg prefix so imports don't collide with each other or with
// the primary package.
type jsFnInfo struct {
	name           string
	jsName         string
	paramNames     []string
	returnCode     string
	returnType     *program.Type
	varMapSnapshot map[string]build.VarInfo // varMap state before this function was compiled
}

// Ensure that jsFnInfo implements the build.FnInfo interface.
var _ build.FnInfo = (*jsFnInfo)(nil)

func (f *jsFnInfo) Name() string { return f.name }
func (f *jsFnInfo) ReturnType() *program.Type {
	if f.returnType != nil {
		return f.returnType
	}
	return jsNumType
}

func (f *jsFnInfo) ParamTypes() []*program.Type {
	types := make([]*program.Type, len(f.paramNames))
	for i := range types {
		types[i] = jsNumType
	}
	return types
}

func (d *Delegate) BuildFunc(
	fnDecl *program.FnType, fn *program.Function,
	pkgName string, args []build.Expression,
) (build.FnInfo, []build.VarInfo, error) {
	if fnDecl.Name == "" {
		return nil, nil, fmt.Errorf("function has no name")
	}
	name := fnDecl.Name
	declBScope := d.GetScope().Parent()
	jsName := d.pkgQualifiedName(declBScope, name)

	params := make([]build.VarInfo, len(fn.Parameters))
	paramNames := make([]string, len(fn.Parameters))
	for i, p := range fn.Parameters {
		paramNames[i] = p.Name
		// Carry the argument's type onto the parameter so collection-valued
		// arguments stay typed as vectors inside the body (otherwise indexing
		// or swizzling them fails the "must be a collection" check). Both the
		// VarInfo and its reference expression must carry the type.
		var paramType *program.Type = jsNumType
		if i < len(args) && args[i] != nil {
			paramType = args[i].Type()
		}
		param := &simpleVarInfo{
			bScope: d.currBScope,
			name:   p.Name,
			typ:    paramType,
			expr:   &jsExpression{code: p.Name, precedence: build.PrecedenceHighest, typ: paramType},
		}
		params[i] = param
	}

	// First time: snapshot varMap, register params, and push a new frame for this function
	snapshot := make(map[string]build.VarInfo, len(d.varMap))
	for k, v := range d.varMap {
		snapshot[k] = v
	}
	for _, p := range params {
		d.varMap[p.Name()] = p
	}

	d.pushFrame(name, paramNames)

	return &jsFnInfo{
		name:           name,
		jsName:         jsName,
		paramNames:     paramNames,
		varMapSnapshot: snapshot,
	}, params, nil
}

func (d *Delegate) BuildFuncReturn(fnInfo build.FnInfo, returnOps []*program.Operand, returnExprs []build.Expression) (build.FnInfo, error) {
	info, ok := fnInfo.(*jsFnInfo)
	if !ok {
		err := fmt.Errorf("expected *jsFnInfo, got %T", fnInfo)
		return nil, internal.Err(err)
	}

	// Capture return expression
	returnCode := ""
	if len(returnExprs) > 0 {
		returnCode = getCode(returnExprs[0])
		info.returnType = returnExprs[0].Type()
	} else {
		info.returnType = jsNumType
	}

	// Generate the JS function definition
	jsParams := make([]string, len(info.paramNames))
	for i, p := range info.paramNames {
		jsParams[i] = toJSIdentifier(p)
	}

	// Pop this function's frame to get its accumulated body
	frame := d.popFrame()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("function %s(%s) {\n", info.jsName, joinArgs(jsParams)))
	if body := strings.TrimSpace(frame.body.String()); body != "" {
		sb.WriteString(body)
		sb.WriteString("\n")
	}
	if returnCode != "" {
		sb.WriteString(fmt.Sprintf("    return %s;\n", returnCode))
	}
	sb.WriteString("}")
	d.fnDefs.WriteString(sb.String())
	d.fnDefs.WriteString("\n")

	// Restore varMap to the snapshot taken before compilation began, removing all
	// function-local variables (parameters and body-level declarations).
	if info.varMapSnapshot != nil {
		d.varMap = info.varMapSnapshot
	}
	d.builtFns[info.jsName] = true

	info.returnCode = returnCode
	return info, nil
}

func (d *Delegate) BuildFuncCall(fnCallOp *program.Operand, fnCall *program.FnCall, fnInfo build.FnInfo, args []build.Expression) (build.Expression, error) {
	info, ok := fnInfo.(*jsFnInfo)
	if !ok {
		err := fmt.Errorf("expected *jsFnInfo, got %T", fnInfo)
		return nil, internal.Err(err)
	}

	argCodes := make([]string, len(args))
	for i, arg := range args {
		argCodes[i] = getCode(arg)
	}

	code := fmt.Sprintf("%s(%s)", info.jsName, joinArgs(argCodes))
	return &jsExpression{code: code, precedence: build.PrecedenceLowest, typ: info.ReturnType()}, nil
}

func (d *Delegate) BuildBuiltinCall(op *program.Operand, fnCall *program.FnCall, builtin *program.BuiltinType, args []build.Expression) (build.Expression, error) {
	d.complexityScore += 2

	builtins := map[string]string{
		"clamp": "clamp", "fract": "fract", "sign": "sign",
		"sinh": "sinh", "cosh": "cosh", "tanh": "tanh",
		"asinh": "asinh", "acosh": "acosh", "atanh": "atanh",
		"exp2": "exp2", "trunc": "trunc",
		"radians": "radians", "degrees": "degrees",
		"normalize": "normalize", "length": "length",
		"distance": "distance", "dot": "dot", "cross": "cross",
	}

	if jsName, ok := builtins[builtin.Name]; ok {
		argCodes := make([]string, len(args))
		for i, arg := range args {
			argCodes[i] = getCode(arg)
		}
		code := fmt.Sprintf("gsl_%s(%s)", jsName, joinArgs(argCodes))
		return newExpression(code), nil
	}

	return nil, fmt.Errorf("unsupported builtin function for audio: %s", builtin.Name)
}

// simpleVarInfo implements build.VarInfo for built-in variables like time.
// Name() returns the GSL name so the common build framework's scope lookups
// resolve correctly (s.localVars is keyed by Name()). Codegen pulls its
// identifier out of expr instead, which carries the mangled JS name for
// imported-package decls.
type simpleVarInfo struct {
	bScope *build.Scope
	name   string
	typ    *program.Type
	expr   build.Expression
}

// Ensure that simpleVarInfo implements the build.VarInfo interface.
var _ build.VarInfo = (*simpleVarInfo)(nil)

func (v *simpleVarInfo) Name() string {
	return v.name
}

func (v *simpleVarInfo) Scope() *build.Scope {
	return v.bScope
}

func (v *simpleVarInfo) Expr() build.Expression {
	if v.expr != nil {
		return v.expr
	}
	return newExpression(v.name)
}

func (v *simpleVarInfo) Type() *program.Type {
	return v.typ
}

func (d *Delegate) BuildVarRef(op *program.Operand, varInfo build.VarInfo) (build.Expression, error) {
	if varInfo == nil {
		return nil, fmt.Errorf("variable not found")
	}
	return varInfo.Expr(), nil
}

func (d *Delegate) BuildSingleAssignment(decl *program.Assignment, op *program.Operand, name string, expr build.Expression) (build.VarInfo, error) {
	jsName := toJSIdentifier(name)
	code := getCode(expr)

	if d.inFunction() {
		d.currentFrame().body.WriteString(fmt.Sprintf("    const %s = %s;\n", jsName, code))
	} else {
		// Top-level decls in imported packages get pkg-qualified names so
		// two imports declaring the same constant don't collide.
		jsName = d.pkgQualifiedName(d.currBScope, jsName)
		d.topDecls.WriteString(fmt.Sprintf("    const %s = %s;\n", jsName, code))
	}
	d.complexityScore++

	// VarInfo.Name() stays the GSL name (scope lookups key on it). The
	// possibly-mangled JS identifier rides in expr.
	var constVals []float64
	if ce, ok := expr.(build.ConstExpression); ok {
		constVals = ce.ConstValues()
	}
	varInfo := &simpleVarInfo{
		bScope: d.currBScope,
		name:   name,
		typ:    expr.Type(),
		expr:   &jsExpression{code: jsName, precedence: build.PrecedenceHighest, typ: expr.Type(), constValues: constVals},
	}
	return varInfo, nil
}

func (d *Delegate) BuildMultiAssignment(decl *program.Assignment, op *program.Operand, fnCall *program.FnCall, names []string, expr build.Expression) ([]build.VarInfo, error) {
	return nil, fmt.Errorf("multi-assignment not supported in audio output")
}

func (d *Delegate) BuildValue(op *program.Operand, value float64) (build.ConstExpression, error) {
	code := fmt.Sprintf("%g", value)
	expr := newExpression(code)
	expr.constValues = []float64{value}
	return expr, nil
}

var memberIndex = map[byte]int{'x': 0, 'y': 1, 'z': 2, 'w': 3}

func (d *Delegate) BuildMemberAccess(op *program.Operand, expr *program.Expression, target build.Expression, member string) (build.Expression, error) {
	// For a single-component swizzle, use array indexing: JS arrays have no .x/.y/.z.
	if len(member) == 1 {
		if idx, ok := memberIndex[member[0]]; ok {
			targetCode := getCode(target)
			// p is [0,0,0] in audio context; its components are always 0.
			if targetCode == "[0, 0, 0]" {
				return newExpression("0"), nil
			}
			return newExpression(fmt.Sprintf("%s[%d]", targetCode, idx)), nil
		}
	}
	// Multi-component swizzle (.xy, .zyx, etc.): gather the selected
	// components into a new array and keep it typed as a vector so further
	// indexing/swizzling type-checks.
	targetCode := getCode(target)
	components := make([]string, len(member))
	for i := 0; i < len(member); i++ {
		idx, ok := memberIndex[member[i]]
		if !ok {
			return newExpression("0"), nil
		}
		components[i] = fmt.Sprintf("%s[%d]", targetCode, idx)
	}
	code := fmt.Sprintf("[%s]", strings.Join(components, ", "))
	return &jsExpression{code: code, precedence: build.PrecedenceHighest, typ: program.NewListType(int32(len(member)), program.NewScalarType())}, nil
}

func (d *Delegate) BuildIndexAccess(op *program.Operand, expr *program.Expression, targetOp *program.Operand, indexOp *program.Operand, target build.Expression, index int) (build.Expression, error) {
	code := fmt.Sprintf("%s[%d]", getCode(target), index)
	// Indexing a list of vectors yields the element vector, so keep the
	// element type for further indexing/swizzling.
	if target.Type().IsMatrix() {
		return &jsExpression{code: code, precedence: build.PrecedenceHighest, typ: target.Type().ElemType}, nil
	}
	return newExpression(code), nil
}

func (d *Delegate) BuildOperation(op *program.Operand, expr *program.Expression, symbol string, operands []*program.Operand, expressions []build.Expression, resultType *program.Type, precedence build.Precedence) (build.Expression, error) {
	d.complexityScore++

	switch expr.Operation {
	case program.Operation_ADD:
		return d.binaryOp("+", expressions)
	case program.Operation_SUB:
		return d.binaryOp("-", expressions)
	case program.Operation_MUL:
		return d.binaryOp("*", expressions)
	case program.Operation_DIV:
		return d.binaryOp("/", expressions)
	case program.Operation_MODULO:
		return d.binaryOp("%", expressions)

	case program.Operation_SIN:
		return d.mathFunc("sin", expressions)
	case program.Operation_COS:
		return d.mathFunc("cos", expressions)
	case program.Operation_TAN:
		return d.mathFunc("tan", expressions)
	case program.Operation_ASIN:
		return d.mathFunc("asin", expressions)
	case program.Operation_ACOS:
		return d.mathFunc("acos", expressions)
	case program.Operation_ATAN:
		return d.mathFunc("atan", expressions)
	case program.Operation_ATAN2:
		return d.mathFunc("atan2", expressions)

	case program.Operation_SQRT:
		return d.mathFunc("sqrt", expressions)
	case program.Operation_ABS:
		return d.mathFunc("abs", expressions)
	case program.Operation_FLOOR:
		return d.mathFunc("floor", expressions)
	case program.Operation_CEIL:
		return d.mathFunc("ceil", expressions)
	case program.Operation_ROUND:
		return d.mathFunc("round", expressions)

	case program.Operation_EXP:
		return d.mathFunc("exp", expressions)
	case program.Operation_LN:
		return d.mathFunc("log", expressions)
	case program.Operation_POW:
		return d.mathFunc("pow", expressions)

	case program.Operation_MIN:
		return d.mathFunc("min", expressions)
	case program.Operation_MAX:
		return d.mathFunc("max", expressions)

	case program.Operation_SQUARE:
		if len(expressions) != 1 {
			return nil, fmt.Errorf("square requires 1 operand")
		}
		e := expressions[0].(*jsExpression).ParenIf(build.PrecedenceMultiplication)
		code := fmt.Sprintf("%s * %s", e.Code(), e.Code())
		return &jsExpression{code: code, precedence: build.PrecedenceMultiplication, typ: jsNumType}, nil

	case program.Operation_AND:
		return d.binaryOp("&&", expressions)
	case program.Operation_OR:
		return d.binaryOp("||", expressions)

	default:
		return nil, fmt.Errorf("unsupported operation for audio: %v", expr.Operation)
	}
}

func (d *Delegate) BuildOpCall(op *program.Operand, expr *program.Expression, operands []*program.Operand, expressions []build.Expression, resultType *program.Type) (build.Expression, error) {
	return d.BuildOperation(op, expr, "", operands, expressions, resultType, 0)
}

func (d *Delegate) BuildNot(op *program.Operand, expr *program.Expression, targetOp *program.Operand, target build.Expression) (build.Expression, error) {
	targetExpr := target.(*jsExpression).ParenIf(build.PrecedenceUnary)
	code := fmt.Sprintf("!%s", targetExpr.Code())
	return &jsExpression{code: code, precedence: build.PrecedenceUnary, typ: target.Type()}, nil
}

func (d *Delegate) BuildNeg(op *program.Operand, expr *program.Expression, targetOp *program.Operand, target build.Expression) (build.Expression, error) {
	targetExpr := target.(*jsExpression).ParenIf(build.PrecedenceUnary)
	targetCode := targetExpr.Code()
	if strings.HasPrefix(targetCode, "-") {
		targetCode = " " + targetCode
	}
	code := fmt.Sprintf("-%s", targetCode)
	return &jsExpression{code: code, precedence: build.PrecedenceUnary, typ: target.Type()}, nil
}

func (d *Delegate) BuildComparison(op *program.Operand, expr *program.Expression, symbol string, operands []*program.Operand, expressions []build.Expression, resultType *program.Type, precedence build.Precedence) (build.Expression, error) {
	if len(expressions) != 2 {
		return nil, fmt.Errorf("comparison requires exactly 2 operands")
	}

	jsOp := ""
	switch expr.Operation {
	case program.Operation_GT:
		jsOp = ">"
	case program.Operation_GE:
		jsOp = ">="
	case program.Operation_LT:
		jsOp = "<"
	case program.Operation_LE:
		jsOp = "<="
	case program.Operation_EQ:
		jsOp = "==="
	case program.Operation_NE:
		jsOp = "!=="
	default:
		return nil, fmt.Errorf("unsupported comparison: %v", expr.Operation)
	}

	expr1 := expressions[0].(*jsExpression).ParenIf(precedence)
	expr2 := expressions[1].(*jsExpression)
	expr2Code := expr2.Code()
	if expr2.Precedence() > precedence {
		expr2Code = "(" + expr2Code + ")"
	}

	code := fmt.Sprintf("%s %s %s ? 1.0 : 0.0", expr1.Code(), jsOp, expr2Code)
	return &jsExpression{code: code, precedence: build.PrecedenceTernary, typ: jsNumType}, nil
}

func (d *Delegate) BuildCompare(op *program.Operand, expr *program.Expression, operands []*program.Operand, expr1 build.Expression, expr2 build.Expression, resultType *program.Type) (build.Expression, error) {
	e1 := expr1.(*jsExpression)
	e2 := expr2.(*jsExpression)
	e1Code := e1.Code()
	if e1.Precedence() > build.PrecedenceComparison {
		e1Code = "(" + e1Code + ")"
	}
	e2Code := e2.Code()
	if e2.Precedence() > build.PrecedenceComparison {
		e2Code = "(" + e2Code + ")"
	}
	code := fmt.Sprintf("%s > %s ? 1.0 : (%s === %s ? 0.0 : -1.0)", e1Code, e2Code, e1Code, e2Code)
	return &jsExpression{code: code, precedence: build.PrecedenceTernary, typ: jsNumType}, nil
}

func (d *Delegate) BuildSelect2(op *program.Operand, expr *program.Expression, operands []*program.Operand, conditionExpr build.Expression, trueResultExpr build.Expression, falseResultExpr build.Expression, resultType *program.Type) (build.Expression, error) {
	cond := conditionExpr.(*jsExpression)
	condCode := cond.Code()
	if cond.Precedence() > build.PrecedenceComparison {
		condCode = "(" + condCode + ")"
	}
	code := fmt.Sprintf("%s > 0 ? %s : %s", condCode, getCode(trueResultExpr), getCode(falseResultExpr))
	return &jsExpression{code: code, precedence: build.PrecedenceTernary, typ: resultType}, nil
}

func (d *Delegate) BuildSelect3(op *program.Operand, expr *program.Expression, operands []*program.Operand, valueExpr build.Expression, ltZeroResultExpr build.Expression, eqZeroResultExpr build.Expression, gtZeroResultExpr build.Expression, resultType *program.Type) (build.Expression, error) {
	val := valueExpr.(*jsExpression)
	valCode := val.Code()
	if val.Precedence() > build.PrecedenceComparison {
		valCode = "(" + valCode + ")"
	}
	code := fmt.Sprintf("%s > 0 ? %s : (%s === 0 ? %s : %s)",
		valCode, getCode(gtZeroResultExpr), valCode, getCode(eqZeroResultExpr), getCode(ltZeroResultExpr))
	return &jsExpression{code: code, precedence: build.PrecedenceTernary, typ: resultType}, nil
}

func (d *Delegate) BuildCollect(op *program.Operand, expr *program.Expression, operands []*program.Operand, expressions []build.Expression) (build.Expression, error) {
	args := make([]string, len(expressions))
	for i, e := range expressions {
		args[i] = getCode(e)
	}
	code := fmt.Sprintf("[%s]", joinArgs(args))
	if len(expressions) > 1 {
		// Vector elements gather into a list of vectors (a nested array).
		// The build layer ensures the elements are uniform.
		if expressions[0].Type().IsList() {
			elemType := expressions[0].Type()
			return &jsExpression{code: code, precedence: build.PrecedenceHighest,
				typ: program.NewMatrixType(int32(len(expressions)), elemType), elements: expressions}, nil
		}
		return &jsExpression{code: code, precedence: build.PrecedenceHighest, typ: program.NewListType(int32(len(expressions)), program.NewScalarType()), elements: expressions}, nil
	}
	res := newExpression(code)
	res.elements = expressions
	return res, nil
}

func (d *Delegate) BuildConstCollect(
	op *program.Operand,
	expr *program.Expression,
	operands []*program.Operand,
	expressions []build.Expression,
	constValues []float64,
) (build.ConstExpression, error) {
	collected, err := d.BuildCollect(op, expr, operands, expressions)
	if err != nil {
		return nil, err
	}
	jsExpr := collected.(*jsExpression)
	jsExpr.constValues = constValues
	return jsExpr, nil
}

// Helper functions

func (d *Delegate) binaryOp(op string, operands []build.Expression) (build.Expression, error) {
	if len(operands) < 2 {
		return nil, fmt.Errorf("binary operation requires at least 2 operands")
	}

	var precedence build.Precedence
	switch op {
	case "*", "/", "%":
		precedence = build.PrecedenceMultiplication
	case "+", "-":
		precedence = build.PrecedenceAddition
	case "&&":
		precedence = build.PrecedenceLogicalAnd
	case "||":
		precedence = build.PrecedenceLogicalOr
	default:
		precedence = build.PrecedenceLowest
	}

	wgslExprs := make([]*jsExpression, len(operands))
	wgslExprs[0] = operands[0].(*jsExpression).ParenIf(precedence)
	for i := 1; i < len(operands); i++ {
		e := operands[i].(*jsExpression)
		if e.Precedence() >= precedence {
			e = e.ParenIf(build.PrecedenceHighest)
		}
		wgslExprs[i] = e
	}

	src := wgslExprs[0].Code()
	for i := 1; i < len(wgslExprs); i++ {
		src += " " + op + " " + wgslExprs[i].Code()
	}

	// Preserve vec3 type if any operand is vec3 (e.g. p - [x,y,z] needed for remap validation)
	var typ *program.Type = jsNumType
	for _, operand := range operands {
		if operand.Type().IsList() {
			typ = operand.Type()
			break
		}
	}
	return &jsExpression{code: src, precedence: precedence, typ: typ}, nil
}

func (d *Delegate) mathFunc(name string, operands []build.Expression) (build.Expression, error) {
	if len(operands) == 0 {
		return nil, fmt.Errorf("math function requires at least 1 operand")
	}

	args := make([]string, len(operands))
	for i, op := range operands {
		args[i] = getCode(op)
	}

	code := fmt.Sprintf("Math.%s(%s)", name, joinArgs(args))

	// Preserve vector type if any operand is a vector (needed for type checking)
	for _, operand := range operands {
		if operand.Type().IsList() {
			return &jsExpression{code: code, precedence: build.PrecedenceHighest, typ: operand.Type()}, nil
		}
	}

	return newExpression(code), nil
}

func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += ", "
		}
		result += arg
	}
	return result
}

func (d *Delegate) NewParamVar(scope *build.Scope, name string, expr build.Expression, typ *program.Type) build.VarInfo {
	return &simpleVarInfo{
		bScope: scope,
		name:   name,
		typ:    typ,
		expr:   expr,
	}
}
