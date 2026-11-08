package assembler_test

import (
	"os"
	"path/filepath"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
)

type stubSourceManager struct {
	pkgs map[string][]string
}

func (m *stubSourceManager) Sources(path string) ([]string, []string, error) {
	if srcs, ok := m.pkgs[path]; ok {
		return srcs, nil, nil
	}
	return nil, nil, nil
}

func (m *stubSourceManager) Asset(pkg, filename string) ([]byte, error) {
	return nil, nil
}

func TestAssemble_EmptySources(t *testing.T) {
	_, err := assembler.Assemble(nil)
	if err == nil {
		t.Fatal("expected error for empty sources, got nil")
	}
}

func TestAssemble_SimpleGSL_DefaultStdLibManager(t *testing.T) {
	src := `import "ease";
x = 10;
return sphere(x);`
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if res == nil || res.Program == nil {
		t.Fatal("expected non-nil result and program")
	}
}

func TestAssemble_SourcePathsWithStdLibFallback(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "mypkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "mypkg.gsl"), []byte("rad = 15;\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	src := `import "mypkg";
return sphere(mypkg.rad);`
	res, err := assembler.Assemble([]string{src}, assembler.SourcePaths(root))
	if err != nil {
		t.Fatalf("Assemble with SourcePaths failed: %v", err)
	}
	if res == nil || res.Program == nil {
		t.Fatal("expected non-nil result and program")
	}
}

func TestAssemble_SourceManagerAndSourcePathsWithStdLibFallback(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "diskpkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "diskpkg.gsl"), []byte("offset = 5;\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	customMgr := &stubSourceManager{
		pkgs: map[string][]string{
			"mempkg": {"base_r = 10;\n"},
		},
	}

	src := `import "mempkg";
import "diskpkg";
return sphere(mempkg.base_r + diskpkg.offset);`
	res, err := assembler.Assemble(
		[]string{src},
		assembler.SourceManager(customMgr),
		assembler.SourcePaths(root),
	)
	if err != nil {
		t.Fatalf("Assemble with SourceManager and SourcePaths failed: %v", err)
	}
	if res == nil || res.Program == nil {
		t.Fatal("expected non-nil result and program")
	}
}
