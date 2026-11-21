package closure_test

import (
	"strings"
	"testing"
)

// TestClosure_CalleeKeepsOwnParameter checks that specializing a higher-order
// function binds its function parameter in its own body only. apply's f is
// bound to voice; osc, reached through voice, has a parameter f of its own,
// which must stay a parameter. The pass used to transform reached callees in
// the caller's environment, which rewrote osc's f to voice.
func TestClosure_CalleeKeepsOwnParameter(t *testing.T) {
	src := `
osc(f, t) {
  return sin(f * t);
}
voice(h) {
  return osc(220, h);
}
apply(f, x) {
  return f(x);
}
return apply(voice, time);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	file := prog.Primary.Files[0]
	for _, ft := range file.FnTypes {
		if ft.Name != "osc" {
			continue
		}
		body := ft.PrettyPrint(0)
		if strings.Contains(body, "voice") {
			t.Fatalf("osc's parameter f was rewritten by apply's binding:\n%s", body)
		}
		return
	}
	t.Fatal("osc not found after the transform")
}
