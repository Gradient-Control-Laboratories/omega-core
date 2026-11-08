package convert_test

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/fidget/internal/convert"
	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// TestBuildDoesNotMutateModel guards against build passes mutating the shared
// parsed model. The variadic const-fold in common/build visit_operation.go
// once removed operands in place from expr.Operands, so the first build of
// "2.0 * f * t" rewrote the model to "2.0 * t * t" and every later build of
// the same function produced corrupted trees. Fidget rebuilds functions on
// every call site, so the second call's geometry silently diverged from the
// wgsl render. The function body below needs a const-foldable operand (f)
// followed by a non-const operand (t) inside a variadic op, and must be
// called more than once.
func TestBuildDoesNotMutateModel(t *testing.T) {
	src := `bez(a, b, c) {
  f = a - 2.0 * b + c;
  g = 2.0 * (b - a);
  t = clamp(p.x * p.y, 0.0, 1.0);
  tangent = g + 2.0 * f * t;
  return length(p - tangent) - 1.0;
}

first = bez([0, 0, -1.0], [0, 0, -2.0], [-2.0, 0, -3.0]);
second = bez([0, 0, -1.0], [0, 0, -2.0], [-2.0, 0, -3.0]);

return min(first, second);
`
	res, err := assembler.Assemble([]string{src}, &source.MockManager{})
	require.NoError(t, err)

	before, err := proto.Marshal(res.Program.Primary)
	require.NoError(t, err)

	err = convert.ToTrees(res.Program, func(treeIndices []fidget.TreeIndex, _ []string) error {
		return nil
	})
	require.NoError(t, err)

	after, err := proto.Marshal(res.Program.Primary)
	require.NoError(t, err)

	if string(before) != string(after) {
		t.Fatal("building the module mutated the parsed model")
	}
}

// TestReduceLambdaDoesNotLeakBindings guards against inline parameter
// bindings leaking into nested function bodies. A reduce lambda's "value"
// binding once shadowed the parameter of a function called from the lambda
// body, so g(1.0) below evaluated as g(p.x) / g(p.y) and the reduce produced
// a position-dependent field instead of the constant 200.
func TestReduceLambdaDoesNotLeakBindings(t *testing.T) {
	src := `g(value) {
  tmp = value * 100.0;
  return tmp;
}

r = reduce(0.0, (acc, value) {
  return acc + g(1.0);
}, [p.x, p.y]);

return r;
`
	res, err := assembler.Assemble([]string{src}, &source.MockManager{})
	require.NoError(t, err)

	var glsl string

	err = convert.ToTrees(res.Program, func(treeIndices []fidget.TreeIndex, _ []string) error {
		canvasIndex := fidget.BuildCanvas(treeIndices[0])
		glsl = fidget.CodegenGLSL(canvasIndex)
		return nil
	})
	require.NoError(t, err)

	if !strings.Contains(glsl, "200.0") || strings.Contains(glsl, "iP.x") {
		t.Fatalf("reduce result should fold to the constant 200, got:\n%s", glsl)
	}
}
