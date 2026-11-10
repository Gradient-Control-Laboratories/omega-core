package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/internal/stdpkg"
	"gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

func TestGenerate_IncrementalSkipForceAndDeterminism(t *testing.T) {
	tmpDir := t.TempDir()
	libs := []gsllibs.PkgInfo{
		{ImportPath: "color", Sources: []string{"red = [1.0, 0.0, 0.0];\n"}},
		{ImportPath: "lattice/beam", Sources: []string{"strut = 1.0;\n"}},
	}
	for _, lib := range libs {
		if err := os.MkdirAll(filepath.Join(tmpDir, filepath.FromSlash(lib.ImportPath)), 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
	}

	parseCalls := 0
	countingParse := func(sources []string) (*parser.Result, error) {
		parseCalls++
		return parser.ParsePackage(sources)
	}

	// 1. Initial generation parses and writes both package.pb files.
	generated, err := generatePackages(tmpDir, libs, false, countingParse)
	if err != nil {
		t.Fatalf("initial generatePackages failed: %v", err)
	}
	if generated != 2 || parseCalls != 2 {
		t.Fatalf("initial generate: generated=%d parseCalls=%d, want 2, 2", generated, parseCalls)
	}

	colorPBPath := filepath.Join(tmpDir, "color", "package.pb")
	firstBytes, err := os.ReadFile(colorPBPath)
	if err != nil {
		t.Fatalf("reading color/package.pb: %v", err)
	}
	wantHash := stdpkg.SourcesHash(libs[0].Sources)
	if len(firstBytes) < stdpkg.HashSize || [stdpkg.HashSize]byte(firstBytes[:stdpkg.HashSize]) != wantHash {
		t.Fatalf("color/package.pb missing expected 32-byte SHA-256 header")
	}
	var unmarshaled program.Package
	if err := proto.Unmarshal(firstBytes[stdpkg.HashSize:], &unmarshaled); err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}
	fresh, err := parser.ParsePackage(libs[0].Sources)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}
	if !proto.Equal(&unmarshaled, fresh.Package) {
		t.Fatal("unmarshaled package.pb does not proto.Equal freshly parsed Package")
	}

	// 2. Second run without --force skips both ParsePackage and writing for unchanged packages.
	parseCalls = 0
	generated, err = generatePackages(tmpDir, libs, false, countingParse)
	if err != nil {
		t.Fatalf("incremental generatePackages failed: %v", err)
	}
	if generated != 0 || parseCalls != 0 {
		t.Fatalf("incremental generate: generated=%d parseCalls=%d, want 0, 0", generated, parseCalls)
	}

	// 3. Modifying only one package re-parses only that single modified package.
	libsModified := []gsllibs.PkgInfo{
		{ImportPath: "color", Sources: []string{"red = [1.0, 0.0, 0.0];\nblue = [0.0, 0.0, 1.0];\n"}},
		libs[1],
	}
	parseCalls = 0
	generated, err = generatePackages(tmpDir, libsModified, false, countingParse)
	if err != nil {
		t.Fatalf("single-edit generatePackages failed: %v", err)
	}
	if generated != 1 || parseCalls != 1 {
		t.Fatalf("single-edit generate: generated=%d parseCalls=%d, want 1, 1", generated, parseCalls)
	}

	// 4. --force re-parses all packages and produces byte-for-byte deterministic output.
	beforeForce, err := os.ReadFile(colorPBPath)
	if err != nil {
		t.Fatalf("ReadFile before force: %v", err)
	}
	parseCalls = 0
	generated, err = generatePackages(tmpDir, libsModified, true, countingParse)
	if err != nil {
		t.Fatalf("forced generatePackages failed: %v", err)
	}
	if generated != 2 || parseCalls != 2 {
		t.Fatalf("forced generate: generated=%d parseCalls=%d, want 2, 2", generated, parseCalls)
	}
	afterForce, err := os.ReadFile(colorPBPath)
	if err != nil {
		t.Fatalf("ReadFile after force: %v", err)
	}
	if !bytes.Equal(beforeForce, afterForce) {
		t.Fatal("--force produced non-deterministic bytes compared to prior run")
	}
}

func TestResolveStdlibDirAndErrorPropagation(t *testing.T) {
	if got, err := resolveStdlibDir("/explicit/path"); err != nil || got != "/explicit/path" {
		t.Fatalf("resolveStdlibDir explicit = (%q, %v), want (\"/explicit/path\", nil)", got, err)
	}

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWD) })

	emptyDir := t.TempDir()
	if err := os.Chdir(emptyDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	if _, err := resolveStdlibDir(""); err == nil {
		t.Fatal("expected error when pkg_info.go cannot be found")
	}

	if err := os.MkdirAll(filepath.Join(emptyDir, "stdlib"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(emptyDir, "stdlib", "pkg_info.go"), []byte("package gsllibs\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if got, err := resolveStdlibDir(""); err != nil || got != "stdlib" {
		t.Fatalf("resolveStdlibDir auto-detect = (%q, %v), want (\"stdlib\", nil)", got, err)
	}

	// Verify parse error surfaces cleanly without writing package.pb.
	libs := []gsllibs.PkgInfo{{ImportPath: "bad", Sources: []string{"invalid"}}}
	_, err = generatePackages(emptyDir, libs, true, func([]string) (*parser.Result, error) {
		return nil, errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected parse error from generatePackages, got nil")
	}

	// Verify run() flag error and real stdlib incremental no-op via --dir.
	if err := run([]string{"--unknown-flag"}); err == nil {
		t.Fatal("expected flag parse error from run(), got nil")
	}
	if err := run([]string{"--dir", filepath.Join(origWD, "..", "..")}); err != nil {
		t.Fatalf("run(--dir stdlib) failed: %v", err)
	}
}
