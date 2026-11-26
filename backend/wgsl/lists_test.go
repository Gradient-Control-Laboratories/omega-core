package wgsl_test

import (
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/wgsl"
)

// foldedRadius assembles a program returning sphere(r) for a constant r and
// reports whether the generated shader folded r to want.
func foldedRadius(t *testing.T, src string, want string) {
	t.Helper()
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("%v\n%s", err, src)
	}
	out, err := wgsl.Render(src, res.Program)
	if err != nil {
		t.Fatalf("%v\n%s", err, src)
	}
	if !strings.Contains(out.WGSL, "length(p) - "+want+",") {
		t.Errorf("want the radius folded to %s:\n%s", want, src)
	}
}

// A vector added to or subtracted from a list of vectors meets each element.
func TestVectorBroadcastsOverAListOfVectors(t *testing.T) {
	// b[1] = [14, 25, 36], b[0] = [11, 22, 33]: 36 + 11
	foldedRadius(t, "a = [[1, 2, 3], [4, 5, 6]];\nb = a + [10, 20, 30];\nreturn sphere(b[1].z + b[0].x);", "47.0")
	// v - a subtracts each element from v: b[1].z = 30 - 6, b[0].x = 10 - 1
	foldedRadius(t, "a = [[1, 2, 3], [4, 5, 6]];\nb = [10, 20, 30] - a;\nreturn sphere(b[1].z * 100 + b[0].x);", "2409.0")
	foldedRadius(t, "a = [[1, 2, 3], [4, 5, 6]];\nb = a - [1, 1, 1];\nreturn sphere(b[1].y);", "4.0")
}

func TestVectorBroadcastNeedsMatchingLengths(t *testing.T) {
	msg := listAssembleErr(t, "a = [[1, 2], [4, 5], [7, 8]];\nb = a + [1, 1, 1];\nreturn sphere(b[0].x);")
	if !strings.Contains(msg, "cannot add a vector of length 3 and a list of vectors of length 2") {
		t.Errorf("unexpected error:\n%s", msg)
	}
}

// min and max of one list fold it: per component for a list of vectors.
func TestMinMaxOfOneList(t *testing.T) {
	// lo = [0, 2, 1], hi.x = 4
	foldedRadius(t, "pts = [[1, 5, 3], [4, 2, 6], [0, 9, 1]];\nlo = min(pts);\nhi = max(pts);\nreturn sphere(lo.x + lo.y * 10 + lo.z * 100 + hi.x * 1000);", "4120.0")
	foldedRadius(t, "v = [3, 1, 2, 7, 5];\nreturn sphere(min(v) + max(v) * 10);", "71.0")
	foldedRadius(t, "return sphere(min(4));", "4.0")
}

// A document's own unionSharp does not stand in for min's fold.
func TestMinOfOneListIgnoresALocalUnionSharp(t *testing.T) {
	foldedRadius(t, "unionSharp(fields) { return 42; }\nreturn sphere(min([3, 1, 2]));", "1.0")
}

// listAssembleErr assembles src and returns its error text, failing the test when
// it assembles.
func listAssembleErr(t *testing.T, src string) string {
	t.Helper()
	_, err := assembler.Assemble([]string{src})
	if err == nil {
		t.Fatalf("expected an error for:\n%s", src)
	}
	return err.Error()
}
