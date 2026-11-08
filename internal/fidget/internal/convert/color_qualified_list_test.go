package convert_test

import (
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/fidget/internal/convert"
)

// A coloured shape whose distance field uses an imported library function that
// maps over a qualified size table (such as fastener.clearanceHole) must lower
// cleanly through the compiler pipeline and convert to coloured shapes.
func TestColorBlendWithFastenerTable(t *testing.T) {
	src := "import \"mechanical/fastener\";\n" +
		"a = fastener.clearanceHole(3.0, fastener.fitMedium);\n" +
		"a.color = [1.0, 0.0, 0.0];\n" +
		"b = remap(a, p - [12.0, 0.0, 0.0]);\n" +
		"b.color = [0.0, 1.0, 0.0];\n" +
		"return min(a, b);\n"
	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	var got int
	err = convert.ToColorShapes(res.Program, 0, func(shapes []convert.ColorShape) error {
		for _, s := range shapes {
			if s.HasColor {
				got++
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("ToColorShapes: %v", err)
	}
	if got == 0 {
		t.Fatal("no shape carried a colour field")
	}
}

// A coloured shape whose distance field directly maps over a qualified library
// list must assemble and convert cleanly without failing member access.
func TestColorBlendWithDirectMapOverQualifiedList(t *testing.T) {
	src := "import \"mechanical/fastener\";\n" +
		"sel(pr) { return pr.y; }\n" +
		"sum(acc, v) { return acc + v; }\n" +
		"rad() { return reduce(sum, map(sel, catalog.torxTable)); }\n" +
		"a = sphere([0.0, 0.0, 0.0], 5.0) - rad() * 0.001;\n" +
		"a.color = [1.0, 0.0, 0.0];\n" +
		"b = remap(a, p - [12.0, 0.0, 0.0]);\n" +
		"b.color = [0.0, 1.0, 0.0];\n" +
		"return min(a, b);\n"
	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	var got int
	err = convert.ToColorShapes(res.Program, 0, func(shapes []convert.ColorShape) error {
		for _, s := range shapes {
			if s.HasColor {
				got++
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("ToColorShapes: %v", err)
	}
	if got == 0 {
		t.Fatal("no shape carried a colour field")
	}
}
