// Command gen is the single source of truth for generating pre-parsed
// <ImportPath>/package.pb files across the GSL standard library.
//
// Why incremental SHA-256 skip + --force:
//   - By default (incremental mode), gen reads any existing package.pb file and
//     compares its 32-byte SHA-256 header against SourcesHash(lib.Sources).
//     Matching packages skip both ParsePackage and disk writes, making edits to
//     a single .gsl library ~50x faster while remaining deterministic.
//   - With --force, gen bypasses the SHA-256 check and re-parses and rewrites
//     all standard library package.pb files. This is required when parser/*.go
//     or program/program.proto changes AST output without altering .gsl text.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/internal/stdpkg"
	"gradientcontrol.com/omega/core/parser"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

type parseFunc func(sources []string) (*parser.Result, error)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "stdlib/gen: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("gen", flag.ContinueOnError)
	force := fs.Bool("force", false, "re-parse and rewrite every package.pb regardless of existing SHA-256 header")
	dir := fs.String("dir", "", "path to stdlib directory (auto-detected from working directory when empty)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	stdlibDir, err := resolveStdlibDir(*dir)
	if err != nil {
		return err
	}

	libs := gsllibs.All()
	generated, err := generatePackages(stdlibDir, libs, *force, parser.ParsePackage)
	if err != nil {
		return err
	}
	fmt.Printf("stdlib/gen: updated %d/%d package.pb files\n", generated, len(libs))
	return nil
}

func resolveStdlibDir(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	// Support both `go generate ./stdlib/...` (CWD = stdlib/) and
	// `go run ./stdlib/internal/gen` from the omega-core repository root.
	for _, candidate := range []string{".", "stdlib"} {
		info, err := os.Stat(filepath.Join(candidate, "pkg_info.go"))
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not locate stdlib/pkg_info.go from working directory")
}

func generatePackages(stdlibDir string, libs []gsllibs.PkgInfo, force bool, parse parseFunc) (int, error) {
	// Disable StdLibRegistry during generation so parser.ParsePackage always
	// parses .gsl source text from scratch rather than returning already-embedded
	// package.pb payloads when --force is used.
	savedRegistry := stdpkg.StdLibRegistry
	stdpkg.StdLibRegistry = nil
	defer func() {
		stdpkg.StdLibRegistry = savedRegistry
	}()

	marshalOpts := proto.MarshalOptions{Deterministic: true}
	generated := 0

	for _, lib := range libs {
		wantHash := stdpkg.SourcesHash(lib.Sources)
		pbPath := filepath.Join(stdlibDir, filepath.FromSlash(lib.ImportPath), "package.pb")

		if !force {
			if existing, err := os.ReadFile(pbPath); err == nil &&
				len(existing) >= stdpkg.HashSize &&
				[stdpkg.HashSize]byte(existing[:stdpkg.HashSize]) == wantHash {
				continue
			}
		}

		res, err := parse(lib.Sources)
		if err != nil {
			return generated, fmt.Errorf("parsing stdlib %q: %w", lib.ImportPath, err)
		}

		payload, err := marshalOpts.Marshal(res.Package)
		if err != nil {
			return generated, fmt.Errorf("marshaling stdlib %q: %w", lib.ImportPath, err)
		}

		out := make([]byte, stdpkg.HashSize+len(payload))
		copy(out[:stdpkg.HashSize], wantHash[:])
		copy(out[stdpkg.HashSize:], payload)

		if err := writeFileAtomic(pbPath, out); err != nil {
			return generated, fmt.Errorf("writing %q: %w", pbPath, err)
		}
		generated++
	}
	return generated, nil
}

func writeFileAtomic(targetPath string, data []byte) error {
	dir := filepath.Dir(targetPath)
	tmp, err := os.CreateTemp(dir, ".package.pb.tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, targetPath); err != nil {
		return err
	}
	cleanup = false
	return nil
}
