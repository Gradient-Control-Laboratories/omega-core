package codegen_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/source"
)

const textFontPath = "../../../../../../stdlib/text/text.gsl"

// Verifies the generated pen-stroke font compiles end-to-end: the helpers,
// per-glyph functions, branchless glyph(code) dispatch, and the map/reduce
// left() renderer all parse, type-check, and generate WGSL. A string literal
// feeds character codes straight into left().
func TestFontCompiles(t *testing.T) {
	font, err := os.ReadFile(textFontPath)
	if err != nil {
		t.Skipf("font not generated at %s: %v", textFontPath, err)
	}

	// Full signature: codes, position, size, thickness, spacing.
	src := string(font) + "\nreturn left(\"Hi\", [0, 0], 1, 0.05, 1.2);\n"

	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatalf("assemble error: %v", err)
	}

	rendered, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if rendered == nil || rendered.WGSL == "" {
		t.Fatal("no WGSL produced")
	}
	fmt.Printf("text font renders to %d bytes of WGSL\n", len(rendered.WGSL))
}

func TestFontCompilesDynamic(t *testing.T) {
	font, err := os.ReadFile(textFontPath)
	if err != nil {
		t.Skipf("font not generated at %s: %v", textFontPath, err)
	}

	src := string(font) + "\nmyText(codes, pos) { return left(codes, pos, 1); }\nreturn myText([104.0, 101.0], [0.0, 0.0]);\n"

	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatalf("assemble error: %v", err)
	}

	rendered, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if rendered == nil || rendered.WGSL == "" {
		t.Fatal("no WGSL produced")
	}
}

func TestFontCompilesDynamicTime(t *testing.T) {
	font, err := os.ReadFile(textFontPath)
	if err != nil {
		t.Skipf("font not generated at %s: %v", textFontPath, err)
	}

	src := string(font) + "\nreturn left([104.0, 101.0 + time], [0.0, 0.0], 1);\n"

	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatalf("assemble error: %v", err)
	}

	rendered, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if rendered == nil || rendered.WGSL == "" {
		t.Fatal("no WGSL produced")
	}
}

func TestFontImportsDynamicTime(t *testing.T) {
	const pkgRoot = "../../../../../../stdlib"
	gsl := pkgRoot + "/text/text.gsl"
	if _, err := os.Stat(gsl); err != nil {
		t.Skipf("font not generated: %v", err)
	}

	src := "import . \"text\";\n" +
		"return left([104.0, 101.0 + time], [0.0, 0.0], 1);\n"

	mgr := source.NewLocalManager(pkgRoot)
	res, err := assembler.Assemble([]string{src}, mgr)
	if err != nil {
		t.Fatalf("assemble error: %v", err)
	}

	rendered, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if rendered == nil || rendered.WGSL == "" {
		t.Fatal("no WGSL produced")
	}
}

// Verifies the font resolves as an importable package and that string literals
// of various scripts (the codes are Unicode codepoints) resolve to glyphs.
func TestFontImports(t *testing.T) {
	const pkgRoot = "../../../../../../stdlib"

	cases := []struct{ name, pkg, text string }{
		{"latin", "text", "Hi"},
		{"greek", "text", "ΓΔΩ"},   // U+0393 U+0394 U+03A9
		{"cyrillic", "text", "Дф"}, // U+0414 U+0444
		{"accented", "text", "café½"},
		{"symbols", "text", "√≤∞→"}, // math + arrow
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gsl := pkgRoot + "/" + c.pkg + "/" + path.Base(c.pkg) + ".gsl"
			if _, err := os.Stat(gsl); err != nil {
				t.Skipf("font not generated: %v", err)
			}
			// 3-arg overload (default thickness + spacing).
			src := "import . \"" + c.pkg + "\";\n" +
				"return max(left(\"" + c.text + "\", [0, 0], 1), abs(p.z) - 1);\n"

			mgr := source.NewLocalManager(pkgRoot)
			res, err := assembler.Assemble([]string{src}, mgr)
			if err != nil {
				t.Fatalf("assemble error: %v", err)
			}

			rendered, err := renderHelper(src, res.Program)
			if err != nil {
				t.Fatalf("render error: %v", err)
			}
			if rendered == nil || rendered.WGSL == "" {
				t.Fatal("no WGSL produced")
			}
		})
	}
}
