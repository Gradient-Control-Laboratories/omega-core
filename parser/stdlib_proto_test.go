package parser_test

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/internal/stdpkg"
	"gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// TestStdlibProtosUpToDate ensures every embedded <ImportPath>/package.pb in
// stdlib matches both the 32-byte SHA-256 digest of its .gsl sources and the
// exact AST currently produced by parser.ParsePackage.
func TestStdlibProtosUpToDate(t *testing.T) {
	for _, lib := range gsllibs.All() {
		raw, ok := stdpkg.RawProto(lib.ImportPath)
		if !ok {
			t.Errorf("stdlib %q is missing package.pb; run ./tool/build-stdlib.sh --force", lib.ImportPath)
			continue
		}
		if len(raw) < stdpkg.HashSize {
			t.Errorf("stdlib %q package.pb is truncated (%d bytes); run ./tool/build-stdlib.sh --force", lib.ImportPath, len(raw))
			continue
		}

		wantHash := stdpkg.SourcesHash(lib.Sources)
		gotHash := [stdpkg.HashSize]byte(raw[:stdpkg.HashSize])
		if gotHash != wantHash {
			t.Errorf("stdlib %q package.pb source hash mismatch; run ./tool/build-stdlib.sh --force", lib.ImportPath)
			continue
		}

		var embeddedPkg program.Package
		if err := proto.Unmarshal(raw[stdpkg.HashSize:], &embeddedPkg); err != nil {
			t.Errorf("stdlib %q package.pb unmarshal error (%v); run ./tool/build-stdlib.sh --force", lib.ImportPath, err)
			continue
		}

		fresh, err := parser.ParsePackage(lib.Sources)
		if err != nil {
			t.Fatalf("ParsePackage(%q) failed: %v", lib.ImportPath, err)
		}
		if !proto.Equal(&embeddedPkg, fresh.Package) {
			t.Errorf("stdlib %q package.pb AST differs from ParsePackage output; run ./tool/build-stdlib.sh --force", lib.ImportPath)
		}
	}
}
