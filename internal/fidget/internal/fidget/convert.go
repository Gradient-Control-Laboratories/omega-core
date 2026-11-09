package fidget

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

type Scope struct {
	parents []*Scope

	// map from variable name to fidget tree index
	nodes map[string][]TreeIndex

	// map from function name to map of arity -> ScopedFunc
	funcs map[string]map[int]*ScopedFunc

	// map from tree index to input
	inputs map[TreeIndex]Input
}

type Input struct {
	Value float64 `json:"value"`

	HasMin bool `json:"hasMin"`
	HasMax bool `json:"hasMax"`

	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type ScopedFunc struct {
	pbFunc *program.Function
	scope  *Scope
}

func (s *Scope) SetTime(t float64) {
	s.nodes[program.TimeVarName] = []TreeIndex{NewConst(t)}
}

func NewScope(parents []*Scope) *Scope {
	if len(parents) == 0 {
		parents = []*Scope{baseScope}
	}
	return &Scope{
		parents: parents,
		nodes:   make(map[string][]TreeIndex, 10),
		funcs:   make(map[string]map[int]*ScopedFunc, 10),
		inputs:  make(map[TreeIndex]Input, 10),
	}
}

var baseScope = &Scope{
	nodes: map[string][]TreeIndex{
		program.XVarName:    {NewX()},
		program.YVarName:    {NewY()},
		program.ZVarName:    {NewZ()},
		program.PVarName:    {NewX(), NewY(), NewZ()},
		program.TimeVarName: {NewConst(0.0)},
	},
	funcs:  map[string]map[int]*ScopedFunc{},
	inputs: map[TreeIndex]Input{},
}

func (s *Scope) NodeNamed(name string) ([]TreeIndex, bool) {
	treeIndices, ok := s.nodes[name]
	if ok && len(treeIndices) > 0 {
		return treeIndices, true
	}
	if len(s.parents) > 0 {
		for _, parent := range s.parents {
			treeIndices, ok := parent.NodeNamed(name)
			if ok && len(treeIndices) > 0 {
				return treeIndices, true
			}
		}
	}
	return noTreeIndices, false
}

func (s *Scope) InputFor(treeIndex TreeIndex) *Input {
	input, ok := s.inputs[treeIndex]
	if ok {
		return &input
	}
	if len(s.parents) > 0 {
		for _, parent := range s.parents {
			input := parent.InputFor(treeIndex)
			if input != nil {
				return input
			}
		}
	}
	return nil
}

func (s *Scope) OneValueFromOperand(operand *program.Operand) (TreeIndex, error) {
	treeIndices, err := s.MultiValueFromOperand(operand)
	if err != nil {
		return zeroIndex, err
	}
	if len(treeIndices) != 1 {
		return zeroIndex, fmt.Errorf("expected 1 value, but got %d", len(treeIndices))
	}
	return treeIndices[0], nil
}

func (s *Scope) MultiValueFromOperand(operand *program.Operand) ([]TreeIndex, error) {
	treeIndices, f, err := s.MultiValueOrFunctionFromOperand(operand)
	if err != nil {
		return noTreeIndices, err
	}
	if f != nil {
		var fName string
		ref := operand.GetReference()
		if ref != nil && ref.Name != "" {
			fName = fmt.Sprintf("function name: %s", ref.Name)
		} else {
			fName = "something else"
		}
		return noTreeIndices, fmt.Errorf("expected value, but got %s", fName)
	}
	return treeIndices, nil
}

func (s *Scope) MultiValueOrFunctionFromOperand(operand *program.Operand) ([]TreeIndex, *ScopedFunc, error) {
	switch t := operand.T.(type) {
	case *program.Operand_Value:
		return s.FromValue(t.Value)
	case *program.Operand_Reference:
		return s.FromReference(t.Reference)
	case *program.Operand_Expression:
		return s.FromExpression(t.Expression)
	case *program.Operand_FnCall:
		return s.FromFunctionCall(t.FnCall)
	}

	return noTreeIndices, nil, fmt.Errorf("not implemented %v", operand)
}

func (s *Scope) FromValue(value float64) ([]TreeIndex, *ScopedFunc, error) {
	return []TreeIndex{NewConst(value)}, nil, nil
}

func (s *Scope) FromReference(ref *program.Reference) ([]TreeIndex, *ScopedFunc, error) {
	name := ref.Name
	if ref.Prefix != "" {
		name = ref.Prefix + "." + ref.Name
	}
	treeIndices, ok := s.nodes[name]
	if ok {
		return treeIndices, nil, nil
	}
	if arityMap, ok := s.funcs[name]; ok {
		for _, f := range arityMap {
			return nil, f, nil
		}
	}

	if len(s.parents) > 0 {
		for _, parent := range s.parents {
			treeIndex, f, err := parent.FromReference(ref)
			if err == nil {
				return treeIndex, f, nil
			}
		}
	}

	return noTreeIndices, nil, fmt.Errorf("reference not found: %s", name)
}

func (s *Scope) FromExpression(expr *program.Expression) ([]TreeIndex, *ScopedFunc, error) {
	switch expr.Operation {

	// basic

	case program.Operation_ADD:
		return s.fromMulti("add", Add, expr)
	case program.Operation_SUB:
		return s.fromMulti("subtract", Sub, expr)
	case program.Operation_MUL:
		return s.fromMulti("multiply", Mul, expr)
	case program.Operation_DIV:
		return s.fromMulti("divide", Div, expr)

	// multi

	case program.Operation_MIN:
		return s.fromMulti("min", Min, expr)
	case program.Operation_MAX:
		return s.fromMulti("max", Max, expr)
	case program.Operation_AND:
		return s.fromMulti("and", And, expr)
	case program.Operation_OR:
		return s.fromMulti("or", Or, expr)

	// unary

	case program.Operation_SQUARE:
		return s.fromUnary("square", Square, expr)
	case program.Operation_FLOOR:
		return s.fromUnary("floor", Floor, expr)
	case program.Operation_CEIL:
		return s.fromUnary("ceiling", Ceil, expr)
	case program.Operation_ROUND:
		return s.fromUnary("round", Round, expr)
	case program.Operation_SQRT:
		return s.fromUnary("sqrt", Sqrt, expr)
	case program.Operation_NEG:
		return s.fromUnary("neg", Neg, expr)
	case program.Operation_SIN:
		return s.fromUnary("sin", Sin, expr)
	case program.Operation_COS:
		return s.fromUnary("cos", Cos, expr)
	case program.Operation_TAN:
		return s.fromUnary("tan", Tan, expr)
	case program.Operation_ASIN:
		return s.fromUnary("asin", Asin, expr)
	case program.Operation_ACOS:
		return s.fromUnary("acos", Acos, expr)
	case program.Operation_ATAN:
		return s.fromUnary("atan", Atan, expr)
	case program.Operation_EXP:
		return s.fromUnary("exp", Exp, expr)
	case program.Operation_LN:
		return s.fromUnary("ln", Ln, expr)
	case program.Operation_NOT:
		return s.fromUnary("not", Not, expr)
	case program.Operation_ABS:
		return s.fromUnary("abs", Abs, expr)

	// binary

	case program.Operation_COMPARE:
		return s.fromBinary("compare", Compare, expr)
	case program.Operation_MODULO:
		return s.fromBinary("modulo", Modulo, expr)
	case program.Operation_ATAN2:
		return s.fromBinary("atan2", Atan2, expr)
	case program.Operation_POW:
		// Two paths. If the exponent is a constant integer, use Tree::pow's
		// exponentiation by squaring — exact, fast, works on negative bases.
		// Otherwise synthesize pow(b, e) = exp(e * ln(b)) with the base
		// clamped, since negative bases would produce NaN bounds and panic
		// the mesher.
		if n, ok := constIntOperand(expr.Operands[1]); ok {
			return s.fromUnary("pow", func(base TreeIndex) TreeIndex {
				return PowInt(base, n)
			}, &program.Expression{Operation: expr.Operation, Operands: expr.Operands[:1]})
		}
		return s.fromBinary("pow", func(base, exponent TreeIndex) TreeIndex {
			return Exp(Mul(exponent, Ln(Max(base, NewConst(1e-12)))))
		}, expr)

	// comparisons (return 1 for true, 0 for false)

	case program.Operation_GT:
		return s.fromBinary("gt", func(a, b TreeIndex) TreeIndex {
			return Max(NewConst(0), Compare(a, b))
		}, expr)
	case program.Operation_GE:
		return s.fromBinary("ge", func(a, b TreeIndex) TreeIndex {
			return Not(Max(NewConst(0), Compare(b, a)))
		}, expr)
	case program.Operation_LT:
		return s.fromBinary("lt", func(a, b TreeIndex) TreeIndex {
			return Max(NewConst(0), Compare(b, a))
		}, expr)
	case program.Operation_LE:
		return s.fromBinary("le", func(a, b TreeIndex) TreeIndex {
			return Not(Max(NewConst(0), Compare(a, b)))
		}, expr)
	case program.Operation_EQ:
		return s.fromBinary("eq", func(a, b TreeIndex) TreeIndex {
			return Not(Abs(Compare(a, b)))
		}, expr)
	case program.Operation_NE:
		return s.fromBinary("ne", func(a, b TreeIndex) TreeIndex {
			return Abs(Compare(a, b))
		}, expr)

	// select (ternary)

	case program.Operation_SELECT:
		return s.fromSelect(expr)

	// access and indexing

	case program.Operation_ACCESS:
		return s.fromAccess(expr)
	case program.Operation_INDEX:
		return s.fromIndex(expr)

	// fidget

	case program.Operation_REMAP:
		return s.fromRemap(expr)

	// custom

	case program.Operation_MAP:
		return s.fromMap(expr)
	case program.Operation_REDUCE:
		return s.fromReduce(expr)
	case program.Operation_COLLECT:
		return s.fromCollect(expr)
	}

	return noTreeIndices, nil, fmt.Errorf("not implemented: %v", expr)
}

// fromSelect handles ternary select: condition ? trueResult : falseResult
// Implemented as: cond * a + (1 - cond) * b where cond is 0 or 1
func (s *Scope) fromSelect(expr *program.Expression) ([]TreeIndex, *ScopedFunc, error) {
	operands := expr.Operands
	if len(operands) != 3 && len(operands) != 4 {
		return noTreeIndices, nil, fmt.Errorf("select requires 3 or 4 operands, got %d", len(operands))
	}

	if len(operands) == 3 {
		// condition ? trueResult : falseResult
		cond, err := s.OneValueFromOperand(operands[0])
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("select condition: %v", err)
		}
		trueVal, err := s.MultiValueFromOperand(operands[1])
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("select true value: %v", err)
		}
		falseVal, err := s.MultiValueFromOperand(operands[2])
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("select false value: %v", err)
		}

		// broadcast
		if len(trueVal) != len(falseVal) {
			if len(trueVal) == 1 {
				trueVal = expand(trueVal[0], len(falseVal))
			} else if len(falseVal) == 1 {
				falseVal = expand(falseVal[0], len(trueVal))
			} else {
				return noTreeIndices, nil, fmt.Errorf("select: true and false branches have different lengths (%d vs %d)",
					len(trueVal), len(falseVal))
			}
		}

		// cond ? a : b = cond * a + (1 - cond) * b, exact for cond 0 or 1:
		// b + cond * (a - b) would round a - b, and with it a, when b is large.
		one := NewConst(1)
		result := make([]TreeIndex, len(trueVal))
		for i := range trueVal {
			result[i] = Add(Mul(cond, trueVal[i]), Mul(Sub(one, cond), falseVal[i]))
		}
		return result, nil, nil
	}

	// Select the positive, zero, or negative branch.
	value, err := s.OneValueFromOperand(operands[0])
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("select3 value: %v", err)
	}
	ltVal, err := s.MultiValueFromOperand(operands[3])
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("select3 lt value: %v", err)
	}
	eqVal, err := s.MultiValueFromOperand(operands[2])
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("select3 eq value: %v", err)
	}
	gtVal, err := s.MultiValueFromOperand(operands[1])
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("select3 gt value: %v", err)
	}

	// broadcast all to same length
	maxLen := len(ltVal)
	if len(eqVal) > maxLen {
		maxLen = len(eqVal)
	}
	if len(gtVal) > maxLen {
		maxLen = len(gtVal)
	}
	if len(ltVal) == 1 && maxLen > 1 {
		ltVal = expand(ltVal[0], maxLen)
	}
	if len(eqVal) == 1 && maxLen > 1 {
		eqVal = expand(eqVal[0], maxLen)
	}
	if len(gtVal) == 1 && maxLen > 1 {
		gtVal = expand(gtVal[0], maxLen)
	}
	if len(ltVal) != maxLen || len(eqVal) != maxLen || len(gtVal) != maxLen {
		return noTreeIndices, nil, fmt.Errorf("select3: branch lengths are incompatible (%d, %d, %d)",
			len(ltVal), len(eqVal), len(gtVal))
	}

	zero := NewConst(0)
	isGt := Max(zero, Compare(value, zero))
	isLt := Max(zero, Compare(zero, value))

	result := make([]TreeIndex, maxLen)
	for i := range result {
		ltTerm := Mul(isLt, Sub(ltVal[i], eqVal[i]))
		gtTerm := Mul(isGt, Sub(gtVal[i], eqVal[i]))
		result[i] = Add(eqVal[i], Add(ltTerm, gtTerm))
	}
	return result, nil, nil
}

// fromAccess handles member access like p.x, p.y, p.z, and swizzles like p.xy
func (s *Scope) fromAccess(expr *program.Expression) ([]TreeIndex, *ScopedFunc, error) {
	operands := expr.Operands
	if len(operands) != 2 {
		return noTreeIndices, nil, fmt.Errorf("access requires 2 operands, got %d", len(operands))
	}

	target, err := s.MultiValueFromOperand(operands[0])
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("access target: %v", err)
	}

	ref := operands[1].GetReference()
	if ref == nil || ref.Name == "" {
		return noTreeIndices, nil, fmt.Errorf("access member must be a reference")
	}
	member := ref.Name

	result := make([]TreeIndex, 0, len(member))
	for _, ch := range member {
		var idx int
		switch ch {
		case 'x':
			idx = 0
		case 'y':
			idx = 1
		case 'z':
			idx = 2
		case 'w':
			idx = 3
		default:
			return noTreeIndices, nil, fmt.Errorf("access: unknown component '%c'", ch)
		}
		if idx >= len(target) {
			return noTreeIndices, nil, fmt.Errorf("access: component '%c' (index %d) out of range for %d-element value", ch, idx, len(target))
		}
		result = append(result, target[idx])
	}

	return result, nil, nil
}

// fromIndex handles list indexing like b[i]
func (s *Scope) fromIndex(expr *program.Expression) ([]TreeIndex, *ScopedFunc, error) {
	operands := expr.Operands
	if len(operands) != 2 {
		return noTreeIndices, nil, fmt.Errorf("index requires 2 operands, got %d", len(operands))
	}
	target, err := s.MultiValueFromOperand(operands[0])
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("index target: %v", err)
	}
	var idx int
	if len(operands[1].ConstValues) > 0 {
		idx = int(operands[1].ConstValues[0])
	} else {
		idxTree, err := s.OneValueFromOperand(operands[1])
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("index: %v", err)
		}
		c, ok := constValues[idxTree]
		if !ok {
			return noTreeIndices, nil, fmt.Errorf("index must be constant in fidget, got tree %d", idxTree)
		}
		idx = int(c)
	}
	if idx < 0 || idx >= len(target) {
		return noTreeIndices, nil, fmt.Errorf("index %d out of range for %d-element value", idx, len(target))
	}
	return []TreeIndex{target[idx]}, nil, nil
}

func (s *Scope) fromUnary(fName string, f func(arg1 TreeIndex) TreeIndex, expr *program.Expression) ([]TreeIndex, *ScopedFunc, error) {
	operands := expr.Operands
	if len(operands) != 1 {
		return noTreeIndices, nil, fmt.Errorf("expected 1 argument, but got %d", len(operands))
	}

	arg1, err := s.MultiValueFromOperand(operands[0])
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("\"%s\" operation arg 0: %v", fName, err)
	}

	// Perform the element-wise operation
	result := make([]TreeIndex, len(arg1))
	for index, treeIndex := range arg1 {
		result[index] = f(treeIndex)
	}

	return result, nil, nil
}

func (s *Scope) fromBinary(fName string, f func(arg1 TreeIndex, arg2 TreeIndex) TreeIndex, expr *program.Expression) ([]TreeIndex, *ScopedFunc, error) {
	operands := expr.Operands
	if len(operands) != 2 {
		return noTreeIndices, nil, fmt.Errorf("expected 2 arguments, but got %d", len(operands))
	}

	arg1, err := s.MultiValueFromOperand(operands[0])
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("\"%s\" operation arg 0: %v", fName, err)
	}
	arg2, err := s.MultiValueFromOperand(operands[1])
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("\"%s\" operation arg 1: %v", fName, err)
	}
	if len(arg2) != len(arg1) {
		if len(arg1) == 1 {
			arg1 = expand(arg1[0], len(arg2))
		} else if len(arg2) == 1 {
			arg2 = expand(arg2[0], len(arg1))
		} else {
			return noTreeIndices, nil, fmt.Errorf("\"%s\" operation expected arg %d to have %d values, but got %d",
				fName, 1, len(arg1), len(arg2))
		}
	}

	// Perform the element-wise operation
	count := len(arg1)
	result := make([]TreeIndex, count)
	for i := 0; i < count; i++ {
		result[i] = f(arg1[i], arg2[i])
	}

	return result, nil, nil
}

// This implements operations using "broadcasting" similar to NumPy, Matlab, and PyTorch
//   - https://numpy.org/doc/stable/user/basics.broadcasting.html
//   - https://www.mathworks.com/help/matlab/matlab_prog/compatible-array-sizes-for-basic-operations.html
//   - https://pytorch.org/docs/stable/notes/broadcasting.html
//
// and Element-wise Product (Hadamard Product)
//   - https://en.wikipedia.org/wiki/Hadamard_product_(matrices)
func (s *Scope) fromMulti(fName string, f func(arg1 TreeIndex, arg2 TreeIndex) TreeIndex, expr *program.Expression) ([]TreeIndex, *ScopedFunc, error) {
	operands := expr.Operands
	if len(operands) < 1 {
		return noTreeIndices, nil, fmt.Errorf("expected at least 1 argument, but got %d", len(operands))
	}

	result, err := s.MultiValueFromOperand(operands[0])
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("\"%s\" operation arg 0: %v", fName, err)
	}
	for argIndex := 1; argIndex < len(operands); argIndex++ {
		arg2, err := s.MultiValueFromOperand(operands[argIndex])
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("\"%s\" operation arg %d: %v", fName, argIndex, err)
		}

		// Check that the two arguments are of the same length
		// expanding ("broadcasting") single element lists as needed
		arg1 := result
		if len(arg2) != len(arg1) {
			if len(arg1) == 1 {
				arg1 = expand(arg1[0], len(arg2))
			} else if len(arg2) == 1 {
				arg2 = expand(arg2[0], len(arg1))
			} else {
				return noTreeIndices, nil, fmt.Errorf("\"%s\" operation expected arg %d to have %d values, but got %d",
					fName, argIndex, len(arg1), len(arg2))
			}
		}

		// Perform the element-wise operation
		count := len(arg1)
		result = make([]TreeIndex, count)
		for i := 0; i < count; i++ {
			result[i] = f(arg1[i], arg2[i])
		}
	}
	return result, nil, nil
}

// Return a new slice of [count] size filled entirely with [value]
func expand(value TreeIndex, count int) []TreeIndex {
	result := make([]TreeIndex, count)
	for i := 0; i < count; i++ {
		result[i] = value
	}
	return result
}

func (s *Scope) FromFunctionCall(call *program.FnCall) ([]TreeIndex, *ScopedFunc, error) {
	// Check builtins first
	if result, err, handled := s.fromBuiltinCall(call); handled {
		return result, nil, err
	}

	arity := len(call.Arguments)
	f := s.findFunc(call.Reference.Name, arity)
	if f == nil {
		return noTreeIndices, nil, fmt.Errorf("function not found: %s with %d args", call.Reference.Name, arity)
	}
	pbF := f.pbFunc

	// Verify that function call arguments match function parameters
	arguments := call.Arguments
	numArguments := len(arguments)
	numParameters := len(pbF.Parameters)
	if numArguments != numParameters {
		return noTreeIndices, nil, fmt.Errorf("call \"%s\" expected %d arguments, but got %d",
			call.Reference.Name, numParameters, numArguments)
	}

	// Process function call arguments in the current scope
	// and place the results into a new call scope
	callScope := NewScope([]*Scope{f.scope})
	for index := 0; index < numParameters; index++ {
		paramName := pbF.Parameters[index].Name
		arg := arguments[index]
		treeIndices, f, err := s.MultiValueOrFunctionFromOperand(arg)
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("call \"%s\" param %d \"%s\": %v",
				call.Reference.Name, index, paramName, err)
		}
		if f != nil {
			if callScope.funcs[paramName] == nil {
				callScope.funcs[paramName] = make(map[int]*ScopedFunc)
			}
			paramArity := len(f.pbFunc.Parameters)
			callScope.funcs[paramName][paramArity] = f
		} else {
			callScope.nodes[paramName] = treeIndices
		}
	}

	// Register nested functions if any
	for _, fnType := range pbF.FnTypes {
		for _, group := range fnType.Groups {
			for _, fn := range group.Functions {
				err := callScope.registerFunction(fnType.Name, fn)
				if err != nil {
					return noTreeIndices, nil, fmt.Errorf("func \"%s\" nested function \"%s\": %v", call.Reference.Name, fnType.Name, err)
				}
			}
		}
	}

	// Process the assignments and returns in the new scope
	err := callScope.EvaluateAssignments(pbF.Assignments)
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("func \"%s\" assignments: %v", call.Reference.Name, err)
	}
	treeIndices, rf, rErr := callScope.EvaluateReturns(pbF.Returns)
	if rErr != nil {
		return noTreeIndices, nil, fmt.Errorf("func \"%s\" returns: %v", call.Reference.Name, rErr)
	}
	return treeIndices, rf, rErr
}

func (s *Scope) findFunc(name string, arity int) *ScopedFunc {
	if arityMap, ok := s.funcs[name]; ok {
		if f, ok := arityMap[arity]; ok {
			return f
		}
	}
	for _, parent := range s.parents {
		f := parent.findFunc(name, arity)
		if f != nil {
			return f
		}
	}
	return nil
}

func (s *Scope) collectVarargs(arguments []*program.Operand, start int) ([]TreeIndex, error) {
	numArguments := len(arguments)
	treeIndices, err := s.MultiValueFromOperand(arguments[start])
	if err != nil {
		return noTreeIndices, err
	}
	if len(treeIndices) == 0 {
		return noTreeIndices, fmt.Errorf("expected at least 1 vararg, but got 0")
	}
	if len(treeIndices) == 1 {
		firstIndex := treeIndices[0]
		treeIndices = make([]TreeIndex, 0, numArguments-start)
		treeIndices = append(treeIndices, firstIndex)
		for index := start + 1; index < numArguments; index++ {
			arg := arguments[index]
			treeIndex, err := s.OneValueFromOperand(arg)
			if err != nil {
				return noTreeIndices, err
			}
			treeIndices = append(treeIndices, treeIndex)
		}
	} else {
		if numArguments > start+1 {
			return noTreeIndices, fmt.Errorf("expected either 1 varargs or multiple arguments," +
				" but got varargs followed by one more more arguments")
		}
	}
	return treeIndices, nil
}

func (s *Scope) EvaluatePackage(pkg *program.Package) error {
	for _, file := range pkg.Files {
		for _, fnType := range file.FnTypes {
			for _, group := range fnType.Groups {
				for _, fn := range group.Functions {
					err := s.registerFunction(fnType.Name, fn)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	for _, file := range pkg.Files {
		err := s.EvaluateAssignments(file.Assignments)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Scope) EvaluateAssignments(assignments []*program.Assignment) error {
	for _, assign := range assignments {
		err := s.evaluateAndCacheAssignment(assign)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Scope) evaluateAndCacheAssignment(assign *program.Assignment) error {
	if len(assign.Variables) == 0 {
		return fmt.Errorf("expected at least 1 variable in assignment, but got 0")
	}

	names := make([]string, len(assign.Variables))
	for i, v := range assign.Variables {
		names[i] = v.Name
		err := s.validateName(v.Name)
		if err != nil {
			return fmt.Errorf("assignment \"%s\": %w", v.Name, err)
		}
	}

	// Extract min/max input constraints from the variable properties if it's an input
	isInput := false
	var minProp, maxProp *program.Property
	if len(assign.Variables) == 1 {
		v := assign.Variables[0]
		for _, prop := range v.Properties {
			switch prop.Name {
			case "input":
				isInput = true
			case "min":
				minProp = prop
			case "max":
				maxProp = prop
			}
		}
	}

	if isInput {
		name := names[0]
		// Get default value
		var val float64
		if assign.Expression != nil {
			if vVal, ok := assign.Expression.T.(*program.Operand_Value); ok {
				val = vVal.Value
			}
		}

		treeIndex := NewInput(name, val)
		s.nodes[name] = []TreeIndex{treeIndex}
		fgInput := Input{Value: val}

		if minProp != nil {
			if vVal, ok := minProp.Expression.T.(*program.Operand_Value); ok {
				fgInput.HasMin = true
				fgInput.Min = vVal.Value
			}
		}
		if maxProp != nil {
			if vVal, ok := maxProp.Expression.T.(*program.Operand_Value); ok {
				fgInput.HasMax = true
				fgInput.Max = vVal.Value
			}
		}
		s.inputs[treeIndex] = fgInput
		return nil
	}

	// Normal assignment evaluation
	treeIndices, err := s.MultiValueFromOperand(assign.Expression)
	if err != nil {
		return fmt.Errorf("assignment \"%s\": %v", names[0], err)
	}
	if len(treeIndices) == 0 {
		return fmt.Errorf("assignment \"%s\": expected at least 1 value, but got 0", names[0])
	}

	if len(names) == len(treeIndices) {
		for index, name := range names {
			s.nodes[name] = []TreeIndex{treeIndices[index]}
		}
	} else if len(names) == 1 {
		s.nodes[names[0]] = treeIndices
	} else {
		return fmt.Errorf("assignment \"%s\": got %d name(s), but %d value(s)", names[0], len(names), len(treeIndices))
	}
	return nil
}

func (s *Scope) registerFunction(name string, fn *program.Function) error {
	err := s.validateName(name)
	if err != nil {
		return err
	}
	if s.funcs[name] == nil {
		s.funcs[name] = make(map[int]*ScopedFunc)
	}
	arity := len(fn.Parameters)
	if _, existing := s.funcs[name][arity]; existing {
		return fmt.Errorf("function %s with %d parameters already defined", name, arity)
	}
	s.funcs[name][arity] = &ScopedFunc{fn, s}
	return nil
}

func (s *Scope) EvaluateReturns(returns []*program.Operand) ([]TreeIndex, *ScopedFunc, error) {
	if len(returns) == 0 {
		return noTreeIndices, nil, nil
	}

	treeIndices, err := s.MultiValueFromOperand(returns[0])
	if err != nil {
		return noTreeIndices, nil, err
	}

	if len(returns) == 1 {
		if len(treeIndices) == 0 {
			return noTreeIndices, nil, fmt.Errorf("expected at least 1 return value, but got 0")
		}
		return treeIndices, nil, nil
	}

	if len(treeIndices) != 1 {
		return noTreeIndices, nil, fmt.Errorf("expected 1 return value, but got %d", len(treeIndices))
	}

	firstIndex := treeIndices[0]
	treeIndices = make([]TreeIndex, 0, len(returns))

	treeIndices = append(treeIndices, firstIndex)
	for index := 1; index < len(returns); index++ {
		treeIndex, err := s.OneValueFromOperand(returns[index])
		if err != nil {
			return noTreeIndices, nil, err
		}
		treeIndices = append(treeIndices, treeIndex)
	}

	return treeIndices, nil, nil
}

func (s *Scope) validateName(name string) error {
	// Assume that the parser has already validated the identifier as valid

	// Verify that the name has not already been used as a variable
	_, existing := s.nodes[name]
	if existing {
		return fmt.Errorf("variable already defined with name: %s", name)
	}

	// TODO warn if name shadows a name defined in a parent scope

	return nil
}

var zeroIndex = TreeIndex(0)
var noTreeIndices = []TreeIndex{}

// constIntOperand returns (n, true) when the operand is a literal whose value
// is an integer in the i64 range. Used to take the native Tree::pow fast path
// for integer exponents.
func constIntOperand(operand *program.Operand) (int64, bool) {
	v, ok := operand.GetT().(*program.Operand_Value)
	if !ok {
		return 0, false
	}
	n := int64(v.Value)
	if float64(n) != v.Value {
		return 0, false
	}
	return n, true
}
