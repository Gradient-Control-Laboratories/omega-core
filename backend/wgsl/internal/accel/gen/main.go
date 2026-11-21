package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/accel"
)

func main() {
	pkgDirFlag := flag.String("pkg-dir", "stdlib", "Path to standard library package directory")
	outFlag := flag.String("out", "backend/wgsl/internal/accel/ops_gen.go", "Path to output Go source file")
	inlinerOutFlag := flag.String("inliner-out", "internal/assembler/inline/wgsl.go", "Path to output inliner registration Go source file")
	flag.Parse()

	pkgDir := resolvePath(*pkgDirFlag)
	outFile := resolvePath(*outFlag)
	inlinerOutFile := resolvePath(*inlinerOutFlag)

	if info, err := os.Stat(pkgDir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: package directory %q not found\n", pkgDir)
		os.Exit(1)
	}

	descriptors, err := accel.ValidateAll(pkgDir, "wgsl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: validation failed: %v\n", err)
		os.Exit(1)
	}

	// 1. Generate WGSL ops table
	code, err := accel.GenerateOps(descriptors)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: ops generation failed: %v\n", err)
		os.Exit(1)
	}
	writeGeneratedFile(outFile, code, "WGSL ops table", len(descriptors))

	// 2. Generate inliner acceleration registration
	inlinerCode, err := accel.GenerateInlinerRegistration(descriptors)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: inliner registration generation failed: %v\n", err)
		os.Exit(1)
	}
	writeGeneratedFile(inlinerOutFile, inlinerCode, "inliner hardware acceleration registry", len(descriptors))
}

func writeGeneratedFile(path string, code []byte, description string, pkgCount int) {
	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, code) {
		fmt.Printf("%s up-to-date: %s\n", description, path)
		return
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(path, code, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated %s (%d packages) -> %s\n", description, pkgCount, path)
}

func resolvePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	if _, err := os.Stat(p); err == nil {
		return p
	}

	candidates := []string{
		filepath.Join("..", "..", "..", "..", p),
		filepath.Join("..", "..", p),
		filepath.Join("..", p),
	}
	for _, cand := range candidates {
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
		if _, err := os.Stat(filepath.Dir(cand)); err == nil {
			return cand
		}
	}

	return p
}
