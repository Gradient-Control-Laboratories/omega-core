package program_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

func TestPrettyPrint_AssignmentRange(t *testing.T) {
	// Explicit @range(1, 10, 2)
	assign := &program.Assignment{
		Variables:  []*program.Variable{{Name: "radius"}},
		Expression: program.NewValue(5.0, nil, nil),
		Range: &program.Range{
			Start:  program.NewLoc(1, 1, 0),
			End:    program.NewLoc(1, 15, 14),
			Min:    program.NewValue(1.0, nil, nil),
			Max:    program.NewValue(10.0, nil, nil),
			Period: program.NewValue(2.0, nil, nil),
		},
	}

	out := assign.PrettyPrint(0)
	if !strings.Contains(out, "@range(1, 10, 2)\nradius = 5") {
		t.Errorf("expected pretty print to include @range annotation, got:\n%s", out)
	}

	// Synthetic @range should NOT print @range
	syntheticAssign := &program.Assignment{
		Variables:  []*program.Variable{{Name: "radius"}},
		Expression: program.NewValue(5.0, nil, nil),
		Range: &program.Range{
			Start: program.NewLoc(1, 1, 0),
			End:   program.NewLoc(1, 1, 0),
			Min:   program.NewValue(1.0, nil, nil),
			Max:   program.NewValue(10.0, nil, nil),
		},
	}
	outSynthetic := syntheticAssign.PrettyPrint(0)
	if strings.Contains(outSynthetic, "@range") {
		t.Errorf("expected synthetic range to NOT print @range, got:\n%s", outSynthetic)
	}
}

func TestPrettyPrint_FunctionParameterRange(t *testing.T) {
	fn := &program.Function{
		Parameters: []*program.Parameter{
			{
				Name: "r",
				Range: &program.Range{
					Start: program.NewLoc(1, 1, 0),
					End:   program.NewLoc(1, 10, 9),
					Min:   program.NewValue(0.1, nil, nil),
					Max:   program.NewValue(5.0, nil, nil),
				},
			},
			{
				Name: "h",
			},
		},
		Returns: []*program.Operand{program.NewValue(1.0, nil, nil)},
	}

	out := fn.PrettyPrint(0)
	expectedParamPrefix := "(@range(0.1, 5) r, h)"
	if !strings.Contains(out, expectedParamPrefix) {
		t.Errorf("expected parameter to include @range prefix, got:\n%s", out)
	}
}
