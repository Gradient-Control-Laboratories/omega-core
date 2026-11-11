package fidget

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

func (s *Scope) fromCollect(expr *program.Expression) ([]TreeIndex, *ScopedFunc, error) {
	operands := expr.Operands
	if len(operands) <= 1 {
		return noTreeIndices, nil, fmt.Errorf("\"collect\" operation expects at least 2 arguments, but got %d", len(operands))
	}
	result := make([]TreeIndex, 0, len(operands))
	for opIndex, op := range operands {
		treeIndex, err := s.OneValueFromOperand(op)
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("\"collect\" operation param %d: %v", opIndex, err)
		}
		result = append(result, treeIndex)
	}
	return result, nil, nil
}

// map(func(index, value), varargs)
// Apply the function to each value and return a new collection of values
func (s *Scope) fromMap(expr *program.Expression) ([]TreeIndex, *ScopedFunc, error) {
	operands := expr.Operands
	if len(operands) < 2 {
		return noTreeIndices, nil, fmt.Errorf("\"map\" operation expects at least 2 arguments, but got %d", len(operands))
	}
	mapF, err := s.namedOrAnonFunction(operands[0], -1, 1)
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("\"map\" operation function: %v", err)
	}
	varargs, err := s.collectVarargs(operands, 1)
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("\"map\" operation varargs: %v", err)
	}

	mapScope := mapF.scope
	mapPbF := mapF.pbFunc
	params := mapPbF.Parameters
	mapAssignments := mapPbF.Assignments
	mapReturns := mapPbF.Returns

	const numFixedParam = 1
	result := make([]TreeIndex, 0, len(varargs))
	for varArgIndex, treeIndex := range varargs {
		callScope := NewScope([]*Scope{mapScope})
		if len(params) == 1 {
			callScope.nodes[params[0].Name] = []TreeIndex{treeIndex}
		} else if len(params) >= 2 {
			callScope.nodes[params[0].Name] = []TreeIndex{treeIndex}
			callScope.nodes[params[1].Name] = []TreeIndex{NewConst(float64(varArgIndex))}
		}

		// Process the assignments and returns in the new scope
		err := callScope.EvaluateAssignments(mapAssignments)
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("\"map\" operation param %d assignments: %v", varArgIndex+numFixedParam, err)
		}
		treeIndices, f, err := callScope.EvaluateReturns(mapReturns)
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("\"map\" operation param %d returns: %v", varArgIndex+numFixedParam, err)
		}
		if f != nil {
			return noTreeIndices, nil, fmt.Errorf("\"map\" operation param %d return expected value, but got function", varArgIndex+numFixedParam)
		}
		result = append(result, treeIndices...)
	}

	return result, nil, nil
}

// reduce(initialValue, func(currentValue, value), varargs)
// Apply the function to each value, using the result of the prior function
// as input to the next function call. Return the final function call value.
func (s *Scope) fromReduce(expr *program.Expression) ([]TreeIndex, *ScopedFunc, error) {
	operands := expr.Operands
	if len(operands) < 3 {
		return noTreeIndices, nil, fmt.Errorf("\"reduce\" operation expects at least 2 arguments, but got %d", len(operands))
	}
	result, err := s.OneValueFromOperand(operands[0])
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("\"reduce\" operation initial value: %v", err)
	}
	reduceF, err := s.namedOrAnonFunction(operands[1], 2, 1)
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("\"reduce\" operation function: %v", err)
	}
	varargs, err := s.collectVarargs(operands, 2)
	if err != nil {
		return noTreeIndices, nil, fmt.Errorf("\"reduce\" operation varargs: %v", err)
	}

	reduceScope := reduceF.scope
	reducePbF := reduceF.pbFunc
	params := reducePbF.Parameters
	reduceAssignments := reducePbF.Assignments
	reduceReturns := reducePbF.Returns

	currentParamName := params[0].Name
	valueParamName := params[1].Name

	const numFixedParam = 2
	for varArgIndex, treeIndex := range varargs {
		callScope := NewScope([]*Scope{reduceScope})
		callScope.nodes[currentParamName] = []TreeIndex{result}
		callScope.nodes[valueParamName] = []TreeIndex{treeIndex}

		// Process the assignments and returns in the new scope
		err := callScope.EvaluateAssignments(reduceAssignments)
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("\"reduce\" operation param %d assignments: %v", varArgIndex+numFixedParam, err)
		}
		treeIndices, f, err := callScope.EvaluateReturns(reduceReturns)
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("\"reduce\" operation param %d returns: %v", varArgIndex+numFixedParam, err)
		}
		if f != nil {
			return noTreeIndices, nil, fmt.Errorf("\"reduce\" operation param %d return expected value, but got function", varArgIndex+numFixedParam)
		}
		if len(treeIndices) != 1 {
			return noTreeIndices, nil, fmt.Errorf("\"reduce\" operation param %d expected 1 return value, but got %d", varArgIndex+numFixedParam, len(treeIndices))
		}
		result = treeIndices[0]
	}
	return []TreeIndex{result}, nil, nil
}

func (s *Scope) fromRemap(expr *program.Expression) ([]TreeIndex, *ScopedFunc, error) {
	operands := expr.Operands

	switch len(operands) {
	case 2:
		// 2-arg form: remap(fn_call, vec3_point)
		// Evaluate the function call to get the target shape(s),
		// then extract x/y/z from the vec3 point for remapping.
		targets, err := s.MultiValueFromOperand(operands[0])
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("\"remap\" operation target: %v", err)
		}
		point, err := s.MultiValueFromOperand(operands[1])
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("\"remap\" operation point: %v", err)
		}
		if len(point) != 3 {
			return noTreeIndices, nil, fmt.Errorf("\"remap\" operation point expected 3 components, but got %d", len(point))
		}
		results := make([]TreeIndex, len(targets))
		for i, target := range targets {
			results[i] = Remap(target, point[0], point[1], point[2])
		}
		return results, nil, nil

	case 4:
		// 4-arg form: remap(target, x, y, z)
		targets, err := s.MultiValueFromOperand(operands[0])
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("\"remap\" operation target parameter: %v", err)
		}
		x, err := s.OneValueFromOperand(operands[1])
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("\"remap\" operation x parameter: %v", err)
		}
		y, err := s.OneValueFromOperand(operands[2])
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("\"remap\" operation y parameter: %v", err)
		}
		z, err := s.OneValueFromOperand(operands[3])
		if err != nil {
			return noTreeIndices, nil, fmt.Errorf("\"remap\" operation z parameter: %v", err)
		}
		results := make([]TreeIndex, len(targets))
		for i, target := range targets {
			results[i] = Remap(target, x, y, z)
		}
		return results, nil, nil

	default:
		return noTreeIndices, nil, fmt.Errorf("\"remap\" operation expects 2 or 4 arguments, but got %d", len(operands))
	}
}

// Return the referenced function or an anonymous function
func (s *Scope) namedOrAnonFunction(operand *program.Operand, expectedNumParam int, expectedNumReturn int) (*ScopedFunc, error) {
	_, f, err := s.MultiValueOrFunctionFromOperand(operand)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, fmt.Errorf("expected function, but got value")
	}
	pbF := f.pbFunc
	numParam := len(pbF.Parameters)
	if expectedNumParam >= 0 && numParam != expectedNumParam {
		return nil, fmt.Errorf("expected %d parameters, but got %d", expectedNumParam, numParam)
	}
	if len(pbF.Returns) != expectedNumReturn {
		return nil, fmt.Errorf("expected %d return values, but got %d", expectedNumReturn, len(pbF.Returns))
	}
	return f, nil
}
