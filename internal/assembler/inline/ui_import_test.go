package inline_test

import (
	"reflect"
	"testing"

	"gradientcontrol.com/omega/core/assembler/ui"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/source"
)

func TestImportedUIIdentityKeepsCallerControl(t *testing.T) {
	src := `import "dep";
frame = uiTransform([[1,0,0,0],[0,1,0,0],[0,0,1,0],[7,8,9,1]]);
return dep.identity(frame);`
	dep := `identity(frame) {
  return uiTransform(frame);
}`
	prog, err := runPipelineWithManager(src, &source.MockManager{PkgSources: map[string][]string{"dep": {dep}}})
	if err != nil {
		t.Fatal(err)
	}
	controls := ui.Extract(prog.Primary, prog.GetImport)
	if len(controls) != 1 || len(controls[0].Params) != 16 {
		t.Fatalf("the caller's editable transform must survive import inlining: %+v", controls)
	}
	want := []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 7, 8, 9, 1}
	got, err := eval.ConstPkg(prog.Primary.Files[0].Returns[0], prog.Primary, prog.GetImport)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("identity changed values: got %v, want %v", got, want)
	}
}
