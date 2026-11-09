package parser2

import (
	"errors"

	"github.com/alecthomas/participle/v2"
	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/parser/internal/identifier2"
	"gradientcontrol.com/omega/core/program"
)

func buildAssignment(buildScope scope.BuilderScope, assignment *Assignment, rng *program.Range) error {
	if assignment == nil {
		return errors.New("missing assignment")
	}
	if len(assignment.Names.Names) == 0 {
		return errors.New("missing identifier")
	}

	// Build expressions before adding variable names
	// to ensure no forward references

	expr, err := buildExpression(buildScope, assignment.Expression)
	if err != nil {
		return err
	}

	names := make([]string, len(assignment.Names.Names))
	for i, nameVar := range assignment.Names.Names {
		name := nameVar.Name
		err := identifier2.Validate(name)
		if err != nil {
			return &participle.ParseError{Pos: nameVar.Pos, Msg: err.Error()}
		}
		names[i] = name
	}
	start := posToLocation(assignment.Names.Names[0].Pos)
	end := expr.End
	err = buildScope.AddAssignment(names, expr, start, end, rng)
	if err != nil {
		return &participle.ParseError{Pos: assignment.Pos, Msg: err.Error()}
	}

	return nil
}

func buildPropertyAssignment(buildScope scope.BuilderScope, propAssignment *PropertyAssignment) error {
	if propAssignment == nil {
		return errors.New("missing property assignment")
	}

	expr, err := buildExpression(buildScope, propAssignment.Expression)
	if err != nil {
		return err
	}

	varName := propAssignment.VarName
	propName := propAssignment.PropName
	start := posToLocation(propAssignment.Pos)
	end := posToLocation(propAssignment.EndPos)
	err = buildScope.AddPropertyAssignment(varName, propName, expr, start, end)
	if err != nil {
		return &participle.ParseError{
			Pos: propAssignment.Pos,
			Msg: err.Error(),
		}
	}

	return nil
}
