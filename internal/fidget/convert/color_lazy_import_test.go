package convert_test

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/fidget/convert"
)

// A coloured shape built on a library body that qualifies a package the
// document never imports — cylinderZ reaches for vector.dirZ — must still
// lower its colour field when it is unioned with another coloured shape. The
// blend inlines cylinderZ's body into the synthesized field, so the builder
// has to bind "vector" lazily; without that the build fails with
// `package "vector" not imported`, and every offline render of such a
// document is lost.
func TestColorBlendBindsInlinedLibraryPrefixes(t *testing.T) {
	src := "a = max(cylinderZ([0.0, 0.0, 0.0], 5.0), abs(p.z) - 2.0);\n" +
		"a.color = [1.0, 0.0, 0.0];\n" +
		"b = remap(a, p - [12.0, 0.0, 0.0]);\n" +
		"b.color = [0.0, 1.0, 0.0];\n" +
		"return min(a, b);\n"
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
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
