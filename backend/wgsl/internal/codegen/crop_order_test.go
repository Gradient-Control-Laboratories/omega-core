package codegen_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/source"
)

// A document function's declared box is applied by cropping the shape where it
// is assigned, so a box reading a local assigned later read it before its
// assignment ("undefined identifier"). A cropped body is now put in an order its
// reads allow, which keeps every reader of the shape reading the cropped field;
// a body already in such an order is left as written.

// cropOrderWGSL compiles src to WGSL, without its comment lines.
func cropOrderWGSL(t *testing.T, src string) string {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
	if err != nil {
		t.Fatalf("assemble: %v\n%s", err, src)
	}
	out, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	var kept []string
	for _, line := range strings.Split(out.WGSL, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "//") {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

func TestCropOrder_BoxReadsALaterLocal(t *testing.T) {
	for _, c := range []struct{ name, written, ordered string }{
		{
			"a local after the shape",
			"f(points) {\n  form = length(p.xy) - 1;\n  lo = points[0];\n" +
				"  form.box = [[lo.x, lo.y, -1], [lo.x + 10, lo.y + 10, 1]];\n  return form;\n}\n" +
				"solid = f([[0, 0], [10, 0], [10, 10]]);\nreturn solid;\n",
			"f(points) {\n  lo = points[0];\n  form = length(p.xy) - 1;\n" +
				"  form.box = [[lo.x, lo.y, -1], [lo.x + 10, lo.y + 10, 1]];\n  return form;\n}\n" +
				"solid = f([[0, 0], [10, 0], [10, 10]]);\nreturn solid;\n",
		},
		{
			"the box read",
			"f(points) {\n  form = length(p.xy) - 1;\n  lo = points[0];\n" +
				"  form.box = [[lo.x, lo.y, -1], [lo.x + 10, lo.y + 10, 1]];\n  return form;\n}\n" +
				"solid = f([[0, 0], [10, 0], [10, 10]]);\nb = solid.box;\nreturn solid;\n",
			"f(points) {\n  lo = points[0];\n  form = length(p.xy) - 1;\n" +
				"  form.box = [[lo.x, lo.y, -1], [lo.x + 10, lo.y + 10, 1]];\n  return form;\n}\n" +
				"solid = f([[0, 0], [10, 0], [10, 10]]);\nb = solid.box;\nreturn solid;\n",
		},
		{
			// g reads the shape between its line and the box's local: it reads the
			// cropped shape, as it would with the local written first.
			"a reader between",
			"f(points) {\n  form = length(p.xy) - 1;\n  g = form - 0.5;\n  lo = points[0];\n" +
				"  form.box = [[lo.x, lo.y, -1], [lo.x + 10, lo.y + 10, 1]];\n  return g;\n}\n" +
				"solid = f([[0, 0], [10, 0], [10, 10]]);\nreturn solid;\n",
			"f(points) {\n  lo = points[0];\n  form = length(p.xy) - 1;\n  g = form - 0.5;\n" +
				"  form.box = [[lo.x, lo.y, -1], [lo.x + 10, lo.y + 10, 1]];\n  return g;\n}\n" +
				"solid = f([[0, 0], [10, 0], [10, 10]]);\nreturn solid;\n",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got, want := cropOrderWGSL(t, c.written), cropOrderWGSL(t, c.ordered); got != want {
				t.Errorf("written with the local after the shape, the WGSL differs from the local written first")
			}
		})
	}
}

// A box that reads the shape itself goes round in a cycle no order breaks: an
// error naming the shape and the local between.
func TestCropOrder_BoxReadingTheShapeIsAnError(t *testing.T) {
	for _, c := range []struct{ box, want string }{
		{"  lo = [form, 0];\n  form.box = [[lo.x, lo.y, -1], [lo.x + 10, lo.y + 10, 1]];\n",
			"form.box: the box reads lo, which reads form, the shape it bounds"},
		{"  form.box = [[form, 0, -1], [10, 10, 1]];\n", "form.box: the box reads form, the shape it bounds"},
	} {
		src := "f(points) {\n  form = length(p.xy) - 1;\n" + c.box + "  return form;\n}\n" +
			"solid = f([[0, 0], [10, 0], [10, 10]]);\nreturn solid;\n"
		res, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
		if err == nil {
			_, err = renderHelper(src, res.Program)
		}
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("want %q, got %v", c.want, err)
		}
	}
}
