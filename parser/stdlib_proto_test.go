package parser

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/parser/internal/parser2"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// TestStdlibProtosUpToDate ensures every embedded <ImportPath>/package.pb in
// stdlib has a valid 32-byte SHA-256 source-hash header (verified by
// program.GetStdPackage via StdLibRegistry.Proto) and unmarshals to the exact
// AST currently produced by parser2.Parse.
func TestStdlibProtosUpToDate(t *testing.T) {
	for _, lib := range gsllibs.All() {
		embeddedPkg := program.GetStdPackage(lib.ImportPath)
		if embeddedPkg == nil {
			t.Errorf("stdlib %q is missing or has a stale package.pb header; run ./tool/build-stdlib.sh --force", lib.ImportPath)
			continue
		}

		freshPkg, err := parser2.Parse(lib.Sources)
		if err != nil {
			t.Fatalf("parser2.Parse(%q) failed: %v", lib.ImportPath, err)
		}
		if !proto.Equal(embeddedPkg, freshPkg) {
			t.Errorf("stdlib %q package.pb AST differs from fresh parser2.Parse output; run ./tool/build-stdlib.sh --force", lib.ImportPath)
		}
	}
}
