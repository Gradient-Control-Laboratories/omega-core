package closure_test

import (
	"testing"
)

// A function that passes its function parameter on to a template only inside
// a fold's callback is a template too: once its call is specialized, the
// unspecialized declaration, still calling the template its callback names,
// has to go with that template, or the re-link fails.
func TestClosure_FoldCallbackPassesParamToTemplate(t *testing.T) {
	src := `@scene(80)
inner(q, blend) {
  return blend([p.x - q, p.y]);
}
f(blend) {
  return reduce([nothing, 0], (best, q, i) { return [min(best.x, inner(q, blend)), 0]; }, [1, 2]).x;
}
return max(f(intersectRound(2)), abs(p.z) - 1);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	for _, name := range []string{"f", "inner"} {
		if fnNamed(file, name) != nil {
			t.Errorf("%s survives the transform; as a specialized template it should be pruned", name)
		}
	}
}

// A callback that binds the name as its own parameter does not read the
// enclosing one, so it does not make the enclosing function a template.
func TestClosure_FoldCallbackShadowingParamIsNotATemplate(t *testing.T) {
	src := `@scene(80)
g(blend) {
  return reduce([nothing, 0], (best, blend) { return [min(best.x, length(p.xy) - blend), 0]; }, [1, 2]).x + blend;
}
return max(g(3), abs(p.z) - 1);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	if fnNamed(prog.Primary.Files[0], "g") == nil {
		t.Errorf("g was pruned, though its callback's blend is its own parameter")
	}
}
