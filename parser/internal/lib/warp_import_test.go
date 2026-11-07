package lib

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// importWarp has two spellings, and which one a document reaches for follows
// whatever its other imports use. Both have to work, and the mistakes have to
// be told apart: parentheses are not the syntax for any import, and a warp is
// not a mesh.

// warpImportsOf gathers the package's warp imports across its files: they hang
// off PackageFile now, where the flat Package accessor used to serve.
func warpImportsOf(pkg *program.Package) []*program.WarpImport {
	var out []*program.WarpImport
	for _, file := range pkg.GetFiles() {
		out = append(out, file.GetWarpImports()...)
	}
	return out
}

func TestWarpImportNumberedForm(t *testing.T) {
	pkg, err := ParsePackage([]string{`importStep 0 "p.hodge";
importWarp 0 "field.warp";
return meshWarped(0, 0, 1);
`})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	imports := warpImportsOf(pkg)
	if len(imports) != 1 {
		t.Fatalf("warp imports = %d, want 1", len(imports))
	}
	if imports[0].GetIndex() != 0 || imports[0].GetPath() != "field.warp" {
		t.Errorf("warp import = %d %q", imports[0].GetIndex(), imports[0].GetPath())
	}
}

// The expression form carries no slot — it lowers to a call the canvas agent
// resolves — so the parser must accept it while leaving WarpImports empty.
func TestWarpImportExpressionForm(t *testing.T) {
	pkg, err := ParsePackage([]string{`importId = importStep "p.hodge";
warpID = importWarp "field.warp";
return meshWarped(importId, warpID, -1);
`})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if n := len(warpImportsOf(pkg)); n != 0 {
		t.Errorf("the expression form should not populate WarpImports, got %d", n)
	}
}

func TestWarpImportRejectsNonSimulationFile(t *testing.T) {
	_, err := ParsePackage([]string{`importWarp 0 "field.stl";
return 0;
`})
	if err == nil {
		t.Fatal("importWarp must reject a path that is not a simulation")
	}
	if !strings.Contains(err.Error(), ".warp") {
		t.Errorf("the error should name the expected extension, got: %s", err.Error())
	}
}

// The two builtins take (meshSlot, warpSlot, scale) ahead of the placement
// arguments they share with mesh() and vdbSdf(), so their arities are the
// unwarped ones shifted by two.
func TestWarpedBuiltinArities(t *testing.T) {
	ok := []string{
		`return meshWarped(0, 0, 1);`,
		`return meshWarped(0, 0, 1, [1,2,3]);`,
		`return meshWarped(0, 0, 1, [1,2,3], [[1,0,0],[0,1,0],[0,0,1]]);`,
		`return vdbSdfWarped(0, 0, 1);`,
		`return vdbSdfWarped(0, 0, 1, 256);`,
	}
	prefix := "importMesh 0 \"p.stl\";\nimportWarp 0 \"c.warp\";\n"
	for _, src := range ok {
		_, err := ParsePackage([]string{prefix + src + "\n"})
		if err != nil {
			t.Errorf("%s -> parse: %v", src, err)
		}
	}
	// Too few arguments: the scale is required, because +1 and -1 are
	// different manufacturing objects with identical geometry counts.
	_, err := ParsePackage([]string{prefix + "return meshWarped(0, 0);\n"})
	if err == nil {
		t.Error("meshWarped without a scale must be rejected")
	}
}
