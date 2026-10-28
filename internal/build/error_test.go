package build

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

func TestErrWithPkgSrc_NamedPrimaryPackage(t *testing.T) {
	primarySrc := "a = 1.0;\nreturn a / 0.0;\n"
	depSrc := "depFunc() { return 2.0; }\n"

	prog := &program.Program{
		Primary: &program.Package{
			Path: "custom/primary_pkg",
			Files: []*program.PackageFile{
				{
					Source: primarySrc,
				},
			},
		},
		Dependencies: []*program.Package{
			{
				Path: "dep_pkg",
				Files: []*program.PackageFile{
					{
						Source: depSrc,
					},
				},
			},
		},
	}

	// 1. Error with matching primary package path
	primaryErr := PkgErrLoc("custom/primary_pkg", &program.Error{
		LineNum: 2,
		Column:  1,
		Message: "division by zero",
	})
	errResult := ErrWithPkgSrc(prog, primaryErr)
	if errResult == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(errResult.Error(), "return a / 0.0") {
		t.Errorf("expected primary source line in error context, got: %v", errResult)
	}

	// 2. Error with dependency package path
	depErr := PkgErrLoc("dep_pkg", &program.Error{
		LineNum: 1,
		Column:  1,
		Message: "bad dep",
	})
	depErrResult := ErrWithPkgSrc(prog, depErr)
	if depErrResult == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(depErrResult.Error(), "depFunc()") {
		t.Errorf("expected dependency source line in error context, got: %v", depErrResult)
	}

	// 3. Error with empty package path (defaults to primary)
	emptyErr := PkgErrLoc("", &program.Error{
		LineNum: 1,
		Column:  1,
		Message: "empty path error",
	})
	emptyErrResult := ErrWithPkgSrc(prog, emptyErr)
	if emptyErrResult == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(emptyErrResult.Error(), "a = 1.0") {
		t.Errorf("expected primary source line in error context, got: %v", emptyErrResult)
	}
}
