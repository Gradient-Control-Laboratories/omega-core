package propval_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/propval"
	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/link"
	"gradientcontrol.com/omega/core/internal/source"
	parser "gradientcontrol.com/omega/core/parser"
)

func parseAndValidate(t *testing.T, src string) error {
	t.Helper()
	srcMgr := source.NewLocalManager("../../../stdlib")
	prog, err := collect.Program([]string{src}, srcMgr, parser.ParsePackage)
	if err != nil {
		return err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return err
	}
	return propval.Validate(prog)
}

func TestValidProperties(t *testing.T) {
	src := `
v = 1.0;
v.min = 0.0;
v.max = 2.0;
v.period = 5.0;

s = sphere(1.0);
s.box = [[-1.0, -1.0, -1.0], [1.0, 1.0, 1.0]];
s.color = [0.2, 0.4, 0.6];
s.material = [1.0, 0.0, 1.45, 1.0, 0.5, 0.0, 2.0, 0.0, 0.0, 0.0];

s2 = sphere(2.0);
s2.box = auto;

s3 = sphere(3.0);
s3.box = [[-5.0, -5.0, auto], [5.0, 5.0, auto + 1.0]];

return s;
`
	if err := parseAndValidate(t, src); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestCycleDetection(t *testing.T) {
	src := `
s1 = sphere(1.0);
s2 = sphere(2.0);

s1.box = s2.box;
s2.box = s1.box;

return s1;
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected cycle detection error, got nil")
	}
	if !strings.Contains(err.Error(), "cyclic dependency detected") {
		t.Fatalf("expected 'cyclic dependency detected', got: %v", err)
	}
}

func TestFunctionLocalSlider(t *testing.T) {
	src := `
f() {
	v = 1.0;
	v.min = 0.0;
	v.max = 2.0;
	return v;
}

return f();
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for function-local slider, got nil")
	}
	if !strings.Contains(err.Error(), "slider properties (.min, .max, .period) are only permitted on top-level file scope declarations") {
		t.Fatalf("expected top-level scope error, got: %v", err)
	}
}

func TestSliderMissingBounds(t *testing.T) {
	src := `
v = 1.0;
v.min = 0.0;

return v;
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for missing .max bound, got nil")
	}
	if !strings.Contains(err.Error(), "requires both .min and .max bounds") {
		t.Fatalf("expected missing bounds error, got: %v", err)
	}
}

func TestSliderInvertedRange(t *testing.T) {
	src := `
v = 1.0;
v.min = 5.0;
v.max = 2.0;

return v;
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for inverted min/max, got nil")
	}
	if !strings.Contains(err.Error(), "invalid slider range: min (5) must be less than max (2)") {
		t.Fatalf("expected invalid slider range error, got: %v", err)
	}
}

func TestPeriodNonPositive(t *testing.T) {
	src := `
v = 1.0;
v.min = 0.0;
v.max = 5.0;
v.period = -1.0;

return v;
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for non-positive period, got nil")
	}
	if !strings.Contains(err.Error(), "property .period must be strictly positive") {
		t.Fatalf("expected strictly positive period error, got: %v", err)
	}
}

func TestBoundingBoxInvertedExtents(t *testing.T) {
	src := `
s = sphere(1.0);
s.box = [[2.0, -1.0, -1.0], [1.0, 1.0, 1.0]];

return s;
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for inverted box extent, got nil")
	}
	if !strings.Contains(err.Error(), "invalid bounding box: lo.x (2) must be strictly less than hi.x (1)") {
		t.Fatalf("expected invalid bounding box error, got: %v", err)
	}
}

func TestBoundingBoxDependsOnP(t *testing.T) {
	src := `
s = sphere(1.0);
s.box = [[p.x, -1.0, -1.0], [1.0, 1.0, 1.0]];

return s;
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for box depending on p, got nil")
	}
	if !strings.Contains(err.Error(), "cannot depend on spatial coordinates p") {
		t.Fatalf("expected depends on p error, got: %v", err)
	}
}

func TestColorOutOfBounds(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{name: "above one", expr: "[1.2, 0.5, 0.5]"},
		{name: "negative", expr: "[-0.1, 0.5, 0.5]"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src := "s = sphere(1.0);\ns.color = " + tc.expr + ";\nreturn s;\n"
			err := parseAndValidate(t, src)
			if err == nil {
				t.Fatalf("expected error for color out of bounds, got nil")
			}
			if !strings.Contains(err.Error(), "color components must be in range [0.0, 1.0]") {
				t.Fatalf("expected color range error, got: %v", err)
			}
		})
	}
}

func TestColorValidRGBForms(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{
			name: "boundary rgb literal",
			src: `
s = sphere(1.0);
s.color = [0.0, 0.5, 1.0];
return s;
`,
		},
		{
			name: "imported rgb constant",
			src: `
import "color";
s = sphere(1.0);
s.color = color.red;
return s;
`,
		},
		{
			name: "varying rgb field",
			src: `
s = sphere(1.0);
s.color = [clamp(p.x, 0.0, 1.0), 0.4, 0.6];
return s;
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := parseAndValidate(t, tc.src); err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestColorVectorLengthValidation(t *testing.T) {
	tests := []struct {
		name        string
		src         string
		wantSubstrs []string
	}{
		{
			name: "2-element literal rejected",
			src: `
s = sphere(1.0);
s.color = [0.2, 0.4];
return s;
`,
			wantSubstrs: []string{
				`property .color on variable "s" must be a 3-element RGB vector [r, g, b], got 2 components`,
			},
		},
		{
			name: "2-element varying literal rejected",
			src: `
s = sphere(1.0);
s.color = [clamp(p.x, 0.0, 1.0), 0.4];
return s;
`,
			wantSubstrs: []string{
				`property .color on variable "s" must be a 3-element RGB vector [r, g, b], got 2 components`,
			},
		},
		{
			name: "2-element constant variable rejected",
			src: `
rg = [0.2, 0.4];
s = sphere(1.0);
s.color = rg;
return s;
`,
			wantSubstrs: []string{
				`property .color on variable "s" must be a 3-element RGB vector [r, g, b], got 2 components`,
			},
		},
		{
			name: "4-element rgba literal rejected with material hint",
			src: `
s = sphere(1.0);
s.color = [0.2, 0.4, 0.6, 1.0];
return s;
`,
			wantSubstrs: []string{
				`property .color on variable "s" must be a 3-element RGB vector [r, g, b], got 4 components`,
				`alpha is not supported on .color`,
				`.material`,
			},
		},
		{
			name: "4-element varying rgba literal rejected with material hint",
			src: `
s = sphere(1.0);
s.color = [clamp(p.x, 0.0, 1.0), 0.4, 0.6, 0.8];
return s;
`,
			wantSubstrs: []string{
				`property .color on variable "s" must be a 3-element RGB vector [r, g, b], got 4 components`,
				`alpha is not supported on .color`,
				`.material`,
			},
		},
		{
			name: "4-element constant variable rejected with material hint",
			src: `
rgba = [0.2, 0.4, 0.6, 0.8];
s = sphere(1.0);
s.color = rgba;
return s;
`,
			wantSubstrs: []string{
				`property .color on variable "s" must be a 3-element RGB vector [r, g, b], got 4 components`,
				`alpha is not supported on .color`,
				`.material`,
			},
		},
		{
			name: "5-element literal rejected",
			src: `
s = sphere(1.0);
s.color = [0.1, 0.2, 0.3, 0.4, 0.5];
return s;
`,
			wantSubstrs: []string{
				`property .color on variable "s" must be a 3-element RGB vector [r, g, b], got 5 components`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := parseAndValidate(t, tc.src)
			if err == nil {
				t.Fatalf("expected validation error, got nil")
			}
			for _, sub := range tc.wantSubstrs {
				if !strings.Contains(err.Error(), sub) {
					t.Fatalf("expected error containing %q, got: %v", sub, err)
				}
			}
		})
	}
}

func TestAutoKeywordOutsideBox(t *testing.T) {
	src := `
y = auto + 1.0;

return y;
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for auto outside .box, got nil")
	}
	if !strings.Contains(err.Error(), "reserved keyword \"auto\" is only permitted within .box property assignments") {
		t.Fatalf("expected reserved keyword auto error, got: %v", err)
	}
}

func TestAutoAsIdentifier(t *testing.T) {
	src := `
auto = 1.0;

return auto;
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for auto as identifier, got nil")
	}
	if !strings.Contains(err.Error(), "\"auto\" is reserved and only permitted within .box property assignments") {
		t.Fatalf("expected reserved keyword identifier error, got: %v", err)
	}
}

func TestReadingMetadataPropertyInExpressions(t *testing.T) {
	src := `
v = 1.0;
v.min = 0.0;
v.max = 2.0;
m = v.min;

return m;
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for reading property in expression, got nil")
	}
	if !strings.Contains(err.Error(), "reading property \"min\" in expressions is not supported; properties are declaration metadata") {
		t.Fatalf("expected metadata property read error, got: %v", err)
	}
}

func TestReadingUndefinedProperty(t *testing.T) {
	src := `
v = 1.0;
m = v.min;

return m;
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for variable has no property, got nil")
	}
	if !strings.Contains(err.Error(), "variable \"v\" has no property \"min\"") {
		t.Fatalf("expected has no property error, got: %v", err)
	}
}

func TestMemberAccessNoCycle(t *testing.T) {
	src := `
x = 10.0;
y = p.x + 1.0;
x.min = 0.0;
x.max = 20.0;

return y;
`
	err := parseAndValidate(t, src)
	if err != nil {
		t.Fatalf("unexpected cycle or validation error: %v", err)
	}
}

func TestQualifiedFunctionCallNoCycle(t *testing.T) {
	src := `
f(a) {
	return a;
}

stain = 1.0;
stain.min = 0.0;
stain.max = 5.0;

v = f(stain);
return v;
`
	err := parseAndValidate(t, src)
	if err != nil {
		t.Fatalf("unexpected cycle or validation error: %v", err)
	}
}

func TestSliderDefaultValueOutOfBounds(t *testing.T) {
	src := `
v = 10.0;
v.min = 0.0;
v.max = 5.0;
return v;
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for default value out of bounds, got nil")
	}
	if !strings.Contains(err.Error(), "slider default value 10 is outside range [0, 5]") {
		t.Fatalf("expected default value out of range error, got: %v", err)
	}
}

func TestExplicitRangeAnnotation_Valid(t *testing.T) {
	src := `
@range(0.5, 10.0, 4.0)
radius = 5.0;

@range(1.0, 20.0)
height = 10.0;

return radius + height;
`
	if err := parseAndValidate(t, src); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestExplicitRangeAnnotation_Inverted(t *testing.T) {
	src := `
@range(10.0, 5.0)
radius = 6.0;
return radius;
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for inverted @range, got nil")
	}
	if !strings.Contains(err.Error(), "range min (10) must be less than max (5)") {
		t.Fatalf("expected range min less than max error, got: %v", err)
	}
}

func TestExplicitRangeAnnotation_NonPositivePeriod(t *testing.T) {
	src := `
@range(1.0, 10.0, -2.0)
radius = 5.0;
return radius;
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for non-positive period in @range, got nil")
	}
	if !strings.Contains(err.Error(), "range period must be strictly positive") {
		t.Fatalf("expected positive period error, got: %v", err)
	}
}

func TestExplicitRangeAnnotation_DefaultValueOutOfBounds(t *testing.T) {
	src := `
@range(1.0, 10.0)
radius = 20.0;
return radius;
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for default value out of bounds, got nil")
	}
	if !strings.Contains(err.Error(), "slider default value 20 is outside range [1, 10]") {
		t.Fatalf("expected default value out of range error, got: %v", err)
	}
}

func TestExplicitRangeAnnotation_FunctionLocal(t *testing.T) {
	src := `
f() {
	@range(1.0, 10.0)
	v = 5.0;
	return v;
}
return f();
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for function-local @range, got nil")
	}
	if !strings.Contains(err.Error(), "slider @range annotations on variables are only permitted on top-level file scope declarations") {
		t.Fatalf("expected top-level scope error, got: %v", err)
	}
}

func TestExplicitRangeAnnotation_NonScalar(t *testing.T) {
	src := `
@range(1.0, 10.0)
s = sphere(5.0);
return s;
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for @range on non-scalar, got nil")
	}
	if !strings.Contains(err.Error(), "range annotation can only be attached to scalar numeric variables") {
		t.Fatalf("expected scalar error, got: %v", err)
	}
}

func TestParameterRangeAnnotation_Valid(t *testing.T) {
	src := `
scaleRadius(@range(0.5, 10.0) r) {
	return r * 2.0;
}
return scaleRadius(5.0);
`
	if err := parseAndValidate(t, src); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestParameterRangeAnnotation_Inverted(t *testing.T) {
	src := `
scaleRadius(@range(10.0, 1.0) r) {
	return r * 2.0;
}
return scaleRadius(5.0);
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for inverted parameter @range, got nil")
	}
	if !strings.Contains(err.Error(), "parameter \"r\" range min (10) must be less than max (1)") {
		t.Fatalf("expected inverted parameter range error, got: %v", err)
	}
}

func TestParameterRangeAnnotation_PeriodRejected(t *testing.T) {
	src := `
scaleRadius(@range(1.0, 10.0, 2.0) r) {
	return r * 2.0;
}
return scaleRadius(5.0);
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected error for period argument in parameter @range, got nil")
	}
	if !strings.Contains(err.Error(), "@range annotation on a parameter requires exactly 2 arguments: @range(min, max)") {
		t.Fatalf("expected period rejection error, got: %v", err)
	}
}
