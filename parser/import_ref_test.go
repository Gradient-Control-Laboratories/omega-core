package parser

import (
	"testing"
)

// TestImportRefExpressionLowers checks that the expression form
// `x = importStep "f"` / `x = importMesh "f"` parses and lowers to a
// FunctionCall named importStep/importMesh carrying the path as a rune list,
// which the slot-assignment pass later resolves to a compiler-assigned id.
func TestImportRefExpressionLowers(t *testing.T) {
	src := "import . \"omega\";\n" +
		"partId = importStep \"part.stp\";\n" +
		"meshId = importMesh \"m.stl\";\n" +
		"return partId, meshId;\n"

	res, err := ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("parse err: %v", err)
	}
	pkg := res.Package

	calls := map[string]string{} // declared name -> import call name
	for _, assign := range pkg.Files[0].Assignments {
		if len(assign.Variables) != 1 {
			continue
		}
		fc := assign.Expression.GetFnCall()
		if fc == nil {
			continue
		}
		if fc.Reference != nil {
			calls[assign.Variables[0].Name] = fc.Reference.Name
		}
	}

	if calls["partId"] != "importStep" {
		t.Fatalf("partId should lower to importStep call, got %q", calls["partId"])
	}
	if calls["meshId"] != "importMesh" {
		t.Fatalf("meshId should lower to importMesh call, got %q", calls["meshId"])
	}

	// The legacy numbered statement form must still parse alongside it.
	legacy := "import . \"omega\";\nimportStep 0 \"upload.stp\";\nreturn mesh(0);\n"
	res2, err := ParsePackage([]string{legacy})
	if err != nil {
		t.Fatalf("legacy numbered form regressed: err=%v", err)
	}
	pkg2 := res2.Package
	if len(pkg2.Files[0].StepImports) != 1 || pkg2.Files[0].StepImports[0].Path != "upload.stp" {
		t.Fatalf("legacy StepImports = %+v", pkg2.Files[0].StepImports)
	}
}

