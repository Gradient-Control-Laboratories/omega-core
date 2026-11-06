package codegen_test

import (
	"strings"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/assembler/variation"
	"gradientcontrol.com/omega/core/internal/slider"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/ui"
)

func TestUIShaderParametersMatchViewer(t *testing.T) {
	source := `radius = 5;
radius.min = 1;
radius.max = 10;
center = uiPoint([1, 2, 3]);
xf = uiTransform([[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [4, 5, 6, 1]]);
path = uiSpline([xf, xf, xf]);
return length(vector.multiplyPoint(path[2], p) - center) - radius;`
	assembled, err := assembler.Assemble([]string{source}, mockMgr)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, control := range slider.Extract(source, assembled.Program.Primary) {
		names = append(names, control.ParamName)
	}
	for _, control := range ui.Extract(assembled.Program.Primary, assembled.Program.GetImport) {
		for _, param := range control.Params {
			names = append(names, param.ParamName)
		}
	}
	if len(names) != 20 {
		t.Fatalf("expected one slider, one point and one transform: %v", names)
	}
	rendered, err := renderHelper(source, assembled.Program)
	if err != nil {
		t.Fatal(err)
	}
	params := []string{"p: vec3<f32>", "time: f32"}
	for _, name := range names {
		params = append(params, name+": f32")
		if strings.Count(rendered.WGSL, name) < 2 {
			t.Fatalf("UI parameter %s is not used by the shader: %s", name, rendered.WGSL)
		}
	}
	want := "fn sdShape(" + strings.Join(params, ", ") + ") -> gsl_shape"
	if !strings.Contains(rendered.WGSL, want) {
		t.Fatalf("shader signature does not match viewer argument order; want %s, got %s", want, rendered.WGSL)
	}
}

func TestUIParametersDoNotCaptureImportedLiterals(t *testing.T) {
	src := "import \"dep\";\ncenter = uiPoint([1, 2, 3]);\nreturn length(p - center) + dep.test(p);"
	dep := "test(q) {\n  return q.x +    7;\n}"
	if strings.Index(strings.Split(src, "\n")[1], "1") != strings.Index(strings.Split(dep, "\n")[1], "7") {
		t.Fatal("fixture literals must occupy the same source span")
	}
	assembled, err := assembler.Assemble([]string{src}, &source.MockManager{PkgSources: map[string][]string{"dep": {dep}}})
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := renderHelper(src, assembled.Program)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered.WGSL, "+ 7.") {
		t.Fatalf("imported constant was replaced by a UI parameter: %s", rendered.WGSL)
	}
}

func TestUIParametersDoNotCaptureInlinedControls(t *testing.T) {
	src := "import \"dep\";\ncenter = uiPoint([1, 2, 3]);\nreturn length(p - center) + length(p - dep.point());"
	dep := "point() {\ncenter = uiPoint([7, 8, 9]);\nreturn center;\n}"
	assembled, err := assembler.Assemble([]string{src}, &source.MockManager{PkgSources: map[string][]string{"dep": {dep}}})
	if err != nil {
		t.Fatal(err)
	}
	controls := ui.Extract(assembled.Program.Primary, assembled.Program.GetImport)
	if len(controls) != 1 || len(controls[0].Params) != 3 {
		t.Fatalf("expected only the primary point: %+v", controls)
	}
	rendered, err := renderHelper(src, assembled.Program)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered.WGSL, "vec3<f32>(7.0, 8.0, 9.0)") {
		t.Fatalf("inlined branch constant was captured by the primary UI parameters: %s", rendered.WGSL)
	}
}

func TestZeroReturnShaderParametersMatchViewer(t *testing.T) {
	src := `importId = importStep "part.step";
radius = 5;
radius.min = 1;
radius.max = 10;
center = uiPoint([1, 2, 3]);
xf = uiTransform([[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [4, 5, 6, 1]]);
cavity = length(p - center) - radius;
return meshDifference(mesh(importId, xf), cavity);`
	assembled, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, control := range slider.Extract(src, assembled.Program.Primary) {
		names = append(names, control.ParamName)
	}
	for _, control := range ui.Extract(assembled.Program.Primary, assembled.Program.GetImport) {
		for _, param := range control.Params {
			names = append(names, param.ParamName)
		}
	}
	if len(names) != 20 {
		t.Fatalf("expected one slider, one point and one transform: %v", names)
	}
	sliced := variation.Program(assembled.Program)
	rendered, err := renderHelper(src, sliced)
	if err != nil {
		t.Fatal(err)
	}
	params := []string{"p: vec3<f32>", "time: f32"}
	for _, name := range names {
		params = append(params, name+": f32")
	}
	want := "fn sdShape(" + strings.Join(params, ", ") + ") -> gsl_shape"
	if !strings.Contains(rendered.WGSL, want) {
		t.Fatalf("zero-return shader signature does not match viewer argument order; want %s, got %s", want, rendered.WGSL)
	}
	if !strings.Contains(rendered.WGSL, "return gsl_shape(1e10, 0u);") {
		t.Fatalf("expected empty shape sentinel in zero-return shader: %s", rendered.WGSL)
	}
}
