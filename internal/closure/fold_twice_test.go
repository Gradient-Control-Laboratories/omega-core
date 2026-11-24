package closure_test

import (
	"testing"

	internalsource "github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// Two folds in one function over the same list, asking its elements different
// questions. Each unrolled step is a specialization of its fold's callback, and
// every one of them needs a name of its own. nextClosureName used to build
// names from a per-base counter alone, so the second step of the first fold
// (gsl_anon, counter 2: gcl_closure_gsl_anon_2) took the same name as the
// first step of the second (gsl_anon_2: gcl_closure_gsl_anon_2); the later
// registration replaced the earlier, and the first fold called the second's
// closures. Shaft Modeling's shaft.body met it asking its segments for their
// fields and then for their planes.
func TestClosure_TwoFoldsOverOneListKeepTheirNames(t *testing.T) {
	lib := `
dia(d, from, to) {
  seg(want) {
    field = length(p.xy) - d / 2.0;
    return want > 1.5 ? to : (want > 0.5 ? from : field);
  }
  return seg;
}
fieldThenPlane(segments) {
  f = reduce((a, seg) { return min(a, seg(0.0)); }, segments, 100000.0);
  lo = reduce((a, seg) { return min(a, seg(1.0)); }, segments, 100000.0);
  return f + 0.0 * lo;
}
`
	m := &internalsource.MockManager{PkgSources: map[string][]string{"geom/ask": {lib}}}
	src := "import ask \"geom/ask\";\nreturn ask.fieldThenPlane([ask.dia(20.0, 5.0, 25.0), ask.dia(12.0, 25.0, 45.0), ask.dia(8.0, 45.0, 50.0)]);\n"
	prog, err := runPipelineWithManager(src, m)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	fc := callOf(file.Returns[0])
	if fc == nil {
		t.Fatalf("return is not a call")
	}
	body := fnNamed(file, fc.Reference.Name)
	if body == nil || len(body.Assignments) < 2 {
		t.Fatalf("specialized fieldThenPlane not found, or without its two folds")
	}
	seen := map[string]int{}
	for i, assign := range body.Assignments[:2] {
		names := foldNames(assign.Expression)
		if len(names) != 3 {
			t.Errorf("fold %d unrolled to %v, want three steps", i, names)
		}
		for _, n := range names {
			seen[n]++
		}
	}
	for n, count := range seen {
		if count > 1 {
			t.Errorf("%q is the name of %d unrolled steps; each needs its own", n, count)
		}
	}
}
