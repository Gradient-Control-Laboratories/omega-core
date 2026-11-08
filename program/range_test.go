package program_test

import (
	"testing"

	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func TestRangeAnnotation_IsSynthetic(t *testing.T) {
	loc1 := program.NewLoc(2, 5, 10)
	loc2 := program.NewLoc(2, 5, 10)
	loc3 := program.NewLoc(2, 10, 15)

	synthetic := &program.Range{
		Start: loc1,
		End:   loc2,
	}
	if !synthetic.IsSynthetic() {
		t.Errorf("expected synthetic.IsSynthetic() to be true")
	}

	explicit := &program.Range{
		Start: loc1,
		End:   loc3,
	}
	if explicit.IsSynthetic() {
		t.Errorf("expected explicit.IsSynthetic() to be false")
	}
}

func TestRangeAnnotation_ConstAccessors(t *testing.T) {
	minOp := program.NewValue(1.5, nil, nil)
	maxOp := &program.Operand{ConstValues: []float64{10.0}}
	periodOp := program.NewValue(2.0, nil, nil)

	r := &program.Range{
		Min:    minOp,
		Max:    maxOp,
		Period: periodOp,
	}

	minVal, ok := r.MinConst()
	if !ok || minVal != 1.5 {
		t.Errorf("expected MinConst() = (1.5, true), got (%v, %v)", minVal, ok)
	}

	maxVal, ok := r.MaxConst()
	if !ok || maxVal != 10.0 {
		t.Errorf("expected MaxConst() = (10.0, true), got (%v, %v)", maxVal, ok)
	}

	periodVal, ok := r.PeriodConst()
	if !ok || periodVal != 2.0 {
		t.Errorf("expected PeriodConst() = (2.0, true), got (%v, %v)", periodVal, ok)
	}

	rNoPeriod := &program.Range{
		Min: minOp,
		Max: maxOp,
	}
	if _, ok := rNoPeriod.PeriodConst(); ok {
		t.Errorf("expected rNoPeriod.PeriodConst() to return false")
	}
}

func TestComputeSliderVars(t *testing.T) {
	src := `@range(1.0, 10.0)
radius = 5.0;
scale = radius * 2.0;
offset = scale + 1.0;
staticVal = 42.0;
return offset;
`
	res, err := parser.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	sliderVars := program.ComputeSliderVars(res.Package)
	if !sliderVars["radius"] {
		t.Errorf("expected radius to be in sliderVars")
	}
	if !sliderVars["scale"] {
		t.Errorf("expected scale to be in sliderVars")
	}
	if !sliderVars["offset"] {
		t.Errorf("expected offset to be in sliderVars")
	}
	if sliderVars["staticVal"] {
		t.Errorf("expected staticVal to NOT be in sliderVars")
	}
}
