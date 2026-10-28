package codegen_test

import (
	"strings"
	"testing"
)

// A collection wider than a vec4 reaching a real function, and arithmetic on
// one, are two separate holes in the WGSL backend. Both need the same setup,
// because both only appear once the value is a runtime one: stud() is
// p-dependent and called twice, so the inliner keeps it a function (a
// code-size heuristic), its argument stays a runtime value, and everything
// below it is compiled rather than folded.

// TestWideTableAsFunctionParameter pins the WGSL type of a table whose rows are
// wider than a vec4 when the table reaches a function as a runtime parameter.
//
// NewCollect used to decide "is this a list of collections?" by testing its
// elements for *vecNumType alone — the type of a row of 2, 3 or 4. A row of 5
// or more is an *arrayNumType, which failed that test, so the whole table
// collapsed to array<f32, rowCount>: the element type was lost and the row
// count stood in for the row width. Reading a row back out of it then typed as
// num, and indexing that row failed with "target must be a collection, got
// num".
func TestWideTableAsFunctionParameter(t *testing.T) {
	src := []byte(`tbl = [[1.0, 2.0, 3.0, 4.0, 5.0, 6.0],
       [2.0, 3.0, 4.0, 5.0, 6.0, 7.0]];

pick(rows, k) {
  key = k;
  return reduce((acc, v) { return acc + v; },
                map((r) { return r[3] * max(0.0, 1.0 - abs(key - r[0]) / 0.2); },
                    rows));
}

stud(d) {
  v = pick(tbl, d);
  body = max(length(p.xy) - v, abs(p.z) - 1.0);
  return body;
}

return min(stud(1.0), stud(2.0));`)

	res, err := runTestBuildOne(t, "", src)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	const want = "fn pick(rows: array<array<f32, 6>, 2>, k: f32) -> f32 {"
	if !strings.Contains(res.WGSL, want) {
		t.Errorf("a table of 6-wide rows must reach a parameter as a list of rows.\nwant a declaration of:\n  %s\ngot:\n%s", want, res.WGSL)
	}
	// The row index comes from the unrolled fold and the column is written at
	// the index, so both are literals and the read is a plain double subscript.
	if !strings.Contains(res.WGSL, "rows[0][3]") {
		t.Errorf("want the row read as rows[0][3]; got:\n%s", res.WGSL)
	}
}

// TestWideCollectionArithmetic pins element-wise arithmetic on a collection
// wider than a vec4.
//
// GSL arithmetic on a collection is element-wise at any width — the fidget
// evaluator gives `[1, 2, 3, 4, 5, 6] * 2` as `[2, 4, 6, 8, 10, 12]` — but the
// WGSL backend pasted the operator between its operands whatever their type.
// vec2/vec3/vec4 have operators, so those were right; anything wider came out
// as `array<f32, 6>(...) * k`, which WGSL has no arithmetic for, so the
// program meshed correctly and then failed to compile as a shader.
func TestWideCollectionArithmetic(t *testing.T) {
	src := []byte(`wide = [1.0, 2.0, 3.0, 4.0, 5.0, 6.0];
narrow = [1.0, 2.0, 3.0];

scaled(k) {
  w = wide * k;
  c = max(wide, k);
  nrw = narrow * k;
  return w[1] + c[4] + nrw[1];
}

stud(d) {
  v = scaled(d);
  body = max(length(p.xy) - v, abs(p.z) - 1.0);
  return body;
}

return min(stud(1.0), stud(2.0));`)

	res, err := runTestBuildOne(t, "", src)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for _, want := range []string{
		// Wider than a vec4: unrolled, one scalar operation per element.
		"let w = array<f32, 6>(1.0 * k, 2.0 * k, 3.0 * k, 4.0 * k, 5.0 * k, 6.0 * k);",
		"let c = array<f32, 6>(max(1.0, k), max(2.0, k), max(3.0, k), max(4.0, k), max(5.0, k), max(6.0, k));",
		// A vec3 has a real WGSL operator, so it must NOT be unrolled.
		"let nrw = vec3<f32>(1.0, 2.0, 3.0) * k;",
	} {
		if !strings.Contains(res.WGSL, want) {
			t.Errorf("want:\n  %s\ngot:\n%s", want, res.WGSL)
		}
	}
}

// TestWideRowListArithmetic covers the recursion: a list of wide rows unrolls
// to rows and each row then unrolls to scalars. Before the fix this emitted
// `array<f32, 6>(...) + array<f32, 6>(...)`, an addition of two WGSL arrays.
func TestWideRowListArithmetic(t *testing.T) {
	src := []byte(`tbl = [[1.0, 2.0, 3.0, 4.0, 5.0, 6.0],
       [7.0, 8.0, 9.0, 10.0, 11.0, 12.0]];

doubled(k) {
  q = tbl + tbl;
  r = q[0];
  return r[1] * k;
}

stud(d) {
  v = doubled(d);
  body = max(length(p.xy) - v, abs(p.z) - 1.0);
  return body;
}

return min(stud(1.0), stud(2.0));`)

	res, err := runTestBuildOne(t, "", src)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if strings.Contains(res.WGSL, "6>(1.0, 2.0, 3.0, 4.0, 5.0, 6.0) + array") {
		t.Errorf("two rows must not be added with a WGSL array operator; got:\n%s", res.WGSL)
	}
}

// TestSingleElementListCodegen tests that arithmetic and operations on single-element
// lists emit valid WGSL without attempting illegal array operator syntax.
func TestSingleElementListCodegen(t *testing.T) {
	src := []byte(`
calc(k) {
  a = [k];
  b = [2.0];
  sum = a + b;
  scale = sum * 3.0;
  return scale[0];
}

stud(d) {
  v = calc(d);
  return max(length(p.xy) - v, abs(p.z) - 1.0);
}

return min(stud(1.0), stud(2.0));
`)

	res, err := runTestBuildOne(t, "", src)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for _, line := range strings.Split(res.WGSL, "\n") {
		if strings.Contains(line, "+") && strings.Contains(line, "array<f32, 1>") {
			t.Errorf("illegal array addition in WGSL: %s", line)
		}
	}
}
